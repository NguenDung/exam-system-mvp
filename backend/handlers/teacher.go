package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"exam-system-mvp/config"
	"exam-system-mvp/models"
	"exam-system-mvp/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ─── GET /api/teacher/exams ──────────────────────────────────────────────────

func TeacherListExams(c *gin.Context) {
	teacherID, _ := c.Get("user_id")

	var exams []models.Exam
	config.DB.Where("teacher_id = ?", teacherID).Order("created_at DESC").Find(&exams)

	c.JSON(http.StatusOK, gin.H{"exams": exams})
}

// ─── POST /api/teacher/exams ─────────────────────────────────────────────────

type CreateExamInput struct {
	Title     string    `json:"title" binding:"required"`
	Duration  int       `json:"duration" binding:"required,min=1"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
}

func TeacherCreateExam(c *gin.Context) {
	teacherID, _ := c.Get("user_id")

	var input CreateExamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !input.EndTime.After(input.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_time must be after start_time"})
		return
	}

	// Generate a unique exam code (include soft-deleted rows via Unscoped)
	var examCode string
	for {
		examCode = utils.GenerateExamCode()
		var existing models.Exam
		err := config.DB.Unscoped().Where("exam_code = ?", examCode).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			break // code not taken
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to allocate exam code"})
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

	if err := config.DB.Create(&exam).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exam"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"exam": exam})
}

// ─── GET /api/teacher/exams/:id ──────────────────────────────────────────────

func TeacherGetExam(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	examID := c.Param("id")

	var exam models.Exam
	if err := config.DB.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Where("id = ? AND teacher_id = ?", examID, teacherID).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exam": exam})
}

// ─── POST /api/teacher/exams/:id/questions ───────────────────────────────────

type QuestionInput struct {
	SortOrder     int     `json:"sort_order"`
	Prompt        string  `json:"prompt" binding:"required"`
	OptionA       string  `json:"option_a" binding:"required"`
	OptionB       string  `json:"option_b" binding:"required"`
	OptionC       string  `json:"option_c" binding:"required"`
	OptionD       string  `json:"option_d" binding:"required"`
	CorrectOption string  `json:"correct_option" binding:"required,oneof=A B C D"`
	Points        float64 `json:"points"`
}

type AddQuestionsInput struct {
	Questions []QuestionInput `json:"questions" binding:"required,min=1,dive"`
}

func TeacherAddQuestions(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	examID := c.Param("id")

	var exam models.Exam
	if err := config.DB.Where("id = ? AND teacher_id = ?", examID, teacherID).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}
	if exam.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot add questions to a published exam"})
		return
	}

	var input AddQuestionsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine current max sort_order
	var maxOrder int
	config.DB.Model(&models.Question{}).Where("exam_id = ?", exam.ID).
		Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)

	newQuestions := make([]models.Question, 0, len(input.Questions))
	for i, qi := range input.Questions {
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

	if err := config.DB.Create(&newQuestions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save questions"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Questions added successfully",
		"questions": newQuestions,
	})
}

// ─── DELETE /api/teacher/questions/:question_id ──────────────────────────────

func TeacherDeleteQuestion(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	questionID, ok := parseUintParam(c, "question_id")
	if !ok {
		return
	}

	var question models.Question
	if err := config.DB.First(&question, questionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	// Strict ownership: question must belong to a draft exam owned by this teacher
	var exam models.Exam
	if err := config.DB.Where("id = ? AND teacher_id = ?", question.ExamID, teacherID).First(&exam).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}
	if exam.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete questions from a published exam"})
		return
	}

	config.DB.Delete(&question)
	c.JSON(http.StatusOK, gin.H{"message": "Question deleted successfully"})
}

// ─── PATCH /api/teacher/exams/:id/publish ────────────────────────────────────

func TeacherPublishExam(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	examID := c.Param("id")

	var exam models.Exam
	if err := config.DB.Where("id = ? AND teacher_id = ?", examID, teacherID).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}
	if exam.Status == "published" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exam is already published"})
		return
	}

	var questionCount int64
	config.DB.Model(&models.Question{}).Where("exam_id = ?", exam.ID).Count(&questionCount)
	if questionCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exam must have at least 1 question before publishing"})
		return
	}

	now := time.Now()
	if err := config.DB.Model(&exam).Updates(map[string]interface{}{
		"status":       "published",
		"published_at": now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish exam"})
		return
	}
	exam.Status = "published"
	exam.PublishedAt = &now
	c.JSON(http.StatusOK, gin.H{"message": "Exam published successfully", "exam": exam})
}

// ─── GET /api/teacher/exams/:id/results ──────────────────────────────────────

type StudentResult struct {
	AttemptID      uint       `json:"attempt_id"`
	UserID         uint       `json:"user_id"`
	FullName       string     `json:"full_name"`
	Email          string     `json:"email"`
	Score          float64    `json:"score"`
	TeacherNote    string     `json:"teacher_note"`
	ViolationCount int        `json:"violation_count"`
	CorrectCount   int        `json:"correct_count"`
	WrongCount     int        `json:"wrong_count"`
	Status         string     `json:"status"`
	SubmittedAt    *time.Time `json:"submitted_at"`
	StartedAt      time.Time  `json:"started_at"`
}

func TeacherGetResults(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	examID := c.Param("id")

	var exam models.Exam
	if err := config.DB.Where("id = ? AND teacher_id = ?", examID, teacherID).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	// Auto-grade any expired in-progress attempts before reporting
	gradeExpiredAttempts(exam.ID)

	var results []StudentResult
	config.DB.Table("exam_attempts ea").
		Select("ea.id as attempt_id, ea.user_id, u.full_name, u.email, ea.score, ea.teacher_note, ea.violation_count, ea.correct_count, ea.wrong_count, ea.status, ea.submitted_at, ea.started_at").
		Joins("JOIN users u ON u.id = ea.user_id").
		Where("ea.exam_id = ?", exam.ID).
		Order("ea.score DESC").
		Scan(&results)

	type Stats struct {
		TotalStudents int     `json:"total_students"`
		AvgScore      float64 `json:"avg_score"`
		MaxScore      float64 `json:"max_score"`
		MinScore      float64 `json:"min_score"`
	}
	stats := Stats{TotalStudents: len(results)}

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

	c.JSON(http.StatusOK, gin.H{
		"exam":       exam,
		"statistics": stats,
		"results":    results,
	})
}

// ─── helper: parse uint param ─────────────────────────────────────────────────

func parseUintParam(c *gin.Context, param string) (uint, bool) {
	v, err := strconv.ParseUint(c.Param(param), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid " + param})
		return 0, false
	}
	return uint(v), true
}
