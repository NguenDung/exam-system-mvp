package handlers

import (
	"errors"
	"exam-system-mvp/apiresponse"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"exam-system-mvp/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultHistoryPageSize = 10
	maxHistoryPageSize     = 100
)

var errAppealClosed = errors.New("appeal is already closed")

type pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

func parsePagination(c *gin.Context) pagination {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(defaultHistoryPageSize)))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultHistoryPageSize
	}
	if pageSize > maxHistoryPageSize {
		pageSize = maxHistoryPageSize
	}
	return pagination{Page: page, PageSize: pageSize}
}

type StudentHistoryItem struct {
	SubmissionID uint       `json:"submission_id"`
	ExamID       uint       `json:"exam_id"`
	ExamTitle    string     `json:"exam_title"`
	PublishedAt  *time.Time `json:"published_at"`
	StartedAt    time.Time  `json:"started_at"`
	SubmittedAt  *time.Time `json:"submitted_at"`
	Score        float64    `json:"score"`
	MaxScore     float64    `json:"max_score"`
	AppealStatus string     `json:"appeal_status"`
	AppealCount  int        `json:"appeal_count"`
}

// StudentGetHistory returns a paginated list of completed exams taken by the student.
func (h *StudentHandler) StudentGetHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// Proactively grade any expired in-progress attempts so they appear in history
	var examIDs []uint
	if err := h.db.Model(&models.ExamAttempt{}).
		Where("user_id = ? AND status = ?", userID, "in_progress").
		Distinct("exam_id").
		Pluck("exam_id", &examIDs).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load exam attempts", err)
		return
	}
	for _, eid := range examIDs {
		if err := gradeExpiredAttempts(h.db, eid); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "Failed to submit expired attempts", err)
			return
		}
	}

	paging := parsePagination(c)
	query := h.db.Table("exam_attempts ea").
		Joins("JOIN exams e ON e.id = ea.exam_id").
		Where("ea.user_id = ? AND ea.status = ?", userID, "completed")

	if search := strings.TrimSpace(c.Query("q")); search != "" {
		query = query.Where("LOWER(e.title) LIKE ?", "%"+strings.ToLower(search)+"%")
	}
	if status := strings.TrimSpace(c.Query("appeal_status")); status != "" {
		if !isAppealStatus(status) {
			apiresponse.Error(c, http.StatusBadRequest, "Invalid appeal_status")
			return
		}
		query = query.Where(`
			EXISTS (
				SELECT 1 FROM appeals ap
				WHERE ap.submission_id = ea.id AND ap.student_id = ea.user_id AND ap.status = ?
			)
		`, status)
	}
	if err := query.Count(&paging.Total).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load exam history", err)
		return
	}

	items := make([]StudentHistoryItem, 0)
	err := query.Select(`
		ea.id AS submission_id,
		e.id AS exam_id,
		e.title AS exam_title,
		COALESCE(e.published_at, e.updated_at, e.created_at) AS published_at,
		ea.started_at,
		ea.submitted_at,
		ea.score,
		COALESCE((SELECT SUM(q.points) FROM questions q WHERE q.exam_id = e.id), 0) AS max_score,
		COALESCE((
			SELECT ap.status FROM appeals ap
			WHERE ap.submission_id = ea.id AND ap.student_id = ea.user_id
			ORDER BY ap.created_at DESC, ap.id DESC LIMIT 1
		), '') AS appeal_status,
		(SELECT COUNT(*) FROM appeals ap
		 WHERE ap.submission_id = ea.id AND ap.student_id = ea.user_id) AS appeal_count
	`).
		Order("ea.submitted_at DESC, ea.id DESC").
		Offset((paging.Page - 1) * paging.PageSize).
		Limit(paging.PageSize).
		Scan(&items).Error
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load exam history", err)
		return
	}

	apiresponse.JSON(c, http.StatusOK, HistoryResponse{ResponseMeta: ResponseMeta{Code: "EXAM_HISTORY_LISTED"}, Items: items, Pagination: paging})
}

type HistoryAnswer struct {
	QuestionID    uint    `json:"question_id"`
	SortOrder     int     `json:"sort_order"`
	Prompt        string  `json:"prompt"`
	OptionA       string  `json:"option_a"`
	OptionB       string  `json:"option_b"`
	OptionC       string  `json:"option_c"`
	OptionD       string  `json:"option_d"`
	ChosenOption  *string `json:"chosen_option"`
	CorrectOption string  `json:"correct_option"`
	IsCorrect     bool    `json:"is_correct"`
	Points        float64 `json:"points"`
	PointsAwarded float64 `json:"points_awarded"`
}

