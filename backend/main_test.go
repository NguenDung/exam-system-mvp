package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"exam-system-mvp/config"
	"exam-system-mvp/models"
	"exam-system-mvp/router"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ─── Test helpers ─────────────────────────────────────────────────────────────

func setupTestServer(t *testing.T) *gin.Engine {
	t.Helper()

	// Use a unique SQLite file per test run (in-process, auto-cleaned)
	dbPath := fmt.Sprintf("file:test_%d?mode=memory&cache=shared&_foreign_keys=on", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
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
		var u models.User
		if err := config.DB.Where("email = ?", "teacher@test.com").First(&u).Error; err != nil {
			t.Fatalf("seeded teacher not found: %v", err)
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
					"prompt":         "Extra question",
					"option_a":       "A", "option_b": "B", "option_c": "C", "option_d": "D",
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
		w := do(t, srv, "POST", url, studentToken, nil)
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
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/exams/%.0f/start", examID), studentToken, nil)
	startBody := parseJSON(t, w)
	attemptID := startBody["attempt"].(map[string]interface{})["id"].(float64)

	// Extract question IDs
	getQID := func(i int) float64 {
		return qs[i].(map[string]interface{})["id"].(float64)
	}

	// Answer: Q1=A(✓), Q2=A(✗), Q3=C(✓), Q4=skip → score=2, correct=2, wrong=1
	w = do(t, srv, "POST", fmt.Sprintf("/api/student/attempts/%.0f/submit", attemptID), studentToken, map[string]interface{}{
		"answers": []map[string]interface{}{
			{"question_id": int(getQID(0)), "chosen_option": "A"}, // correct
			{"question_id": int(getQID(1)), "chosen_option": "A"}, // wrong (B)
			{"question_id": int(getQID(2)), "chosen_option": "C"}, // correct
			// Q4 not answered = skip
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
	if wc := result["wrong_count"].(float64); wc != 1 {
		t.Errorf("wrong_count: want 1, got %.0f", wc)
	}
}
