package main

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"exam-system-mvp/config"
	"exam-system-mvp/models"
	"exam-system-mvp/router"
	"exam-system-mvp/utils"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ─── Custom SQLite Driver for Tests ──────────────────────────────────────────
// SQLite returns expression results (such as COALESCE(e.published_at, ...))
// with an empty decltype, causing go-sqlite3 to yield raw strings instead of
// time.Time. This driver wrapper detects timestamp/datetime columns by name
// and parses them so that sql.Scan into *time.Time succeeds seamlessly in tests.

func init() {
	sql.Register("sqlite3_test", &testDriver{parent: &sqlite3.SQLiteDriver{}})
}

type testDriver struct {
	parent driver.Driver
}

func (d *testDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.parent.Open(name)
	if err != nil {
		return nil, err
	}
	return &testConn{Conn: conn}, nil
}

type testConn struct {
	driver.Conn
}

func (c *testConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if pc, ok := c.Conn.(driver.ConnPrepareContext); ok {
		stmt, err := pc.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}
		return &testStmt{Stmt: stmt}, nil
	}
	stmt, err := c.Conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &testStmt{Stmt: stmt}, nil
}

func (c *testConn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *testConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if e, ok := c.Conn.(driver.Execer); ok {
		return e.Exec(query, args)
	}
	return nil, driver.ErrSkip
}

func (c *testConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if ec, ok := c.Conn.(driver.ExecerContext); ok {
		return ec.ExecContext(ctx, query, args)
	}
	dargs := make([]driver.Value, len(args))
	for i, a := range args {
		dargs[i] = a.Value
	}
	return c.Exec(query, dargs)
}

func (c *testConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if q, ok := c.Conn.(driver.Queryer); ok {
		rows, err := q.Query(query, args)
		if err != nil {
			return nil, err
		}
		return &testRows{Rows: rows, cols: rows.Columns()}, nil
	}
	return nil, driver.ErrSkip
}

func (c *testConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if qc, ok := c.Conn.(driver.QueryerContext); ok {
		rows, err := qc.QueryContext(ctx, query, args)
		if err != nil {
			return nil, err
		}
		return &testRows{Rows: rows, cols: rows.Columns()}, nil
	}
	dargs := make([]driver.Value, len(args))
	for i, a := range args {
		dargs[i] = a.Value
	}
	return c.Query(query, dargs)
}

func (c *testConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if bt, ok := c.Conn.(driver.ConnBeginTx); ok {
		return bt.BeginTx(ctx, opts)
	}
	return c.Conn.Begin()
}

func (c *testConn) Ping(ctx context.Context) error {
	if p, ok := c.Conn.(driver.Pinger); ok {
		return p.Ping(ctx)
	}
	return nil
}

type testStmt struct {
	driver.Stmt
}

func (s *testStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if ec, ok := s.Stmt.(driver.StmtExecContext); ok {
		return ec.ExecContext(ctx, args)
	}
	dargs := make([]driver.Value, len(args))
	for i, a := range args {
		dargs[i] = a.Value
	}
	return s.Stmt.Exec(dargs)
}

func (s *testStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if qc, ok := s.Stmt.(driver.StmtQueryContext); ok {
		rows, err := qc.QueryContext(ctx, args)
		if err != nil {
			return nil, err
		}
		return &testRows{Rows: rows, cols: rows.Columns()}, nil
	}
	dargs := make([]driver.Value, len(args))
	for i, a := range args {
		dargs[i] = a.Value
	}
	rows, err := s.Stmt.Query(dargs)
	if err != nil {
		return nil, err
	}
	return &testRows{Rows: rows, cols: rows.Columns()}, nil
}

type testRows struct {
	driver.Rows
	cols []string
}

func (r *testRows) Next(dest []driver.Value) error {
	err := r.Rows.Next(dest)
	if err != nil {
		return err
	}
	for i, col := range r.cols {
		if strings.HasSuffix(col, "_at") || strings.HasSuffix(col, "_time") {
			if s, ok := dest[i].(string); ok {
				for _, format := range []string{
					time.RFC3339Nano,
					time.RFC3339,
					"2006-01-02 15:04:05.999999999-07:00",
					"2006-01-02 15:04:05.999999999",
					"2006-01-02 15:04:05-07:00",
					"2006-01-02 15:04:05",
					"2006-01-02T15:04:05",
				} {
					if t, parseErr := time.Parse(format, s); parseErr == nil {
						dest[i] = t
						break
					}
				}
			}
		}
	}
	return nil
}

// ─── Test helpers ─────────────────────────────────────────────────────────────

func setupTestServer(t *testing.T) *gin.Engine {
	t.Helper()

	// Use a unique SQLite file per test run (in-process, auto-cleaned)
	dbPath := fmt.Sprintf("file:test_%d?mode=memory&cache=shared&_foreign_keys=on", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Dialector{DriverName: "sqlite3_test", DSN: dbPath}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}

	config.DB = db
	os.Setenv("JWT_SECRET", "test_secret_key_2026")

	// Migrate schema
	if err := db.AutoMigrate(
		&models.User{},
		&models.Exam{},
		&models.Question{},
		&models.ExamAttempt{},
		&models.AttemptAnswer{},
		&models.Appeal{},
	); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// SQLite unique index for (exam_id, user_id) and (attempt_id, question_id)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_ea_exam_user ON exam_attempts(exam_id, user_id)`)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_aa_attempt_question ON attempt_answers(attempt_id, question_id)`)

	models.SeedData(db)

	gin.SetMode(gin.TestMode)
	return router.SetupRouter()
}

// do sends an HTTP request to the test server and returns the response.
func do(t *testing.T, srv *gin.Engine, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w
}

// parseJSON parses the response body into a map.
func parseJSON(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("parseJSON failed (status=%d body=%s): %v", w.Code, w.Body.String(), err)
	}
	return m
}

// mustString extracts a nested string by dot-path from map.
func getString(m map[string]interface{}, keys ...string) string {
	var cur interface{} = m
	for _, k := range keys {
		sub, ok := cur.(map[string]interface{})
		if !ok {
			return ""
		}
		cur = sub[k]
	}
	s, _ := cur.(string)
	return s
}

func getFloat(m map[string]interface{}, keys ...string) float64 {
	var cur interface{} = m
	for _, k := range keys {
		sub, ok := cur.(map[string]interface{})
		if !ok {
			return 0
		}
		cur = sub[k]
	}
	f, _ := cur.(float64)
	return f
}

// ─── Full integration test ────────────────────────────────────────────────────