// StudentGetHistoryDetail returns full details of a completed exam submission including review answers and appeals.
func (h *StudentHandler) StudentGetHistoryDetail(c *gin.Context) {
	userID, _ := c.Get("user_id")
	submissionID, ok := parseUintParam(c, "submissionId")
	if !ok {
		return
	}

	var attempt models.ExamAttempt
	if err := h.db.Where("id = ? AND user_id = ? AND status = 'completed'", submissionID, userID).
		First(&attempt).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Submission not found", "Failed to load submission")
		return
	}

	var exam models.Exam
	if err := h.db.First(&exam, attempt.ExamID).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Exam not found", "Failed to load exam")
		return
	}

	answers := make([]HistoryAnswer, 0)
	err := h.db.Table("questions q").
		Select(`
			q.id AS question_id, q.sort_order, q.prompt,
			q.option_a, q.option_b, q.option_c, q.option_d,
			aa.chosen_option, q.correct_option, COALESCE(aa.is_correct, false) AS is_correct,
			q.points, COALESCE(aa.points_awarded, 0) AS points_awarded
		`).
		Joins("LEFT JOIN attempt_answers aa ON aa.question_id = q.id AND aa.attempt_id = ?", attempt.ID).
		Where("q.exam_id = ?", exam.ID).
		Order("q.sort_order ASC, q.id ASC").
		Scan(&answers).Error
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load submission details", err)
		return
	}

	var maxScore float64
	if err := h.db.Model(&models.Question{}).Where("exam_id = ?", exam.ID).
		Select("COALESCE(SUM(points), 0)").Scan(&maxScore).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load exam score", err)
		return
	}

	var appeals []models.Appeal
	if err := h.db.Where("submission_id = ? AND student_id = ?", attempt.ID, userID).
		Order("created_at DESC, id DESC").Find(&appeals).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load appeals", err)
		return
	}

	apiresponse.JSON(c, http.StatusOK, HistoryDetailResponse{
		ResponseMeta: ResponseMeta{Code: "SUBMISSION_DETAIL_RETRIEVED"},
		SubmissionID: attempt.ID,
		Exam:         ExamSummaryDTO{ID: exam.ID, Title: exam.Title, ExamCode: exam.ExamCode},
		Score:        attempt.Score, MaxScore: maxScore, StartedAt: attempt.StartedAt,
		SubmittedAt: attempt.SubmittedAt, Answers: answers, Appeals: appeals,
	})
}

type CreateAppealInput struct {
	SubmissionID uint   `json:"submissionId" binding:"required"`
	Title        string `json:"title" binding:"max=255"`
	Content      string `json:"content" binding:"required,min=10,max=1000"`
}

// StudentCreateAppeal creates a new appeal for a completed submission.
func (h *StudentHandler) StudentCreateAppeal(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input CreateAppealInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiresponse.Validation(c, err)
		return
	}

	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)
	if utf8.RuneCountInString(input.Content) < 10 || utf8.RuneCountInString(input.Content) > 1000 {
		apiresponse.Error(c, http.StatusBadRequest, "content must contain between 10 and 1000 characters")
		return
	}

	var attempt models.ExamAttempt
	if err := h.db.Where("id = ? AND user_id = ? AND status = 'completed'", input.SubmissionID, userID).
		First(&attempt).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Completed submission not found", "Failed to load submission for appeal")
		return
	}

	var exam models.Exam
	if err := h.db.First(&exam, attempt.ExamID).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Exam not found", "Failed to load exam")
		return
	}

	// Prevent duplicate pending appeals on the same submission
	var pendingCount int64
	if err := h.db.Model(&models.Appeal{}).
		Where("submission_id = ? AND student_id = ? AND status = ?", attempt.ID, userID, models.AppealStatusPending).
		Count(&pendingCount).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to check pending appeals", err)
		return
	}
	if pendingCount > 0 {
		apiresponse.Error(c, http.StatusConflict, "Bạn đã có khiếu nại đang chờ xử lý cho bài thi này.")
		return
	}

	title := input.Title
	if title == "" {
		title = "Khiếu nại điểm thi"
	}

	appeal := models.Appeal{
		ExamID:       exam.ID,
		SubmissionID: attempt.ID,
		StudentID:    attempt.UserID,
		TeacherID:    exam.TeacherID,
		Title:        title,
		Content:      input.Content,
		Status:       models.AppealStatusPending,
	}
	if err := h.db.Create(&appeal).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to create appeal", err)
		return
	}
	apiresponse.JSON(c, http.StatusCreated, AppealCreatedResponse{ResponseMeta: ResponseMeta{Code: "APPEAL_CREATED"}, Message: apiresponse.SuccessMessage(c, "APPEAL_CREATED"), Appeal: appeal})
}

