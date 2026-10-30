package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"exam-system-mvp/config"
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

func StudentGetHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var examIDs []uint
	config.DB.Model(&models.ExamAttempt{}).
		Where("user_id = ? AND status = ?", userID, "in_progress").
		Distinct("exam_id").
		Pluck("exam_id", &examIDs)
	for _, eid := range examIDs {
		gradeExpiredAttempts(eid)
	}

	paging := parsePagination(c)
	query := config.DB.Table("exam_attempts ea").
		Joins("JOIN exams e ON e.id = ea.exam_id").
		Where("ea.user_id = ? AND ea.status = ?", userID, "completed")

	if search := strings.TrimSpace(c.Query("q")); search != "" {
		query = query.Where("LOWER(e.title) LIKE ?", "%"+strings.ToLower(search)+"%")
	}
	if status := strings.TrimSpace(c.Query("appeal_status")); status != "" {
		if !isAppealStatus(status) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appeal_status"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load exam history"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load exam history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "pagination": paging})
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

func StudentGetHistoryDetail(c *gin.Context) {
	userID, _ := c.Get("user_id")
	submissionID, ok := parseUintParam(c, "submissionId")
	if !ok {
		return
	}

	var attempt models.ExamAttempt
	if err := config.DB.Where("id = ? AND user_id = ? AND status = 'completed'", submissionID, userID).
		First(&attempt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	var exam models.Exam
	if err := config.DB.First(&exam, attempt.ExamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	answers := make([]HistoryAnswer, 0)
	err := config.DB.Table("questions q").
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load submission details"})
		return
	}

	var maxScore float64
	if err := config.DB.Model(&models.Question{}).Where("exam_id = ?", exam.ID).
		Select("COALESCE(SUM(points), 0)").Scan(&maxScore).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load exam score"})
		return
	}

	var appeals []models.Appeal
	if err := config.DB.Where("submission_id = ? AND student_id = ?", attempt.ID, userID).
		Order("created_at DESC, id DESC").Find(&appeals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load appeals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"submission_id": attempt.ID,
		"exam": gin.H{
			"id": exam.ID, "title": exam.Title, "exam_code": exam.ExamCode,
		},
		"score": attempt.Score, "max_score": maxScore,
		"started_at": attempt.StartedAt, "submitted_at": attempt.SubmittedAt,
		"answers": answers, "appeals": appeals,
	})
}

type CreateAppealInput struct {
	SubmissionID uint   `json:"submissionId" binding:"required"`
	Title        string `json:"title" binding:"max=255"`
	Content      string `json:"content" binding:"required"`
}

func StudentCreateAppeal(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input CreateAppealInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)
	if utf8.RuneCountInString(input.Content) < 10 || utf8.RuneCountInString(input.Content) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content must contain between 10 and 1000 characters"})
		return
	}

	var attempt models.ExamAttempt
	if err := config.DB.Where("id = ? AND user_id = ? AND status = 'completed'", input.SubmissionID, userID).
		First(&attempt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Completed submission not found"})
		return
	}

	var exam models.Exam
	if err := config.DB.First(&exam, attempt.ExamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	var pendingCount int64
	config.DB.Model(&models.Appeal{}).
		Where("submission_id = ? AND student_id = ? AND status = ?", attempt.ID, userID, models.AppealStatusPending).
		Count(&pendingCount)
	if pendingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Bạn đã có khiếu nại đang chờ xử lý cho bài thi này."})
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
	if err := config.DB.Create(&appeal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create appeal"})
		return
	}
	CreateNotification(exam.TeacherID, "Khiếu nại / Phúc khảo mới từ sinh viên", "Sinh viên đã gửi đơn phúc khảo cho bài thi "+exam.Title+": "+appeal.Title, fmt.Sprintf("/teacher/exams/%d", exam.ID))
	c.JSON(http.StatusCreated, gin.H{"message": "Gửi khiếu nại thành công.", "appeal": appeal})
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

func TeacherListExamAppeals(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var exam models.Exam
	if err := config.DB.Where("id = ? AND teacher_id = ?", examID, teacherID).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	paging := parsePagination(c)
	query := config.DB.Table("appeals ap").
		Joins("JOIN users u ON u.id = ap.student_id").
		Where("ap.exam_id = ? AND ap.teacher_id = ?", exam.ID, teacherID)
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		if !isAppealStatus(status) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
			return
		}
		query = query.Where("ap.status = ?", status)
	}
	if search := strings.TrimSpace(c.Query("q")); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("(LOWER(u.full_name) LIKE ? OR LOWER(ap.title) LIKE ? OR LOWER(ap.content) LIKE ?)", like, like, like)
	}
	if err := query.Count(&paging.Total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load appeals"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load appeals"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "pagination": paging})
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

