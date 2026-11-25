package handlers

import (
	"errors"
	"exam-system-mvp/apiresponse"
	"net/http"
	"strconv"
	"strings"
	"time"

	"exam-system-mvp/models"
	"exam-system-mvp/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ─── GET /api/teacher/exams ──────────────────────────────────────────────────

func (h *TeacherHandler) TeacherListExams(c *gin.Context) {
	teacherID, _ := c.Get("user_id")

	var exams []models.Exam
	if err := h.db.Where("teacher_id = ?", teacherID).Order("created_at DESC").Find(&exams).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load exams", err)
		return
	}

	apiresponse.JSON(c, http.StatusOK, ExamListResponse{ResponseMeta: ResponseMeta{Code: "EXAMS_LISTED"}, Exams: exams})
}

// ─── POST /api/teacher/exams ─────────────────────────────────────────────────

type CreateExamInput struct {
	Title     string    `json:"title" binding:"required,min=1,max=200"`
	Duration  int       `json:"duration" binding:"required,min=1,max=360"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
}

func (h *TeacherHandler) TeacherCreateExam(c *gin.Context) {
	teacherID, _ := c.Get("user_id")

	var input CreateExamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiresponse.Validation(c, err)
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		apiresponse.Error(c, http.StatusBadRequest, "Exam title cannot be blank")
		return
	}

	if !input.EndTime.After(input.StartTime) {
		apiresponse.Error(c, http.StatusBadRequest, "end_time must be after start_time")
		return
	}

	// Generate a unique exam code (include soft-deleted rows via Unscoped)
	var examCode string
	for {
		examCode = utils.GenerateExamCode()
		var existing models.Exam
		err := h.db.Unscoped().Where("exam_code = ?", examCode).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			break // code not taken
		}
		if err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "Failed to allocate exam code", err)
			return
		}
	}

	exam := models.Exam{
		TeacherID: teacherID.(uint),
		ExamCode:  examCode,
		Title:     input.Title,
		Duration:  input.Duration,
		StartTime: input.StartTime,
		EndTime:   input.EndTime,
		Status:    "draft",
	}

	if err := h.db.Create(&exam).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to create exam", err)
		return
	}

	apiresponse.JSON(c, http.StatusCreated, ExamResponse{ResponseMeta: ResponseMeta{Code: "EXAM_CREATED"}, Exam: exam})
}

// ─── GET /api/teacher/exams/:id ──────────────────────────────────────────────

func (h *TeacherHandler) TeacherGetExam(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	examID := c.Param("id")

	var exam models.Exam
	if err := h.db.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Where("id = ? AND teacher_id = ?", examID, teacherID).First(&exam).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Exam not found", "Failed to load exam")
		return
	}

	apiresponse.JSON(c, http.StatusOK, ExamResponse{ResponseMeta: ResponseMeta{Code: "EXAM_RETRIEVED"}, Exam: exam})
}

// ─── POST /api/teacher/exams/:id/questions ───────────────────────────────────

type QuestionInput struct {
	SortOrder     int     `json:"sort_order" binding:"gte=0"`
	Prompt        string  `json:"prompt" binding:"required,max=2000"`
	OptionA       string  `json:"option_a" binding:"required,max=500"`
	OptionB       string  `json:"option_b" binding:"required,max=500"`
	OptionC       string  `json:"option_c" binding:"required,max=500"`
	OptionD       string  `json:"option_d" binding:"required,max=500"`
	CorrectOption string  `json:"correct_option" binding:"required,oneof=A B C D"`
	Points        float64 `json:"points" binding:"gte=0,lte=100"`
}

type AddQuestionsInput struct {
	Questions []QuestionInput `json:"questions" binding:"required,min=1,dive"`
}

func (h *TeacherHandler) TeacherAddQuestions(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	examID := c.Param("id")

	var exam models.Exam
	if err := h.db.Where("id = ? AND teacher_id = ?", examID, teacherID).First(&exam).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Exam not found", "Failed to load exam")
		return
	}
	if exam.Status != "draft" {
		apiresponse.Error(c, http.StatusBadRequest, "Cannot add questions to a published exam")
		return
	}

	var input AddQuestionsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiresponse.Validation(c, err)
		return
	}

	// Determine current max sort_order
	var maxOrder int
	if err := h.db.Model(&models.Question{}).Where("exam_id = ?", exam.ID).
		Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to allocate question order", err)
		return
	}

	newQuestions := make([]models.Question, 0, len(input.Questions))
	for i, qi := range input.Questions {
		qi.Prompt = strings.TrimSpace(qi.Prompt)
		qi.OptionA = strings.TrimSpace(qi.OptionA)
		qi.OptionB = strings.TrimSpace(qi.OptionB)
		qi.OptionC = strings.TrimSpace(qi.OptionC)
		qi.OptionD = strings.TrimSpace(qi.OptionD)
		if qi.Prompt == "" || qi.OptionA == "" || qi.OptionB == "" || qi.OptionC == "" || qi.OptionD == "" {
			apiresponse.Error(c, http.StatusBadRequest, "Question prompt and options cannot be blank")
			return
		}
		points := qi.Points
		if points <= 0 {
			points = 1.0
		}
		sortOrder := qi.SortOrder
		if sortOrder == 0 {
			sortOrder = maxOrder + i + 1
		}
		newQuestions = append(newQuestions, models.Question{
			ExamID:        exam.ID,
			SortOrder:     sortOrder,
			Prompt:        qi.Prompt,
			OptionA:       qi.OptionA,
			OptionB:       qi.OptionB,
			OptionC:       qi.OptionC,
			OptionD:       qi.OptionD,
			CorrectOption: qi.CorrectOption,
			Points:        points,
		})
	}

	if err := h.db.Create(&newQuestions).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to save questions", err)
		return
	}

	apiresponse.JSON(c, http.StatusCreated, QuestionsCreatedResponse{ResponseMeta: ResponseMeta{Code: "QUESTIONS_CREATED"}, Message: apiresponse.SuccessMessage(c, "QUESTIONS_CREATED"), Questions: newQuestions})
}

// ─── DELETE /api/teacher/questions/:question_id ──────────────────────────────

func (h *TeacherHandler) TeacherDeleteQuestion(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	questionID, ok := parseUintParam(c, "question_id")
	if !ok {
		return
	}

	var question models.Question
	if err := h.db.First(&question, questionID).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Question not found", "Failed to load question")
		return
	}

	// Strict ownership: question must belong to a draft exam owned by this teacher
	var exam models.Exam
	if err := h.db.Where("id = ? AND teacher_id = ?", question.ExamID, teacherID).First(&exam).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apiresponse.Error(c, http.StatusForbidden, "Access denied")
		} else {
			apiresponse.Error(c, http.StatusInternalServerError, "Failed to verify exam ownership", err)
		}
		return
	}
	if exam.Status != "draft" {
		apiresponse.Error(c, http.StatusBadRequest, "Cannot delete questions from a published exam")
		return
	}

	if err := h.db.Delete(&question).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to delete question", err)
		return
	}
	apiresponse.JSON(c, http.StatusOK, MessageResponse{ResponseMeta: ResponseMeta{Code: "QUESTION_DELETED"}, Message: apiresponse.SuccessMessage(c, "QUESTION_DELETED")})
}

// ─── PATCH /api/teacher/exams/:id/publish ────────────────────────────────────

func (h *TeacherHandler) TeacherPublishExam(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	examID := c.Param("id")

	var exam models.Exam
	if err := h.db.Where("id = ? AND teacher_id = ?", examID, teacherID).First(&exam).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Exam not found", "Failed to load exam")
		return
	}
	if exam.Status == "published" {
		apiresponse.Error(c, http.StatusBadRequest, "Exam is already published")
		return
	}

	var questionCount int64
	if err := h.db.Model(&models.Question{}).Where("exam_id = ?", exam.ID).Count(&questionCount).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to count exam questions", err)
		return
	}
	if questionCount == 0 {
		apiresponse.Error(c, http.StatusBadRequest, "Exam must have at least 1 question before publishing")
		return
	}

	now := time.Now()
	if err := h.db.Model(&exam).Updates(map[string]interface{}{
		"status":       "published",
		"published_at": now,
	}).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to publish exam", err)
		return
	}
	exam.Status = "published"
	exam.PublishedAt = &now
	apiresponse.JSON(c, http.StatusOK, PublishedExamResponse{ResponseMeta: ResponseMeta{Code: "EXAM_PUBLISHED"}, Message: apiresponse.SuccessMessage(c, "EXAM_PUBLISHED"), Exam: exam})
}

// ─── GET /api/teacher/exams/:id/results ──────────────────────────────────────

type StudentResult struct {
	AttemptID    uint       `json:"attempt_id"`
	UserID       uint       `json:"user_id"`
	FullName     string     `json:"full_name"`
	Email        string     `json:"email"`
	Score        float64    `json:"score"`
	CorrectCount int        `json:"correct_count"`
	WrongCount   int        `json:"wrong_count"`
	Status       string     `json:"status"`
	SubmittedAt  *time.Time `json:"submitted_at"`
	StartedAt    time.Time  `json:"started_at"`
}

func (h *TeacherHandler) TeacherGetResults(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	examID := c.Param("id")

	var exam models.Exam
	if err := h.db.Where("id = ? AND teacher_id = ?", examID, teacherID).First(&exam).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Exam not found", "Failed to load exam")
		return
	}

	// Auto-grade any expired in-progress attempts before reporting
	if err := gradeExpiredAttempts(h.db, exam.ID); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to submit expired attempts", err)
		return
	}

	var results []StudentResult
	if err := h.db.Table("exam_attempts ea").
		Select("ea.id as attempt_id, ea.user_id, u.full_name, u.email, ea.score, ea.correct_count, ea.wrong_count, ea.status, ea.submitted_at, ea.started_at").
		Joins("JOIN users u ON u.id = ea.user_id").
		Where("ea.exam_id = ?", exam.ID).
		Order("ea.score DESC").
		Scan(&results).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load exam results", err)
		return
	}

	stats := ExamStatisticsDTO{TotalStudents: len(results)}

	// Only aggregate scores from completed attempts
	completed := make([]StudentResult, 0)
	for _, r := range results {
		if r.Status == "completed" {
			completed = append(completed, r)
		}
	}
	if len(completed) > 0 {
		sum := 0.0
		stats.MaxScore = completed[0].Score
		stats.MinScore = completed[0].Score
		for _, r := range completed {
			sum += r.Score
			if r.Score > stats.MaxScore {
				stats.MaxScore = r.Score
			}
			if r.Score < stats.MinScore {
				stats.MinScore = r.Score
			}
		}
		stats.AvgScore = sum / float64(len(completed))
	}

	apiresponse.JSON(c, http.StatusOK, TeacherResultsResponse{ResponseMeta: ResponseMeta{Code: "EXAM_RESULTS_RETRIEVED"}, Exam: exam, Statistics: stats, Results: results})
}

// ─── helper: parse uint param ─────────────────────────────────────────────────

func parseUintParam(c *gin.Context, param string) (uint, bool) {
	v, err := strconv.ParseUint(c.Param(param), 10, 64)
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "Invalid "+param)
		return 0, false
	}
	return uint(v), true
}