func isAppealStatus(status string) bool {
	switch status {
	case models.AppealStatusPending, models.AppealStatusReplied,
		models.AppealStatusRejected, models.AppealStatusResolved:
		return true
	default:
		return false
	}
}

type TeacherAppealItem struct {
	ID           uint      `json:"id"`
	SubmissionID uint      `json:"submission_id"`
	StudentID    uint      `json:"student_id"`
	StudentName  string    `json:"student_name"`
	StudentEmail string    `json:"student_email"`
	ExamTitle    string    `json:"exam_title"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	Score        float64   `json:"score"`
}

// TeacherListExamAppeals lists all appeals for a specific exam owned by the teacher.
func (h *TeacherHandler) TeacherListExamAppeals(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var exam models.Exam
	if err := h.db.Where("id = ? AND teacher_id = ?", examID, teacherID).First(&exam).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Exam not found", "Failed to load exam")
		return
	}

	paging := parsePagination(c)
	query := h.db.Table("appeals ap").
		Joins("JOIN users u ON u.id = ap.student_id").
		Where("ap.exam_id = ? AND ap.teacher_id = ?", exam.ID, teacherID)
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		if !isAppealStatus(status) {
			apiresponse.Error(c, http.StatusBadRequest, "Invalid status")
			return
		}
		query = query.Where("ap.status = ?", status)
	}
	if search := strings.TrimSpace(c.Query("q")); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("(LOWER(u.full_name) LIKE ? OR LOWER(ap.title) LIKE ? OR LOWER(ap.content) LIKE ?)", like, like, like)
	}
	if err := query.Count(&paging.Total).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load appeals", err)
		return
	}

	items := make([]TeacherAppealItem, 0)
	err := query.Select(`
		ap.id, ap.submission_id, ap.student_id, u.full_name AS student_name,
		u.email AS student_email, e.title AS exam_title, ap.title, ap.content,
		ap.status, ap.created_at, ea.score
	`).
		Joins("JOIN exams e ON e.id = ap.exam_id").
		Joins("JOIN exam_attempts ea ON ea.id = ap.submission_id").
		Order("ap.created_at DESC, ap.id DESC").
		Offset((paging.Page - 1) * paging.PageSize).
		Limit(paging.PageSize).
		Scan(&items).Error
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load appeals", err)
		return
	}
	apiresponse.JSON(c, http.StatusOK, AppealsResponse{ResponseMeta: ResponseMeta{Code: "APPEALS_LISTED"}, Items: items, Pagination: paging})
}

type TeacherAppealAnswer struct {
	QuestionID    uint    `json:"question_id"`
	SortOrder     int     `json:"sort_order"`
	Prompt        string  `json:"prompt"`
	OptionA       string  `json:"option_a"`
	OptionB       string  `json:"option_b"`
	OptionC       string  `json:"option_c"`
	OptionD       string  `json:"option_d"`
	ChosenOption  *string `json:"chosen_option"`
	CorrectOption string  `json:"correct_option"`
	IsCorrect     bool    `json:"is_correct"`
	Points        float64 `json:"points"`
}

// TeacherGetAppeal gets detail of a single appeal.
func (h *TeacherHandler) TeacherGetAppeal(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	appealID, ok := parseUintParam(c, "appealId")
	if !ok {
		return
	}

	var appeal models.Appeal
	if err := h.db.Where(`
		id = ? AND teacher_id = ?
		AND EXISTS (SELECT 1 FROM exams e WHERE e.id = appeals.exam_id AND e.teacher_id = ?)
	`, appealID, teacherID, teacherID).First(&appeal).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Appeal not found", "Failed to load appeal")
		return
	}

	var student StudentSummaryDTO
	if err := h.db.Select("id, full_name, email").First(&student, appeal.StudentID).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load student", err)
		return
	}
	var exam models.Exam
	if err := h.db.First(&exam, appeal.ExamID).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load exam", err)
		return
	}
	var attempt models.ExamAttempt
	if err := h.db.Where("id = ? AND exam_id = ? AND user_id = ?", appeal.SubmissionID, appeal.ExamID, appeal.StudentID).
		First(&attempt).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load submission", err)
		return
	}
	var maxScore float64
	if err := h.db.Model(&models.Question{}).Where("exam_id = ?", exam.ID).
		Select("COALESCE(SUM(points), 0)").Scan(&maxScore).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load exam score", err)
		return
	}
	answers := make([]TeacherAppealAnswer, 0)
	err := h.db.Table("questions q").
		Select(`
			q.id AS question_id, q.sort_order, q.prompt,
			q.option_a, q.option_b, q.option_c, q.option_d,
			aa.chosen_option, q.correct_option, COALESCE(aa.is_correct, false) AS is_correct, q.points
		`).
		Joins("LEFT JOIN attempt_answers aa ON aa.question_id = q.id AND aa.attempt_id = ?", attempt.ID).
		Where("q.exam_id = ?", exam.ID).
		Order("q.sort_order ASC, q.id ASC").
		Scan(&answers).Error
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load submission answers", err)
		return
	}

	apiresponse.JSON(c, http.StatusOK, TeacherAppealDetailResponse{
		ResponseMeta: ResponseMeta{Code: "APPEAL_DETAIL_RETRIEVED"},
		Appeal:       appeal, Student: student, Exam: ExamSummaryDTO{ID: exam.ID, Title: exam.Title},
		Submission: SubmissionSummaryDTO{ID: attempt.ID, Score: attempt.Score, MaxScore: maxScore,
			StartedAt: attempt.StartedAt, SubmittedAt: attempt.SubmittedAt}, Answers: answers,
	})
}

type AppealReplyInput struct {
	Response string `json:"response" binding:"omitempty,max=2000"`
}

// TeacherReplyAppeal allows a teacher to respond to an appeal.
func (h *TeacherHandler) TeacherReplyAppeal(c *gin.Context) {
	var input AppealReplyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiresponse.Validation(c, err)
		return
	}
	response := strings.TrimSpace(input.Response)
	if response == "" || utf8.RuneCountInString(response) > 2000 {
		apiresponse.Error(c, http.StatusBadRequest, "response must contain between 1 and 2000 characters")
		return
	}
	h.updateAppealStatus(c, models.AppealStatusReplied, &response)
}

// TeacherRejectAppeal marks an appeal as rejected, with an optional teacher response note.
func (h *TeacherHandler) TeacherRejectAppeal(c *gin.Context) {
	var input AppealReplyInput
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		apiresponse.Validation(c, err)
		return
	}
	var resp *string
	trimmed := strings.TrimSpace(input.Response)
	if trimmed != "" {
		resp = &trimmed
	}
	h.updateAppealStatus(c, models.AppealStatusRejected, resp)
}

// TeacherResolveAppeal marks an appeal as resolved, with an optional teacher response note.
func (h *TeacherHandler) TeacherResolveAppeal(c *gin.Context) {
	var input AppealReplyInput
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		apiresponse.Validation(c, err)
		return
	}
	var resp *string
	trimmed := strings.TrimSpace(input.Response)
	if trimmed != "" {
		resp = &trimmed
	}
	h.updateAppealStatus(c, models.AppealStatusResolved, resp)
}

func (h *TeacherHandler) updateAppealStatus(c *gin.Context, status string, response *string) {
	teacherID, _ := c.Get("user_id")
	appealID, ok := parseUintParam(c, "appealId")
	if !ok {
		return
	}

	var appeal models.Appeal
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(`
				id = ? AND teacher_id = ?
				AND EXISTS (SELECT 1 FROM exams e WHERE e.id = appeals.exam_id AND e.teacher_id = ?)
			`, appealID, teacherID, teacherID).
			First(&appeal).Error; err != nil {
			return err
		}
		if appeal.Status != models.AppealStatusPending && appeal.Status != models.AppealStatusReplied {
			return errAppealClosed
		}
		now := time.Now()
		updates := map[string]interface{}{
			"status":       status,
			"responded_at": now,
		}
		if response != nil {
			updates["teacher_response"] = response
		}
		return tx.Model(&appeal).Updates(updates).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		apiresponse.Error(c, http.StatusNotFound, "Appeal not found")
		return
	}
	if err != nil {
		if errors.Is(err, errAppealClosed) {
			apiresponse.Error(c, http.StatusConflict, err.Error())
			return
		}
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to update appeal", err)
		return
	}
	if err := h.db.First(&appeal, appeal.ID).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load updated appeal", err)
		return
	}
	apiresponse.JSON(c, http.StatusOK, AppealResponse{ResponseMeta: ResponseMeta{Code: "APPEAL_UPDATED"}, Appeal: appeal})
}