// TestMVP_FullFlow runs all 6 use-cases sequentially in a single test to share
// state (tokens, IDs) across the flow.
func TestMVP_FullFlow(t *testing.T) {
	srv := setupTestServer(t)

	var teacherToken, studentToken, examCode string
	var examID, q1ID, q2ID, attemptID float64

	// ══════════════════════════════════════════════════════════════════════════
	// UC01 – Register & Login
	// ══════════════════════════════════════════════════════════════════════════
	t.Run("UC01_Register_Student", func(t *testing.T) {
		w := do(t, srv, "POST", "/api/auth/register", "", map[string]interface{}{
			"full_name": "Học Sinh Test",
			"email":     "student_test@exam.local",
			"password":  "123456",
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("register student: want 201, got %d — %s", w.Code, w.Body)
		}
		body := parseJSON(t, w)
		if getString(body, "user", "role") != "student" {
			t.Errorf("expected role=student, got %q", getString(body, "user", "role"))
		}
	})

	t.Run("UC01_Register_Duplicate_Email_Should_Fail", func(t *testing.T) {
		w := do(t, srv, "POST", "/api/auth/register", "", map[string]interface{}{
			"full_name": "Học Sinh Test",
			"email":     "student_test@exam.local",
			"password":  "123456",
		})
		if w.Code != http.StatusConflict {
			t.Fatalf("want 409 conflict, got %d", w.Code)
		}
	})

	t.Run("UC01_Verify_Seeded_Accounts", func(t *testing.T) {
		users := []struct {
			email string
			role  string
		}{
			{email: "teacher@test.com", role: models.RoleTeacher},
			{email: "student@test.com", role: models.RoleStudent},
		}
		for _, expected := range users {
			var user models.User
			if err := config.DB.Where("email = ?", expected.email).First(&user).Error; err != nil {
				t.Fatalf("seeded user %q not found: %v", expected.email, err)
			}
			if user.Role != expected.role {
				t.Errorf("user %q: expected role=%s, got %s", expected.email, expected.role, user.Role)
			}
		}
	})

	t.Run("UC01_Login_Teacher", func(t *testing.T) {
		w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
			"email":    "teacher@test.com",
			"password": "123456",
		})
		if w.Code != http.StatusOK {
			t.Fatalf("login teacher: want 200, got %d — %s", w.Code, w.Body)
		}
		body := parseJSON(t, w)
		teacherToken = getString(body, "token")
		if teacherToken == "" {
			t.Fatal("teacher token is empty")
		}
		if getString(body, "user", "role") != "teacher" {
			t.Errorf("expected role=teacher, got %s", getString(body, "user", "role"))
		}
	})

	t.Run("UC01_Login_Student", func(t *testing.T) {
		w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
			"email":    "student_test@exam.local",
			"password": "123456",
		})
		if w.Code != http.StatusOK {
			t.Fatalf("login student: want 200, got %d — %s", w.Code, w.Body)
		}
		body := parseJSON(t, w)
		studentToken = getString(body, "token")
		if studentToken == "" {
			t.Fatal("student token is empty")
		}
	})

	t.Run("UC01_Login_Wrong_Password_Should_Fail", func(t *testing.T) {
		w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
			"email":    "teacher@test.com",
			"password": "wrongpass",
		})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	// ══════════════════════════════════════════════════════════════════════════
	// UC02 – Teacher creates exam, adds questions, publishes
	// ══════════════════════════════════════════════════════════════════════════
	t.Run("UC02_Create_Exam", func(t *testing.T) {
		now := time.Now()
		w := do(t, srv, "POST", "/api/teacher/exams", teacherToken, map[string]interface{}{
			"title":      "Kiểm tra Go Integration Test",
			"duration":   30,
			"start_time": now.Add(-1 * time.Hour).Format(time.RFC3339),
			"end_time":   now.Add(7 * 24 * time.Hour).Format(time.RFC3339),
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("create exam: want 201, got %d — %s", w.Code, w.Body)
		}
		body := parseJSON(t, w)
		exam, _ := body["exam"].(map[string]interface{})
		examID = exam["id"].(float64)
		if examID == 0 {
			t.Fatal("exam ID is 0")
		}
		if exam["status"] != "draft" {
			t.Errorf("expected status=draft, got %v", exam["status"])
		}
		examCode, _ = exam["exam_code"].(string)
		if examCode == "" {
			t.Error("exam_code should not be empty")
		}
	})

	t.Run("UC02_Add_Questions", func(t *testing.T) {
		url := fmt.Sprintf("/api/teacher/exams/%.0f/questions", examID)
		w := do(t, srv, "POST", url, teacherToken, map[string]interface{}{
			"questions": []map[string]interface{}{
				{
					"prompt":         "HTTP là viết tắt của gì?",
					"option_a":       "HyperText Transfer Protocol",
					"option_b":       "High Transfer Text Protocol",
					"option_c":       "HyperText Transmission Protocol",
					"option_d":       "High Text Transfer Protocol",
					"correct_option": "A",
					"points":         1.0,
				},
				{
					"prompt":         "Ngôn ngữ lập trình nào được dùng viết backend này?",
					"option_a":       "Python",
					"option_b":       "JavaScript",
					"option_c":       "Go",
					"option_d":       "Java",
					"correct_option": "C",
					"points":         2.0,
				},
			},
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("add questions: want 201, got %d — %s", w.Code, w.Body)
		}
		body := parseJSON(t, w)
		qs, _ := body["questions"].([]interface{})
		if len(qs) != 2 {
			t.Fatalf("expected 2 questions, got %d", len(qs))
		}
		q1, _ := qs[0].(map[string]interface{})
		q2, _ := qs[1].(map[string]interface{})
		q1ID = q1["id"].(float64)
		q2ID = q2["id"].(float64)
	})

	t.Run("UC02_Publish_Exam", func(t *testing.T) {
		url := fmt.Sprintf("/api/teacher/exams/%.0f/publish", examID)
		w := do(t, srv, "PATCH", url, teacherToken, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("publish: want 200, got %d — %s", w.Code, w.Body)
		}
	})

	t.Run("UC02_Cannot_Add_Questions_After_Publish", func(t *testing.T) {
		url := fmt.Sprintf("/api/teacher/exams/%.0f/questions", examID)
		w := do(t, srv, "POST", url, teacherToken, map[string]interface{}{
			"questions": []map[string]interface{}{
				{
					"prompt":   "Extra question",
					"option_a": "A", "option_b": "B", "option_c": "C", "option_d": "D",
					"correct_option": "A",
				},
			},
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("should reject adding question to published exam: want 400, got %d — %s", w.Code, w.Body)
		}
	})

	t.Run("UC02_Student_Cannot_Access_Teacher_API", func(t *testing.T) {
		w := do(t, srv, "GET", "/api/teacher/exams", studentToken, nil)
		if w.Code != http.StatusForbidden {
			t.Fatalf("want 403 for student accessing teacher API, got %d", w.Code)
		}
	})

	t.Run("UC02_Teacher_Cannot_Access_Student_API", func(t *testing.T) {
		w := do(t, srv, "GET", "/api/student/attempts", teacherToken, nil)
		if w.Code != http.StatusForbidden {
			t.Fatalf("want 403 for teacher accessing student API, got %d", w.Code)
		}
	})

	t.Run("UC02_Unauthenticated_Cannot_Access", func(t *testing.T) {
		w := do(t, srv, "GET", "/api/teacher/exams", "", nil)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401 for unauthenticated, got %d", w.Code)
		}
	})

	// ══════════════════════════════════════════════════════════════════════════
	// UC03 – Student looks up exam by code
	// ══════════════════════════════════════════════════════════════════════════
	t.Run("UC03_Lookup_Exam_By_Code", func(t *testing.T) {
		// Student looks up created exam by code
		w2 := do(t, srv, "GET", "/api/student/exams/lookup?code="+examCode, studentToken, nil)
		if w2.Code != http.StatusOK {
			t.Fatalf("lookup: want 200, got %d — %s", w2.Code, w2.Body)
		}
		body2 := parseJSON(t, w2)
		if body2["exam_status"] != "open" {
			t.Errorf("expected exam_status=open, got %v", body2["exam_status"])
		}
		if body2["attempt_status"] != "not_attempted" {
			t.Errorf("expected attempt_status=not_attempted, got %v", body2["attempt_status"])
		}
	})

	t.Run("UC03_Lookup_Invalid_Code", func(t *testing.T) {
		w := do(t, srv, "GET", "/api/student/exams/lookup?code=INVALID99", studentToken, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	// ══════════════════════════════════════════════════════════════════════════
	// UC04 – Student starts exam, saves answer, resumes
	// ══════════════════════════════════════════════════════════════════════════
	t.Run("UC04_Start_Exam", func(t *testing.T) {
		url := fmt.Sprintf("/api/student/exams/%.0f/start", examID)
		w := do(t, srv, "POST", url, studentToken, map[string]bool{"confirmed": true})
		if w.Code != http.StatusOK {
			t.Fatalf("start exam: want 200, got %d — %s", w.Code, w.Body)
		}
		body := parseJSON(t, w)

		// Store attempt ID
		attemptMap, _ := body["attempt"].(map[string]interface{})
		attemptID = attemptMap["id"].(float64)
		if attemptID == 0 {
			t.Fatal("attempt ID is 0")
		}

		// Verify remaining_seconds is reasonable
		remaining := body["remaining_seconds"].(float64)
		if remaining <= 0 || remaining > 1800 { // 30 min = 1800s
			t.Errorf("remaining_seconds out of range: %v", remaining)
		}

		// CRITICAL: questions must NOT contain correct_option
		questions, _ := body["questions"].([]interface{})
		if len(questions) != 2 {
			t.Fatalf("expected 2 questions, got %d", len(questions))
		}
		for i, qRaw := range questions {
			q, _ := qRaw.(map[string]interface{})
			if _, hasCorrect := q["correct_option"]; hasCorrect {
				t.Errorf("question[%d] LEAKS correct_option — security violation!", i)
			}
		}
	})

	t.Run("UC04_Save_Answer_Q1", func(t *testing.T) {
		url := fmt.Sprintf("/api/student/attempts/%.0f/save", attemptID)
		w := do(t, srv, "POST", url, studentToken, map[string]interface{}{
			"question_id":   int(q1ID),
			"chosen_option": "A", // correct for Q1
		})
		if w.Code != http.StatusOK {
			t.Fatalf("save answer: want 200, got %d — %s", w.Code, w.Body)
		}
	})

	t.Run("UC04_Resume_Exam_Restores_Answers", func(t *testing.T) {
		url := fmt.Sprintf("/api/student/exams/%.0f/start", examID)
		w := do(t, srv, "POST", url, studentToken, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("resume: want 200, got %d — %s", w.Code, w.Body)
		}
		body := parseJSON(t, w)

		// remaining_seconds should still be > 0
		remaining := body["remaining_seconds"].(float64)
		if remaining <= 0 {
			t.Error("remaining_seconds should be > 0 on resume")
		}

		// saved_answers should contain q1 answer
		savedAnswers, _ := body["saved_answers"].([]interface{})
		if len(savedAnswers) == 0 {
			t.Fatal("saved_answers should not be empty after saving Q1")
		}
		found := false
		for _, aRaw := range savedAnswers {
			a, _ := aRaw.(map[string]interface{})
			qid := a["question_id"].(float64)
			opt, _ := a["chosen_option"].(string)
			if qid == q1ID && opt == "A" {
				found = true
			}
		}
		if !found {
			t.Errorf("Q1 answer 'A' not found in saved_answers: %v", savedAnswers)
		}
	})

	// ══════════════════════════════════════════════════════════════════════════
	// UC05 – Submit and auto-grade
	// ══════════════════════════════════════════════════════════════════════════
	t.Run("UC05_Submit_With_Answers", func(t *testing.T) {
		url := fmt.Sprintf("/api/student/attempts/%.0f/submit", attemptID)
		// Q1: A (correct, 1pt), Q2: A (wrong, correct=C, 0pt)
		w := do(t, srv, "POST", url, studentToken, map[string]interface{}{
			"answers": []map[string]interface{}{
				{"question_id": int(q1ID), "chosen_option": "A"},
				{"question_id": int(q2ID), "chosen_option": "A"}, // intentionally wrong
			},
		})
		if w.Code != http.StatusOK {
			t.Fatalf("submit: want 200, got %d — %s", w.Code, w.Body)
		}
		body := parseJSON(t, w)
		attempt, _ := body["attempt"].(map[string]interface{})

		if attempt["status"] != "completed" {
			t.Errorf("expected status=completed, got %v", attempt["status"])
		}
		score := attempt["score"].(float64)
		if score != 1.0 {
			t.Errorf("expected score=1.0 (only Q1 correct), got %.2f", score)
		}
		correctCount := attempt["correct_count"].(float64)
		if correctCount != 1 {
			t.Errorf("expected correct_count=1, got %.0f", correctCount)
		}
		wrongCount := attempt["wrong_count"].(float64)
		if wrongCount != 1 {
			t.Errorf("expected wrong_count=1, got %.0f", wrongCount)
		}
	})

	t.Run("UC05_Cannot_Submit_Twice", func(t *testing.T) {
		url := fmt.Sprintf("/api/student/attempts/%.0f/submit", attemptID)
		w := do(t, srv, "POST", url, studentToken, map[string]interface{}{})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("double submit: want 400, got %d", w.Code)
		}
	})

	// ══════════════════════════════════════════════════════════════════════════
	// UC06 – View results (student & teacher)
	// ══════════════════════════════════════════════════════════════════════════
	t.Run("UC06_Student_View_Result", func(t *testing.T) {
		url := fmt.Sprintf("/api/student/attempts/%.0f/result", attemptID)
		w := do(t, srv, "GET", url, studentToken, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("result: want 200, got %d — %s", w.Code, w.Body)
		}
		body := parseJSON(t, w)

		attempt, _ := body["attempt"].(map[string]interface{})
		if attempt["status"] != "completed" {
			t.Errorf("expected completed, got %v", attempt["status"])
		}

		answers, _ := body["answers"].([]interface{})
		if len(answers) != 2 {
			t.Fatalf("expected 2 answers, got %d", len(answers))
		}

		// Exam is still open (end_time = 7 days from now) → correct_option hidden
		showCorrect, _ := body["show_correct_answers"].(bool)
		if showCorrect {
			t.Error("should NOT show correct answers while exam is still open")
		}
		for _, aRaw := range answers {
			a, _ := aRaw.(map[string]interface{})
			if _, has := a["correct_option"]; has && a["correct_option"] != nil {
				t.Errorf("correct_option should be hidden while exam open, but got: %v", a["correct_option"])
			}
		}
	})

	t.Run("UC06_Teacher_View_Results_Statistics", func(t *testing.T) {
		url := fmt.Sprintf("/api/teacher/exams/%.0f/results", examID)
		w := do(t, srv, "GET", url, teacherToken, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("teacher results: want 200, got %d — %s", w.Code, w.Body)
		}
		body := parseJSON(t, w)

		stats, _ := body["statistics"].(map[string]interface{})
		totalStudents := stats["total_students"].(float64)
		if totalStudents != 1 {
			t.Errorf("expected total_students=1, got %.0f", totalStudents)
		}
		avgScore := stats["avg_score"].(float64)
		if avgScore != 1.0 {
			t.Errorf("expected avg_score=1.0, got %.2f", avgScore)
		}
		maxScore := stats["max_score"].(float64)
		if maxScore != 1.0 {
			t.Errorf("expected max_score=1.0, got %.2f", maxScore)
		}

		results, _ := body["results"].([]interface{})
		if len(results) != 1 {
			t.Fatalf("expected 1 result row, got %d", len(results))
		}
	})

	// Suppress unused variable warnings
	_ = q1ID
	_ = q2ID
	_ = getFloat
}

// ══════════════════════════════════════════════════════════════════════════════
// Individual focused tests
// ══════════════════════════════════════════════════════════════════════════════

func TestHealth(t *testing.T) {
	srv := setupTestServer(t)
	w := do(t, srv, "GET", "/health", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("health check: want 200, got %d", w.Code)
	}
}

func TestAuth_BadInputs(t *testing.T) {
	srv := setupTestServer(t)

	tests := []struct {
		name     string
		body     interface{}
		wantCode int
	}{
		{"missing email", map[string]interface{}{"password": "123456"}, http.StatusBadRequest},
		{"missing password", map[string]interface{}{"email": "x@y.com"}, http.StatusBadRequest},
		{"invalid email", map[string]interface{}{"email": "notanemail", "password": "123456"}, http.StatusBadRequest},
		{"short password", map[string]interface{}{"full_name": "X", "email": "x@y.com", "password": "123"}, http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Try both register and login endpoints
			w := do(t, srv, "POST", "/api/auth/register", "", tc.body)
			if w.Code != tc.wantCode {
				// login also validates
				w2 := do(t, srv, "POST", "/api/auth/login", "", tc.body)
				if w2.Code != tc.wantCode && w.Code != tc.wantCode {
					t.Errorf("[%s] want %d, register got %d / login got %d", tc.name, tc.wantCode, w.Code, w2.Code)
				}
			}
		})
	}
}

func TestTeacher_ExamLifecycle(t *testing.T) {
	srv := setupTestServer(t)
	models.SeedData(config.DB)

	// Login teacher
	w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "teacher@test.com", "password": "123456",
	})
	body := parseJSON(t, w)
	token := getString(body, "token")

	now := time.Now()

	// Create exam
	w = do(t, srv, "POST", "/api/teacher/exams", token, map[string]interface{}{
		"title": "Test Exam", "duration": 10,
		"start_time": now.Add(-time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(time.Hour).Format(time.RFC3339),
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create exam: %d %s", w.Code, w.Body)
	}
	examBody := parseJSON(t, w)
	exam := examBody["exam"].(map[string]interface{})
	id := exam["id"].(float64)

	// Cannot publish with no questions
	w = do(t, srv, "PATCH", fmt.Sprintf("/api/teacher/exams/%.0f/publish", id), token, nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("publish with 0 questions: want 400, got %d", w.Code)
	}

	// Add 1 question
	w = do(t, srv, "POST", fmt.Sprintf("/api/teacher/exams/%.0f/questions", id), token, map[string]interface{}{
		"questions": []map[string]interface{}{{
			"prompt": "Test?", "option_a": "A", "option_b": "B", "option_c": "C", "option_d": "D",
			"correct_option": "A",
		}},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("add question: %d %s", w.Code, w.Body)
	}
	qBody := parseJSON(t, w)
	qs := qBody["questions"].([]interface{})
	qID := qs[0].(map[string]interface{})["id"].(float64)

	// Delete question
	w = do(t, srv, "DELETE", fmt.Sprintf("/api/teacher/questions/%.0f", qID), token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete question: %d %s", w.Code, w.Body)
	}

	// Re-add and publish
	w = do(t, srv, "POST", fmt.Sprintf("/api/teacher/exams/%.0f/questions", id), token, map[string]interface{}{
		"questions": []map[string]interface{}{{
			"prompt": "Final Q?", "option_a": "A", "option_b": "B", "option_c": "C", "option_d": "D",
			"correct_option": "B",
		}},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("re-add question: %d %s", w.Code, w.Body)
	}

	w = do(t, srv, "PATCH", fmt.Sprintf("/api/teacher/exams/%.0f/publish", id), token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("publish: %d %s", w.Code, w.Body)
	}

	// Verify exam detail shows questions
	w = do(t, srv, "GET", fmt.Sprintf("/api/teacher/exams/%.0f", id), token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get exam: %d %s", w.Code, w.Body)
	}
}

func TestStudent_GradingAccuracy(t *testing.T) {
	srv := setupTestServer(t)
	models.SeedData(config.DB)

	now := time.Now()

	// Get teacher token
	w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "teacher@test.com", "password": "123456",
	})
	teacherToken := getString(parseJSON(t, w), "token")

	// Create & publish exam with 4 questions (1pt each)
	w = do(t, srv, "POST", "/api/teacher/exams", teacherToken, map[string]interface{}{
		"title": "Grading Test", "duration": 60,
		"start_time": now.Add(-time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(time.Hour).Format(time.RFC3339),
	})
	examID := parseJSON(t, w)["exam"].(map[string]interface{})["id"].(float64)

	questions := []map[string]interface{}{
		{"prompt": "Q1?", "option_a": "A", "option_b": "B", "option_c": "C", "option_d": "D", "correct_option": "A", "points": 1.0},
		{"prompt": "Q2?", "option_a": "A", "option_b": "B", "option_c": "C", "option_d": "D", "correct_option": "B", "points": 1.0},
		{"prompt": "Q3?", "option_a": "A", "option_b": "B", "option_c": "C", "option_d": "D", "correct_option": "C", "points": 1.0},
		{"prompt": "Q4?", "option_a": "A", "option_b": "B", "option_c": "C", "option_d": "D", "correct_option": "D", "points": 1.0},
	}
	w = do(t, srv, "POST", fmt.Sprintf("/api/teacher/exams/%.0f/questions", examID), teacherToken, map[string]interface{}{"questions": questions})
	qs := parseJSON(t, w)["questions"].([]interface{})

	do(t, srv, "PATCH", fmt.Sprintf("/api/teacher/exams/%.0f/publish", examID), teacherToken, nil)

	// Register & login student
	do(t, srv, "POST", "/api/auth/register", "", map[string]interface{}{
		"full_name": "Grade Student", "email": "grade@exam.local", "password": "password123",
	})
	w = do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "grade@exam.local", "password": "password123",
	})
	studentToken := getString(parseJSON(t, w), "token")

	// Start exam
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/exams/%.0f/start", examID), studentToken, map[string]bool{"confirmed": true})
	startBody := parseJSON(t, w)
	attemptID := startBody["attempt"].(map[string]interface{})["id"].(float64)

	// Extract question IDs
	getQID := func(i int) float64 {
		return qs[i].(map[string]interface{})["id"].(float64)
	}

	// Answer: Q1=A(✓), Q2=A(✗), Q3=C(✓), Q4=skip → score=2, correct=2, wrong=2
	// (blank answers count as wrong so correct+wrong == total questions)
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/attempts/%.0f/submit", attemptID), studentToken, map[string]interface{}{
		"answers": []map[string]interface{}{
			{"question_id": int(getQID(0)), "chosen_option": "A"}, // correct
			{"question_id": int(getQID(1)), "chosen_option": "A"}, // wrong (B)
			{"question_id": int(getQID(2)), "chosen_option": "C"}, // correct
			// Q4 not answered = skip → counted as wrong
		},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("submit: %d %s", w.Code, w.Body)
	}
	result := parseJSON(t, w)["attempt"].(map[string]interface{})

	if score := result["score"].(float64); score != 2.0 {
		t.Errorf("score: want 2.0, got %.2f", score)
	}
	if cc := result["correct_count"].(float64); cc != 2 {
		t.Errorf("correct_count: want 2, got %.0f", cc)
	}
	if wc := result["wrong_count"].(float64); wc != 2 {
		t.Errorf("wrong_count: want 2 (1 wrong + 1 blank), got %.0f", wc)
	}
	totalCounted := result["correct_count"].(float64) + result["wrong_count"].(float64)
	if totalCounted != 4 {
		t.Errorf("correct+wrong should equal total questions (4), got %.0f", totalCounted)
	}
}