func TeacherGetAppeal(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	appealID, ok := parseUintParam(c, "appealId")
	if !ok {
		return
	}

	var appeal models.Appeal
	if err := config.DB.Where(`
		id = ? AND teacher_id = ?
		AND EXISTS (SELECT 1 FROM exams e WHERE e.id = appeals.exam_id AND e.teacher_id = ?)
	`, appealID, teacherID, teacherID).First(&appeal).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appeal not found"})
		return
	}

	var student models.User
	if err := config.DB.Select("id, full_name, email").First(&student, appeal.StudentID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load student"})
		return
	}
	var exam models.Exam
	if err := config.DB.First(&exam, appeal.ExamID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load exam"})
		return
	}
	var attempt models.ExamAttempt
	if err := config.DB.Where("id = ? AND exam_id = ? AND user_id = ?", appeal.SubmissionID, appeal.ExamID, appeal.StudentID).
		First(&attempt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load submission"})
		return
	}
	var maxScore float64
	if err := config.DB.Model(&models.Question{}).Where("exam_id = ?", exam.ID).
		Select("COALESCE(SUM(points), 0)").Scan(&maxScore).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load exam score"})
		return
	}
	answers := make([]TeacherAppealAnswer, 0)
	err := config.DB.Table("questions q").
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load submission answers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"appeal": appeal, "student": student,
		"exam": gin.H{"id": exam.ID, "title": exam.Title},
		"submission": gin.H{
			"id": attempt.ID, "score": attempt.Score, "max_score": maxScore,
			"started_at": attempt.StartedAt, "submitted_at": attempt.SubmittedAt,
		},
		"answers": answers,
	})
}

type AppealReplyInput struct {
	Response string `json:"response"`
}

func TeacherReplyAppeal(c *gin.Context) {
	var input AppealReplyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response := strings.TrimSpace(input.Response)
	if response == "" || utf8.RuneCountInString(response) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "response must contain between 1 and 2000 characters"})
		return
	}
	updateAppealStatus(c, models.AppealStatusReplied, &response)
}

func TeacherRejectAppeal(c *gin.Context) {
	var input AppealReplyInput
	_ = c.ShouldBindJSON(&input)
	var resp *string
	trimmed := strings.TrimSpace(input.Response)
	if trimmed != "" {
		resp = &trimmed
	}
	updateAppealStatus(c, models.AppealStatusRejected, resp)
}

func TeacherResolveAppeal(c *gin.Context) {
	var input AppealReplyInput
	_ = c.ShouldBindJSON(&input)
	var resp *string
	trimmed := strings.TrimSpace(input.Response)
	if trimmed != "" {
		resp = &trimmed
	}
	updateAppealStatus(c, models.AppealStatusResolved, resp)
}

func updateAppealStatus(c *gin.Context, status string, response *string) {
	teacherID, _ := c.Get("user_id")
	appealID, ok := parseUintParam(c, "appealId")
	if !ok {
		return
	}

	var appeal models.Appeal
	err := config.DB.Transaction(func(tx *gorm.DB) error {
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Appeal not found"})
		return
	}
	if err != nil {
		if errors.Is(err, errAppealClosed) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update appeal"})
		return
	}
	if err := config.DB.First(&appeal, appeal.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated appeal"})
		return
	}

	var statusMsg string
	switch status {
	case models.AppealStatusReplied:
		statusMsg = "Giảng viên đã phản hồi đơn phúc khảo"
	case models.AppealStatusResolved:
		statusMsg = "Đơn phúc khảo của bạn đã được chấp thuận"
	case models.AppealStatusRejected:
		statusMsg = "Đơn phúc khảo của bạn đã bị từ chối"
	default:
		statusMsg = "Đơn phúc khảo của bạn có cập nhật mới"
	}
	CreateNotification(appeal.StudentID, "Cập nhật đơn phúc khảo", statusMsg+": "+appeal.Title, fmt.Sprintf("/student/history/%d", appeal.SubmissionID))

	c.JSON(http.StatusOK, gin.H{"appeal": appeal})
}