// ─── Security & business-rule focused tests ───────────────────────────────────

func TestSecurity_SQLi_DeleteQuestion_InvalidID(t *testing.T) {
	srv := setupTestServer(t)
	models.SeedData(config.DB)

	w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "teacher@test.com", "password": "123456",
	})
	token := getString(parseJSON(t, w), "token")

	// SQLi-like path param must be rejected by parseUintParam (400), not executed
	w = do(t, srv, "DELETE", "/api/teacher/questions/1%20OR%201=1", token, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("SQLi question_id: want 400, got %d — %s", w.Code, w.Body)
	}

	w = do(t, srv, "DELETE", "/api/teacher/questions/abc", token, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("non-numeric question_id: want 400, got %d — %s", w.Code, w.Body)
	}
}

func TestSecurity_IDOR_SaveForeignQuestion(t *testing.T) {
	srv := setupTestServer(t)
	models.SeedData(config.DB)
	now := time.Now()

	w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "teacher@test.com", "password": "123456",
	})
	teacherToken := getString(parseJSON(t, w), "token")

	// Exam A
	w = do(t, srv, "POST", "/api/teacher/exams", teacherToken, map[string]interface{}{
		"title": "Exam A", "duration": 30,
		"start_time": now.Add(-time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(time.Hour).Format(time.RFC3339),
	})
	examAID := parseJSON(t, w)["exam"].(map[string]interface{})["id"].(float64)
	w = do(t, srv, "POST", fmt.Sprintf("/api/teacher/exams/%.0f/questions", examAID), teacherToken, map[string]interface{}{
		"questions": []map[string]interface{}{{
			"prompt": "QA?", "option_a": "A", "option_b": "B", "option_c": "C", "option_d": "D",
			"correct_option": "A",
		}},
	})
	qAID := parseJSON(t, w)["questions"].([]interface{})[0].(map[string]interface{})["id"].(float64)
	do(t, srv, "PATCH", fmt.Sprintf("/api/teacher/exams/%.0f/publish", examAID), teacherToken, nil)

	// Exam B (foreign questions)
	w = do(t, srv, "POST", "/api/teacher/exams", teacherToken, map[string]interface{}{
		"title": "Exam B", "duration": 30,
		"start_time": now.Add(-time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(time.Hour).Format(time.RFC3339),
	})
	examBID := parseJSON(t, w)["exam"].(map[string]interface{})["id"].(float64)
	w = do(t, srv, "POST", fmt.Sprintf("/api/teacher/exams/%.0f/questions", examBID), teacherToken, map[string]interface{}{
		"questions": []map[string]interface{}{{
			"prompt": "QB?", "option_a": "A", "option_b": "B", "option_c": "C", "option_d": "D",
			"correct_option": "B",
		}},
	})
	qBID := parseJSON(t, w)["questions"].([]interface{})[0].(map[string]interface{})["id"].(float64)
	do(t, srv, "PATCH", fmt.Sprintf("/api/teacher/exams/%.0f/publish", examBID), teacherToken, nil)

	do(t, srv, "POST", "/api/auth/register", "", map[string]interface{}{
		"full_name": "IDOR Student", "email": "idor@exam.local", "password": "password123",
	})
	w = do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "idor@exam.local", "password": "password123",
	})
	studentToken := getString(parseJSON(t, w), "token")

	// Start Exam A
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/exams/%.0f/start", examAID), studentToken, map[string]bool{"confirmed": true})
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/exams/%.0f/start", examAID), studentToken, nil)
	attemptID := parseJSON(t, w)["attempt"].(map[string]interface{})["id"].(float64)

	// Try to save a question belonging to Exam B → must 400
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/attempts/%.0f/save", attemptID), studentToken, map[string]interface{}{
		"answers": []map[string]interface{}{
			{"question_id": int(qBID), "chosen_option": "A"},
		},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("IDOR foreign question_id: want 400, got %d — %s", w.Code, w.Body)
	}

	// Valid question from Exam A still works
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/attempts/%.0f/save", attemptID), studentToken, map[string]interface{}{
		"question_id": int(qAID), "chosen_option": "A",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("valid save: want 200, got %d — %s", w.Code, w.Body)
	}
}

func TestSecurity_InvalidChosenOption(t *testing.T) {
	srv := setupTestServer(t)
	models.SeedData(config.DB)
	now := time.Now()

	w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "teacher@test.com", "password": "123456",
	})
	teacherToken := getString(parseJSON(t, w), "token")

	w = do(t, srv, "POST", "/api/teacher/exams", teacherToken, map[string]interface{}{
		"title": "Opt Test", "duration": 30,
		"start_time": now.Add(-time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(time.Hour).Format(time.RFC3339),
	})
	examID := parseJSON(t, w)["exam"].(map[string]interface{})["id"].(float64)
	w = do(t, srv, "POST", fmt.Sprintf("/api/teacher/exams/%.0f/questions", examID), teacherToken, map[string]interface{}{
		"questions": []map[string]interface{}{{
			"prompt": "Q?", "option_a": "A", "option_b": "B", "option_c": "C", "option_d": "D",
			"correct_option": "A",
		}},
	})
	qID := parseJSON(t, w)["questions"].([]interface{})[0].(map[string]interface{})["id"].(float64)
	do(t, srv, "PATCH", fmt.Sprintf("/api/teacher/exams/%.0f/publish", examID), teacherToken, nil)

	do(t, srv, "POST", "/api/auth/register", "", map[string]interface{}{
		"full_name": "Opt Student", "email": "opt@exam.local", "password": "password123",
	})
	w = do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "opt@exam.local", "password": "password123",
	})
	studentToken := getString(parseJSON(t, w), "token")

	w = do(t, srv, "POST", fmt.Sprintf("/api/student/exams/%.0f/start", examID), studentToken, map[string]bool{"confirmed": true})
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/exams/%.0f/start", examID), studentToken, nil)
	attemptID := parseJSON(t, w)["attempt"].(map[string]interface{})["id"].(float64)

	w = do(t, srv, "POST", fmt.Sprintf("/api/student/attempts/%.0f/save", attemptID), studentToken, map[string]interface{}{
		"answers": []map[string]interface{}{
			{"question_id": int(qID), "chosen_option": "Z"},
		},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("chosen_option=Z: want 400, got %d — %s", w.Code, w.Body)
	}
}

func TestTeacher_Stats_OnlyCompleted(t *testing.T) {
	srv := setupTestServer(t)
	models.SeedData(config.DB)
	now := time.Now()

	w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "teacher@test.com", "password": "123456",
	})
	teacherToken := getString(parseJSON(t, w), "token")

	w = do(t, srv, "POST", "/api/teacher/exams", teacherToken, map[string]interface{}{
		"title": "Stats Exam", "duration": 60,
		"start_time": now.Add(-time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(2 * time.Hour).Format(time.RFC3339),
	})
	examID := parseJSON(t, w)["exam"].(map[string]interface{})["id"].(float64)
	w = do(t, srv, "POST", fmt.Sprintf("/api/teacher/exams/%.0f/questions", examID), teacherToken, map[string]interface{}{
		"questions": []map[string]interface{}{{
			"prompt": "Q?", "option_a": "A", "option_b": "B", "option_c": "C", "option_d": "D",
			"correct_option": "A", "points": 5.0,
		}},
	})
	qID := parseJSON(t, w)["questions"].([]interface{})[0].(map[string]interface{})["id"].(float64)
	do(t, srv, "PATCH", fmt.Sprintf("/api/teacher/exams/%.0f/publish", examID), teacherToken, nil)

	// Student 1: complete with score 5
	do(t, srv, "POST", "/api/auth/register", "", map[string]interface{}{
		"full_name": "S1", "email": "s1stats@exam.local", "password": "password123",
	})
	w = do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "s1stats@exam.local", "password": "password123",
	})
	s1Token := getString(parseJSON(t, w), "token")
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/exams/%.0f/start", examID), s1Token, map[string]bool{"confirmed": true})
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/exams/%.0f/start", examID), s1Token, nil)
	a1 := parseJSON(t, w)["attempt"].(map[string]interface{})["id"].(float64)
	do(t, srv, "POST", fmt.Sprintf("/api/student/attempts/%.0f/submit", a1), s1Token, map[string]interface{}{
		"answers": []map[string]interface{}{{"question_id": int(qID), "chosen_option": "A"}},
	})

	// Student 2: start but leave in_progress (score still 0) — must NOT skew avg/min/max
	do(t, srv, "POST", "/api/auth/register", "", map[string]interface{}{
		"full_name": "S2", "email": "s2stats@exam.local", "password": "password123",
	})
	w = do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "s2stats@exam.local", "password": "password123",
	})
	s2Token := getString(parseJSON(t, w), "token")
	do(t, srv, "POST", fmt.Sprintf("/api/student/exams/%.0f/start", examID), s2Token, map[string]bool{"confirmed": true})
	do(t, srv, "POST", fmt.Sprintf("/api/student/exams/%.0f/start", examID), s2Token, nil)

	w = do(t, srv, "GET", fmt.Sprintf("/api/teacher/exams/%.0f/results", examID), teacherToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("results: want 200, got %d — %s", w.Code, w.Body)
	}
	body := parseJSON(t, w)
	stats := body["statistics"].(map[string]interface{})
	if stats["total_students"].(float64) != 2 {
		t.Errorf("total_students: want 2, got %v", stats["total_students"])
	}
	if stats["avg_score"].(float64) != 5.0 {
		t.Errorf("avg_score should ignore in_progress: want 5.0, got %v", stats["avg_score"])
	}
	if stats["max_score"].(float64) != 5.0 {
		t.Errorf("max_score: want 5.0, got %v", stats["max_score"])
	}
	if stats["min_score"].(float64) != 5.0 {
		t.Errorf("min_score should ignore in_progress zero: want 5.0, got %v", stats["min_score"])
	}
}

func TestAuth_EmailNormalization(t *testing.T) {
	srv := setupTestServer(t)

	w := do(t, srv, "POST", "/api/auth/register", "", map[string]interface{}{
		"full_name": "Norm User", "email": "  Norm.User@Exam.LOCAL ", "password": "password123",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("register: want 201, got %d — %s", w.Code, w.Body)
	}
	email := getString(parseJSON(t, w), "user", "email")
	if email != "norm.user@exam.local" {
		t.Errorf("email not normalized: got %q", email)
	}

	w = do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "NORM.USER@EXAM.LOCAL", "password": "password123",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("login with different case: want 200, got %d — %s", w.Code, w.Body)
	}
}

func TestStudent_ConsentAndDeadlineAutograding(t *testing.T) {
	srv := setupTestServer(t)

	w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "student@test.com", "password": "123456",
	})
	studentToken := getString(parseJSON(t, w), "token")

	var exam models.Exam
	if err := config.DB.Where("exam_code = ?", "IT2026").First(&exam).Error; err != nil {
		t.Fatalf("seeded exam not found: %v", err)
	}
	startURL := fmt.Sprintf("/api/student/exams/%d/start", exam.ID)

	w = do(t, srv, "POST", startURL, studentToken, nil)
	if w.Code != http.StatusPreconditionRequired {
		t.Fatalf("start without consent: want 428, got %d — %s", w.Code, w.Body)
	}
	var attemptCount int64
	config.DB.Model(&models.ExamAttempt{}).Count(&attemptCount)
	if attemptCount != 0 {
		t.Fatalf("start without consent created %d attempts", attemptCount)
	}

	w = do(t, srv, "POST", startURL, studentToken, map[string]bool{"confirmed": true})
	if w.Code != http.StatusOK {
		t.Fatalf("confirmed start: want 200, got %d — %s", w.Code, w.Body)
	}
	body := parseJSON(t, w)
	attemptData := body["attempt"].(map[string]interface{})
	attemptID := uint(attemptData["id"].(float64))

	var question models.Question
	if err := config.DB.Where("exam_id = ?", exam.ID).Order("sort_order ASC").First(&question).Error; err != nil {
		t.Fatalf("seeded question not found: %v", err)
	}
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/attempts/%d/save", attemptID), studentToken, map[string]interface{}{
		"question_id": question.ID, "chosen_option": "A",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("save answer: want 200, got %d — %s", w.Code, w.Body)
	}

	pastDeadline := time.Now().Add(-time.Duration(exam.Duration+1) * time.Minute)
	if err := config.DB.Model(&models.ExamAttempt{}).Where("id = ?", attemptID).
		Update("started_at", pastDeadline).Error; err != nil {
		t.Fatalf("expire attempt: %v", err)
	}
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/attempts/%d/save", attemptID), studentToken, map[string]interface{}{
		"question_id": question.ID, "chosen_option": question.CorrectOption,
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("save after deadline: want 409 auto-submit, got %d — %s", w.Code, w.Body)
	}
	if !parseJSON(t, w)["auto_submitted"].(bool) {
		t.Fatal("expired save response should report automatic submission")
	}

	w = do(t, srv, "POST", startURL, studentToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("resume expired exam: want 200, got %d — %s", w.Code, w.Body)
	}
	if !parseJSON(t, w)["auto_submitted"].(bool) {
		t.Fatal("expired resume response should report automatic submission")
	}

	var attempt models.ExamAttempt
	if err := config.DB.First(&attempt, attemptID).Error; err != nil {
		t.Fatalf("load auto-submitted attempt: %v", err)
	}
	if attempt.Status != "completed" {
		t.Fatalf("expired attempt status: want completed, got %s", attempt.Status)
	}
	if attempt.Score != 0 {
		t.Fatalf("auto-submit should grade saved wrong answer only; want 0, got %v", attempt.Score)
	}

	w = do(t, srv, "POST", fmt.Sprintf("/api/student/attempts/%d/save", attemptID), studentToken, map[string]interface{}{
		"question_id": question.ID, "chosen_option": question.CorrectOption,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("save after auto-submit: want 400, got %d — %s", w.Code, w.Body)
	}

	w = do(t, srv, "GET", fmt.Sprintf("/api/student/attempts/%d/result", attemptID), studentToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("view auto-submitted result: want 200, got %d — %s", w.Code, w.Body)
	}
	if score := getFloat(parseJSON(t, w), "attempt", "score"); score != 0 {
		t.Fatalf("auto-submitted result score: want 0, got %v", score)
	}
}

func TestStudentHistoryAndAppealsOwnershipAndWorkflow(t *testing.T) {
	srv := setupTestServer(t)

	login := func(email string) string {
		t.Helper()
		w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
			"email": email, "password": "123456",
		})
		if w.Code != http.StatusOK {
			t.Fatalf("login %s: want 200, got %d — %s", email, w.Code, w.Body)
		}
		return getString(parseJSON(t, w), "token")
	}
	studentToken := login("student@test.com")
	teacherToken := login("teacher@test.com")

	var exam models.Exam
	if err := config.DB.Where("exam_code = ?", "IT2026").First(&exam).Error; err != nil {
		t.Fatalf("seeded exam not found: %v", err)
	}
	w := do(t, srv, "POST", fmt.Sprintf("/api/student/exams/%d/start", exam.ID), studentToken, map[string]bool{"confirmed": true})
	if w.Code != http.StatusOK {
		t.Fatalf("start exam: want 200, got %d — %s", w.Code, w.Body)
	}
	startBody := parseJSON(t, w)
	attemptID := uint(startBody["attempt"].(map[string]interface{})["id"].(float64))
	questions := startBody["questions"].([]interface{})
	firstQuestion := questions[0].(map[string]interface{})
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/attempts/%d/submit", attemptID), studentToken, map[string]interface{}{
		"answers": []map[string]interface{}{{
			"question_id": firstQuestion["id"], "chosen_option": "A",
		}},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("submit exam: want 200, got %d — %s", w.Code, w.Body)
	}

	w = do(t, srv, "GET", "/api/student/history?page=1&page_size=5", studentToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("history: want 200, got %d — %s", w.Code, w.Body)
	}
	history := parseJSON(t, w)
	items := history["items"].([]interface{})
	if len(items) != 1 || history["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatalf("expected one owned history entry, got %s", w.Body)
	}
	historyItem := items[0].(map[string]interface{})
	if historyItem["max_score"].(float64) != 5 {
		t.Errorf("history max_score: want 5, got %v", historyItem["max_score"])
	}

	w = do(t, srv, "GET", fmt.Sprintf("/api/student/history/%d", attemptID), studentToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("history detail: want 200, got %d — %s", w.Code, w.Body)
	}
	detailAnswers := parseJSON(t, w)["answers"].([]interface{})
	if len(detailAnswers) != 5 {
		t.Fatalf("history detail should include all exam questions, got %d", len(detailAnswers))
	}
	if detailAnswers[0].(map[string]interface{})["correct_option"] == nil {
		t.Fatal("student history detail should show the correct option")
	}

	otherStudent := models.User{
		Email: "other-history-student@test.com", PasswordHash: "unused",
		FullName: "Other Student", Role: models.RoleStudent,
	}
	if err := config.DB.Create(&otherStudent).Error; err != nil {
		t.Fatalf("create other student: %v", err)
	}
	otherToken, err := utils.GenerateJWT(otherStudent.ID, otherStudent.Email, otherStudent.Role)
	if err != nil {
		t.Fatalf("generate other student token: %v", err)
	}
	w = do(t, srv, "GET", fmt.Sprintf("/api/student/history/%d", attemptID), otherToken, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("other student history access: want 404, got %d", w.Code)
	}

	createAppeal := func(content string) *httptest.ResponseRecorder {
		t.Helper()
		return do(t, srv, "POST", "/api/student/appeals", studentToken, map[string]interface{}{
			"submissionId": attemptID,
			"title":        "Phúc khảo câu 1",
			"content":      content,
		})
	}
	w = createAppeal("Quá ngắn")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("short appeal: want 400, got %d — %s", w.Code, w.Body)
	}
	w = createAppeal(strings.Repeat("x", 1001))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("long appeal: want 400, got %d — %s", w.Code, w.Body)
	}
	w = do(t, srv, "POST", "/api/student/appeals", otherToken, map[string]interface{}{
		"submissionId": attemptID, "content": "Tôi muốn khiếu nại điểm bài thi này.",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("other student appeal: want 404, got %d — %s", w.Code, w.Body)
	}
	w = createAppeal("Em cho rằng đáp án của em đúng.")
	if w.Code != http.StatusCreated {
		t.Fatalf("create appeal: want 201, got %d — %s", w.Code, w.Body)
	}
	appeal := parseJSON(t, w)["appeal"].(map[string]interface{})
	appealID := uint(appeal["id"].(float64))
	if appeal["status"] != models.AppealStatusPending {
		t.Fatalf("new appeal status: want PENDING, got %v", appeal["status"])
	}

	w = do(t, srv, "GET", fmt.Sprintf("/api/teacher/exams/%d/appeals?status=PENDING", exam.ID), teacherToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("teacher list appeals: want 200, got %d — %s", w.Code, w.Body)
	}
	appealItems := parseJSON(t, w)["items"].([]interface{})
	if len(appealItems) != 1 || appealItems[0].(map[string]interface{})["student_name"] != "Trần Thị Học" {
		t.Fatalf("teacher appeal list should include student information, got %s", w.Body)
	}
	w = do(t, srv, "GET", fmt.Sprintf("/api/teacher/appeals/%d", appealID), teacherToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("teacher appeal detail: want 200, got %d — %s", w.Code, w.Body)
	}
	detail := parseJSON(t, w)
	if detail["submission"].(map[string]interface{})["max_score"].(float64) != 5 {
		t.Errorf("teacher appeal detail should include maximum score: %s", w.Body)
	}

	otherTeacher := models.User{
		Email: "other-appeal-teacher@test.com", PasswordHash: "unused",
		FullName: "Other Teacher", Role: models.RoleTeacher,
	}
	if err := config.DB.Create(&otherTeacher).Error; err != nil {
		t.Fatalf("create other teacher: %v", err)
	}
	otherTeacherToken, err := utils.GenerateJWT(otherTeacher.ID, otherTeacher.Email, otherTeacher.Role)
	if err != nil {
		t.Fatalf("generate other teacher token: %v", err)
	}
	w = do(t, srv, "GET", fmt.Sprintf("/api/teacher/appeals/%d", appealID), otherTeacherToken, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("other teacher appeal access: want 404, got %d", w.Code)
	}
	w = do(t, srv, "GET", fmt.Sprintf("/api/teacher/exams/%d/appeals", exam.ID), otherTeacherToken, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("other teacher exam appeals access: want 404, got %d", w.Code)
	}
	w = do(t, srv, "PATCH", fmt.Sprintf("/api/teacher/appeals/%d/reply", appealID), otherTeacherToken, map[string]string{
		"response": "Đã kiểm tra lại nội dung câu hỏi.",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("other teacher appeal reply: want 404, got %d", w.Code)
	}

	w = do(t, srv, "PATCH", fmt.Sprintf("/api/teacher/appeals/%d/reply", appealID), teacherToken, map[string]string{
		"response": "Đã kiểm tra lại nội dung câu hỏi.",
	})
	if w.Code != http.StatusOK || parseJSON(t, w)["appeal"].(map[string]interface{})["status"] != models.AppealStatusReplied {
		t.Fatalf("teacher reply failed: status=%d body=%s", w.Code, w.Body)
	}
	w = do(t, srv, "PATCH", fmt.Sprintf("/api/teacher/appeals/%d/resolve", appealID), teacherToken, nil)
	if w.Code != http.StatusOK || parseJSON(t, w)["appeal"].(map[string]interface{})["status"] != models.AppealStatusResolved {
		t.Fatalf("teacher resolve failed: status=%d body=%s", w.Code, w.Body)
	}
	w = do(t, srv, "PATCH", fmt.Sprintf("/api/teacher/appeals/%d/reject", appealID), teacherToken, nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("closed appeal update: want 409, got %d — %s", w.Code, w.Body)
	}

	w = do(t, srv, "GET", fmt.Sprintf("/api/student/history/%d", attemptID), studentToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("reload history detail: want 200, got %d — %s", w.Code, w.Body)
	}
	storedAppeals := parseJSON(t, w)["appeals"].([]interface{})
	storedAppeal := storedAppeals[0].(map[string]interface{})
	if storedAppeal["status"] != models.AppealStatusResolved || storedAppeal["teacher_response"] == nil {
		t.Fatalf("student history should preserve appeal response and final status: %s", w.Body)
	}

	w = createAppeal("Em xin kiểm tra lại điểm tổng của bài thi.")
	if w.Code != http.StatusCreated {
		t.Fatalf("create second appeal: want 201, got %d — %s", w.Code, w.Body)
	}
	secondAppealID := uint(parseJSON(t, w)["appeal"].(map[string]interface{})["id"].(float64))
	w = do(t, srv, "PATCH", fmt.Sprintf("/api/teacher/appeals/%d/reject", secondAppealID), teacherToken, nil)
	if w.Code != http.StatusOK || parseJSON(t, w)["appeal"].(map[string]interface{})["status"] != models.AppealStatusRejected {
		t.Fatalf("teacher reject failed: status=%d body=%s", w.Code, w.Body)
	}
	w = do(t, srv, "GET", "/api/student/history?appeal_status=REJECTED", studentToken, nil)
	if w.Code != http.StatusOK || parseJSON(t, w)["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatalf("student appeal status filter failed: status=%d body=%s", w.Code, w.Body)
	}
}

func TestSeedDataEnsuresDemoUsersOnExistingDatabase(t *testing.T) {
	dbPath := fmt.Sprintf("file:seed_existing_%d?mode=memory&cache=shared&_foreign_keys=on", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test DB: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get test DB connection: %v", err)
	}
	defer sqlDB.Close()

	if err := db.AutoMigrate(&models.User{}, &models.Exam{}, &models.Question{}); err != nil {
		t.Fatalf("migrate test DB: %v", err)
	}
	if err := db.Create(&models.User{
		Email: "already-existing@test.com", PasswordHash: "existing", FullName: "Existing User", Role: models.RoleStudent,
	}).Error; err != nil {
		t.Fatalf("create existing user: %v", err)
	}

	models.SeedData(db)
	for _, email := range []string{"student@test.com", "student2@test.com", "student3@test.com"} {
		var user models.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			t.Fatalf("seeded student %s missing from populated database: %v", email, err)
		}
		if user.Role != models.RoleStudent {
			t.Errorf("seeded account %s role: want student, got %s", email, user.Role)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("123456")); err != nil {
			t.Errorf("seeded account %s should accept demo password: %v", email, err)
		}
	}

	models.SeedData(db)

	var userCount, examCount, questionCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if err := db.Model(&models.Exam{}).Count(&examCount).Error; err != nil {
		t.Fatalf("count exams: %v", err)
	}
	if err := db.Model(&models.Question{}).Count(&questionCount).Error; err != nil {
		t.Fatalf("count questions: %v", err)
	}
	if userCount != 5 || examCount != 1 || questionCount != 5 {
		t.Fatalf("repeated seed duplicated data: users=%d exams=%d questions=%d", userCount, examCount, questionCount)
	}
}

func TestTeacher_ExamCRUD_And_BulkQuestions(t *testing.T) {
	srv := setupTestServer(t)

	// Login teacher
	w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "teacher@test.com", "password": "123456",
	})
	teacherToken := getString(parseJSON(t, w), "token")
	now := time.Now()

	// 1. Create exam
	w = do(t, srv, "POST", "/api/teacher/exams", teacherToken, map[string]interface{}{
		"title": "CRUD Exam", "duration": 20,
		"start_time": now.Add(-time.Hour).Format(time.RFC3339),
		"end_time":   now.Add(2 * time.Hour).Format(time.RFC3339),
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create exam: %d %s", w.Code, w.Body)
	}
	examID := uint(parseJSON(t, w)["exam"].(map[string]interface{})["id"].(float64))

	// 2. Update exam (PUT)
	newTitle := "Updated CRUD Exam"
	newDesc := "Description for updated exam"
	durMin := 35
	w = do(t, srv, "PUT", fmt.Sprintf("/api/teacher/exams/%d", examID), teacherToken, map[string]interface{}{
		"title":            newTitle,
		"description":      newDesc,
		"duration_minutes": durMin,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("update exam: %d %s", w.Code, w.Body)
	}
	upExam := parseJSON(t, w)["exam"].(map[string]interface{})
	if upExam["title"].(string) != newTitle || upExam["description"].(string) != newDesc || int(upExam["duration"].(float64)) != 35 {
		t.Fatalf("updated exam fields mismatch: %v", upExam)
	}

	// 3. Bulk import questions (POST /api/teacher/exams/:id/questions/bulk)
	bulkQs := []map[string]interface{}{
		{
			"prompt": "Bulk Q1", "option_a": "A1", "option_b": "B1", "option_c": "C1", "option_d": "D1",
			"correct_option": "A", "points": 2.0,
		},
		{
			"prompt": "Bulk Q2", "option_a": "A2", "option_b": "B2", "option_c": "C2", "option_d": "D2",
			"correct_option": "B", "points": 3.0,
		},
	}
	w = do(t, srv, "POST", fmt.Sprintf("/api/teacher/exams/%d/questions/bulk", examID), teacherToken, map[string]interface{}{
		"questions": bulkQs,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("bulk add questions: %d %s", w.Code, w.Body)
	}
	createdQs := parseJSON(t, w)["questions"].([]interface{})
	if len(createdQs) != 2 {
		t.Fatalf("bulk add should create 2 questions, got %d", len(createdQs))
	}
	q1ID := uint(createdQs[0].(map[string]interface{})["id"].(float64))

	// 4. Update single question (PUT /api/teacher/questions/:id)
	w = do(t, srv, "PUT", fmt.Sprintf("/api/teacher/questions/%d", q1ID), teacherToken, map[string]interface{}{
		"prompt": "Updated Bulk Q1 Prompt", "option_a": "New A", "option_b": "New B",
		"option_c": "New C", "option_d": "New D", "correct_option": "C", "points": 2.5,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("update question: %d %s", w.Code, w.Body)
	}
	upQ := parseJSON(t, w)["question"].(map[string]interface{})
	if upQ["prompt"].(string) != "Updated Bulk Q1 Prompt" || upQ["correct_option"].(string) != "C" || upQ["points"].(float64) != 2.5 {
		t.Fatalf("updated question mismatch: %v", upQ)
	}

	// 5. Delete exam (DELETE /api/teacher/exams/:id)
	w = do(t, srv, "DELETE", fmt.Sprintf("/api/teacher/exams/%d", examID), teacherToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete exam: %d %s", w.Code, w.Body)
	}
	w = do(t, srv, "GET", fmt.Sprintf("/api/teacher/exams/%d", examID), teacherToken, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("deleted exam should return 404, got %d", w.Code)
	}
}

func TestTeacher_ScoreUpdate_And_ResetAttempt(t *testing.T) {
	srv := setupTestServer(t)

	w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "teacher@test.com", "password": "123456",
	})
	teacherToken := getString(parseJSON(t, w), "token")

	w = do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "student@test.com", "password": "123456",
	})
	studentToken := getString(parseJSON(t, w), "token")

	var exam models.Exam
	config.DB.Where("exam_code = ?", "IT2026").First(&exam)

	// Student starts and submits exam
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/exams/%d/start", exam.ID), studentToken, map[string]bool{"confirmed": true})
	attemptID := uint(parseJSON(t, w)["attempt"].(map[string]interface{})["id"].(float64))

	w = do(t, srv, "POST", fmt.Sprintf("/api/student/attempts/%d/submit", attemptID), studentToken, map[string]interface{}{})
	if w.Code != http.StatusOK {
		t.Fatalf("submit: %d %s", w.Code, w.Body)
	}

	// Teacher updates score manually with note (PATCH /api/teacher/attempts/:id/score)
	w = do(t, srv, "PATCH", fmt.Sprintf("/api/teacher/attempts/%d/score", attemptID), teacherToken, map[string]interface{}{
		"score": 4.5,
		"note":  "Cộng điểm thưởng câu khó",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("teacher update score: %d %s", w.Code, w.Body)
	}
	updatedAttempt := parseJSON(t, w)["attempt"].(map[string]interface{})
	if updatedAttempt["score"].(float64) != 4.5 || updatedAttempt["teacher_note"].(string) != "Cộng điểm thưởng câu khó" {
		t.Fatalf("score update mismatch: %v", updatedAttempt)
	}

	// Teacher resets attempt (DELETE /api/teacher/attempts/:id)
	w = do(t, srv, "DELETE", fmt.Sprintf("/api/teacher/attempts/%d", attemptID), teacherToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("reset attempt: %d %s", w.Code, w.Body)
	}

	// Student can now take exam again from scratch
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/exams/%d/start", exam.ID), studentToken, map[string]bool{"confirmed": true})
	if w.Code != http.StatusOK {
		t.Fatalf("retake exam after reset: want 200, got %d — %s", w.Code, w.Body)
	}
	newAttemptID := uint(parseJSON(t, w)["attempt"].(map[string]interface{})["id"].(float64))
	if newAttemptID == 0 || newAttemptID == attemptID {
		t.Fatalf("new attempt expected, got %d", newAttemptID)
	}
}

func TestTeacher_StudentManagementCRUD(t *testing.T) {
	srv := setupTestServer(t)

	w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "teacher@test.com", "password": "123456",
	})
	teacherToken := getString(parseJSON(t, w), "token")

	// 1. List students (GET /api/teacher/students)
	w = do(t, srv, "GET", "/api/teacher/students", teacherToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list students: %d %s", w.Code, w.Body)
	}
	students := parseJSON(t, w)["students"].([]interface{})
	if len(students) < 3 {
		t.Fatalf("expected at least 3 seeded students, got %d", len(students))
	}

	// 2. Create student (POST /api/teacher/students)
	newEmail := "newstudent@univ.edu.vn"
	w = do(t, srv, "POST", "/api/teacher/students", teacherToken, map[string]interface{}{
		"full_name": "Sinh Viên Mới",
		"email":     newEmail,
		"password":  "studentpass123",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create student: %d %s", w.Code, w.Body)
	}
	studentID := uint(parseJSON(t, w)["student"].(map[string]interface{})["id"].(float64))

	// Verify new student can login
	w = do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": newEmail, "password": "studentpass123",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("login new student: %d %s", w.Code, w.Body)
	}

	// 3. Search students with ?q=
	w = do(t, srv, "GET", "/api/teacher/students?q=newstudent", teacherToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("search students: %d %s", w.Code, w.Body)
	}
	searchRes := parseJSON(t, w)["students"].([]interface{})
	if len(searchRes) != 1 {
		t.Fatalf("search should find 1 student, got %d", len(searchRes))
	}

	// 4. Update student (PUT /api/teacher/students/:id)
	w = do(t, srv, "PUT", fmt.Sprintf("/api/teacher/students/%d", studentID), teacherToken, map[string]interface{}{
		"full_name": "Sinh Viên Đã Sửa",
		"password":  "newpass456789",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("update student: %d %s", w.Code, w.Body)
	}

	// Verify login with updated password
	w = do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": newEmail, "password": "newpass456789",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("login with updated password: %d %s", w.Code, w.Body)
	}

	// 5. Delete student (DELETE /api/teacher/students/:id)
	w = do(t, srv, "DELETE", fmt.Sprintf("/api/teacher/students/%d", studentID), teacherToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete student: %d %s", w.Code, w.Body)
	}

	// Verify student cannot login anymore
	w = do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": newEmail, "password": "newpass456789",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("deleted student login should fail 401, got %d", w.Code)
	}
}

func TestStudent_AvailableExams(t *testing.T) {
	srv := setupTestServer(t)

	w := do(t, srv, "POST", "/api/auth/login", "", map[string]interface{}{
		"email": "student@test.com", "password": "123456",
	})
	studentToken := getString(parseJSON(t, w), "token")

	// Get available exams (GET /api/student/exams/available)
	w = do(t, srv, "GET", "/api/student/exams/available", studentToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("available exams: %d %s", w.Code, w.Body)
	}
	exams := parseJSON(t, w)["exams"].([]interface{})
	if len(exams) == 0 {
		t.Fatal("expected at least 1 published exam in available list")
	}
	first := exams[0].(map[string]interface{})
	if first["exam_code"].(string) != "IT2026" {
		t.Errorf("expected IT2026, got %v", first["exam_code"])
	}
	if first["attempt_status"].(string) != "not_attempted" {
		t.Errorf("expected not_attempted, got %v", first["attempt_status"])
	}
}
