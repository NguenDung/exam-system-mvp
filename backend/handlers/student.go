package handlers

import (
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"

	"exam-system-mvp/config"
	"exam-system-mvp/models"

	"github.com/gin-gonic/gin"
)

// ─── GET /api/student/exams/lookup?code=XYZ ──────────────────────────────────

func StudentLookupExam(c *gin.Context) {
	userID, _ := c.Get("user_id")
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'code' is required"})
		return
	}

	var exam models.Exam
	if err := config.DB.Where("exam_code = ? AND status = 'published'", code).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found or not published"})
		return
	}

	now := time.Now()

	// Determine exam time status
	var examTimeStatus string
	switch {
	case now.Before(exam.StartTime):
		examTimeStatus = "not_started_yet"
	case now.After(exam.EndTime):
		examTimeStatus = "expired"
	default:
		examTimeStatus = "open"
	}

	// Check student's attempt
	var attempt models.ExamAttempt
	hasAttempt := config.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&attempt).Error == nil

	var attemptStatus string
	if hasAttempt {
		if attempt.Status == "completed" {
			attemptStatus = "submitted"
		} else {
			// in_progress — check if time is up
			elapsed := now.Sub(attempt.StartedAt)
			remaining := time.Duration(exam.Duration)*time.Minute - elapsed
			if remaining <= 0 {
				attemptStatus = "time_up"
			} else {
				attemptStatus = "in_progress"
			}
		}
	} else {
		attemptStatus = "not_attempted"
	}

	response := gin.H{
		"exam":           exam,
		"exam_status":    examTimeStatus,
		"attempt_status": attemptStatus,
	}
	if hasAttempt {
		response["attempt"] = attempt
	}

	c.JSON(http.StatusOK, response)
}

// ─── GET /api/student/attempts ───────────────────────────────────────────────

func StudentListAttempts(c *gin.Context) {
	userID, _ := c.Get("user_id")

	type AttemptWithExam struct {
		models.ExamAttempt
		ExamTitle    string `json:"exam_title"`
		ExamCode     string `json:"exam_code"`
		ExamDuration int    `json:"exam_duration"`
	}

	var attempts []AttemptWithExam
	config.DB.Table("exam_attempts ea").
		Select("ea.*, e.title as exam_title, e.exam_code, e.duration as exam_duration").
		Joins("JOIN exams e ON e.id = ea.exam_id").
		Where("ea.user_id = ?", userID).
		Order("ea.started_at DESC").
		Scan(&attempts)

	c.JSON(http.StatusOK, gin.H{"attempts": attempts})
}

// ─── POST /api/student/exams/:id/start ───────────────────────────────────────

func StudentStartExam(c *gin.Context) {
	userID, _ := c.Get("user_id")
	examID := c.Param("id")

	var exam models.Exam
	if err := config.DB.Where("id = ? AND status = 'published'", examID).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found or not published"})
		return
	}

	now := time.Now()
	if now.Before(exam.StartTime) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Exam has not started yet"})
		return
	}
	if now.After(exam.EndTime) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Exam has expired"})
		return
	}

	// Load questions (without correct_option)
	var questions []models.Question
	config.DB.Where("exam_id = ?", exam.ID).Order("sort_order ASC").Find(&questions)

	safeQuestions := make([]models.QuestionSafe, len(questions))
	for i, q := range questions {
		safeQuestions[i] = models.ToQuestionSafe(q)
	}

	// Find or create attempt
	var attempt models.ExamAttempt
	result := config.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&attempt)

	var remainingSeconds int64

	if result.Error != nil {
		// No existing attempt — create new
		attempt = models.ExamAttempt{
			ExamID:    exam.ID,
			UserID:    userID.(uint),
			Status:    "in_progress",
			StartedAt: now,
		}
		if err := config.DB.Create(&attempt).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create attempt"})
			return
		}
		remainingSeconds = int64(exam.Duration) * 60
	} else {
		// Existing attempt
		if attempt.Status == "completed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already submitted this exam"})
			return
		}

		// Calculate remaining time
		elapsed := now.Sub(attempt.StartedAt)
		remaining := time.Duration(exam.Duration)*time.Minute - elapsed
		if remaining <= 0 {
			// Auto-submit
			autoSubmitAttempt(attempt.ID)
			c.JSON(http.StatusOK, gin.H{
				"message":           "Time is up. Your exam has been automatically submitted.",
				"remaining_seconds": 0,
				"attempt":           attempt,
			})
			return
		}
		remainingSeconds = int64(remaining.Seconds())
	}

	// Load saved answers
	var savedAnswers []struct {
		QuestionID   uint    `json:"question_id"`
		ChosenOption *string `json:"chosen_option"`
	}
	config.DB.Table("attempt_answers").
		Select("question_id, chosen_option").
		Where("attempt_id = ?", attempt.ID).
		Scan(&savedAnswers)

	c.JSON(http.StatusOK, gin.H{
		"attempt":           attempt,
		"remaining_seconds": remainingSeconds,
		"questions":         safeQuestions,
		"saved_answers":     savedAnswers,
	})
}

// ─── POST /api/student/attempts/:attempt_id/save ─────────────────────────────

type SaveAnswerInput struct {
	QuestionID   uint   `json:"question_id" binding:"required"`
	ChosenOption string `json:"chosen_option" binding:"omitempty,oneof=A B C D"`
}

type SaveAnswersInput struct {
	Answers []SaveAnswerInput `json:"answers"`
	// single answer convenience fields
	QuestionID   uint   `json:"question_id"`
	ChosenOption string `json:"chosen_option"`
}

func StudentSaveAnswers(c *gin.Context) {
	userID, _ := c.Get("user_id")
	attemptID, ok := parseUintParam(c, "attempt_id")
	if !ok {
		return
	}

	var attempt models.ExamAttempt
	if err := config.DB.Where("id = ? AND user_id = ?", attemptID, userID).First(&attempt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}
	if attempt.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exam already submitted"})
		return
	}

	var input SaveAnswersInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Support both single-answer and batch
	answers := input.Answers
	if len(answers) == 0 && input.QuestionID != 0 {
		answers = []SaveAnswerInput{{
			QuestionID:   input.QuestionID,
			ChosenOption: input.ChosenOption,
		}}
	}
	if len(answers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No answers provided"})
		return
	}

	now := time.Now()
	for _, ans := range answers {
		chosen := ans.ChosenOption
		var chosenPtr *string
		if chosen != "" {
			chosenPtr = &chosen
		}

		// UPSERT via ON CONFLICT
		config.DB.Exec(`
			INSERT INTO attempt_answers (attempt_id, question_id, chosen_option, is_correct, points_awarded, updated_at)
			VALUES (?, ?, ?, false, 0, ?)
			ON CONFLICT (attempt_id, question_id)
			DO UPDATE SET chosen_option = EXCLUDED.chosen_option, updated_at = EXCLUDED.updated_at
		`, attempt.ID, ans.QuestionID, chosenPtr, now)
	}

	// Update last_saved_at
	config.DB.Model(&attempt).Update("last_saved_at", now)

	c.JSON(http.StatusOK, gin.H{"message": "Answers saved successfully", "saved_at": now})
}

// ─── POST /api/student/attempts/:attempt_id/submit ───────────────────────────

type FinalAnswerInput struct {
	Answers []SaveAnswerInput `json:"answers"`
}

func StudentSubmitExam(c *gin.Context) {
	userID, _ := c.Get("user_id")
	attemptIDStr := c.Param("attempt_id")
	attemptIDUint, err := strconv.ParseUint(attemptIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attempt_id"})
		return
	}
	attemptID := uint(attemptIDUint)

	var attempt models.ExamAttempt
	if err := config.DB.Where("id = ? AND user_id = ?", attemptID, userID).First(&attempt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}
	if attempt.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exam already submitted"})
		return
	}

	// Optional: UPSERT final answers before grading
	var input FinalAnswerInput
	if err := c.ShouldBindJSON(&input); err == nil && len(input.Answers) > 0 {
		now := time.Now()
		for _, ans := range input.Answers {
			chosen := ans.ChosenOption
			var chosenPtr *string
			if chosen != "" {
				chosenPtr = &chosen
			}
			config.DB.Exec(`
				INSERT INTO attempt_answers (attempt_id, question_id, chosen_option, is_correct, points_awarded, updated_at)
				VALUES (?, ?, ?, false, 0, ?)
				ON CONFLICT (attempt_id, question_id)
				DO UPDATE SET chosen_option = EXCLUDED.chosen_option, updated_at = EXCLUDED.updated_at
			`, attempt.ID, ans.QuestionID, chosenPtr, now)
		}
	}

	// Grade in a transaction
	gradeErr := gradeAttemptInTransaction(attemptID)
	if gradeErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to grade exam"})
		return
	}

	// Return updated attempt
	config.DB.First(&attempt, attemptID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Exam submitted and graded successfully",
		"attempt": attempt,
	})
}

// ─── GET /api/student/attempts/:attempt_id/result ────────────────────────────

func StudentGetResult(c *gin.Context) {
	userID, _ := c.Get("user_id")
	attemptIDStr := c.Param("attempt_id")
	attemptIDUint, err := strconv.ParseUint(attemptIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attempt_id"})
		return
	}
	attemptID := uint(attemptIDUint)

	var attempt models.ExamAttempt
	if err := config.DB.Where("id = ? AND user_id = ?", attemptID, userID).First(&attempt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}
	if attempt.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exam not yet submitted"})
		return
	}

	var exam models.Exam
	config.DB.First(&exam, attempt.ExamID)

	now := time.Now()
	showCorrectAnswers := now.After(exam.EndTime)

	// Load answers with question info
	type AnswerDetail struct {
		QuestionID    uint    `json:"question_id"`
		Prompt        string  `json:"prompt"`
		OptionA       string  `json:"option_a"`
		OptionB       string  `json:"option_b"`
		OptionC       string  `json:"option_c"`
		OptionD       string  `json:"option_d"`
		CorrectOption *string `json:"correct_option,omitempty"`
		ChosenOption  *string `json:"chosen_option"`
		IsCorrect     bool    `json:"is_correct"`
		PointsAwarded float64 `json:"points_awarded"`
		Points        float64 `json:"points"`
	}

	var answers []AnswerDetail
	config.DB.Table("attempt_answers aa").
		Select("aa.question_id, q.prompt, q.option_a, q.option_b, q.option_c, q.option_d, q.correct_option, aa.chosen_option, aa.is_correct, aa.points_awarded, q.points").
		Joins("JOIN questions q ON q.id = aa.question_id").
		Where("aa.attempt_id = ?", attemptID).
		Order("q.sort_order ASC").
		Scan(&answers)

	if !showCorrectAnswers {
		for i := range answers {
			answers[i].CorrectOption = nil
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"attempt":              attempt,
		"exam":                 exam,
		"show_correct_answers": showCorrectAnswers,
		"answers":              answers,
	})
}

// ─── Internal: grade attempt in DB transaction ────────────────────────────────

func gradeAttemptInTransaction(attemptID uint) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		// Load all attempt answers with their questions
		var answers []models.AttemptAnswer
		if err := tx.Preload("Question").Where("attempt_id = ?", attemptID).Find(&answers).Error; err != nil {
			return err
		}

		// Load all questions for the attempt's exam (to catch unanswered ones too)
		var attempt models.ExamAttempt
		if err := tx.First(&attempt, attemptID).Error; err != nil {
			return err
		}

		var questions []models.Question
		tx.Where("exam_id = ?", attempt.ExamID).Find(&questions)

		// Build a map of question_id -> question
		qMap := make(map[uint]models.Question, len(questions))
		for _, q := range questions {
			qMap[q.ID] = q
		}

		// Build a map of existing answers
		ansMap := make(map[uint]*models.AttemptAnswer, len(answers))
		for i := range answers {
			ansMap[answers[i].QuestionID] = &answers[i]
		}

		totalScore := 0.0
		correctCount := 0
		wrongCount := 0
		now := time.Now()

		for _, q := range questions {
			ans, exists := ansMap[q.ID]
			if !exists {
				// No answer recorded — unanswered question, create record with 0 points
				newAns := models.AttemptAnswer{
					AttemptID:     attemptID,
					QuestionID:    q.ID,
					ChosenOption:  nil,
					IsCorrect:     false,
					PointsAwarded: 0,
					UpdatedAt:     now,
				}
				tx.Exec(`
					INSERT INTO attempt_answers (attempt_id, question_id, chosen_option, is_correct, points_awarded, updated_at)
					VALUES (?, ?, NULL, false, 0, ?)
					ON CONFLICT (attempt_id, question_id) DO NOTHING
				`, attemptID, q.ID, now)
				_ = newAns
				continue
			}

			isCorrect := false
			pointsAwarded := 0.0
			if ans.ChosenOption != nil && *ans.ChosenOption == q.CorrectOption {
				isCorrect = true
				pointsAwarded = q.Points
				totalScore += pointsAwarded
				correctCount++
			} else if ans.ChosenOption != nil {
				wrongCount++
			}

			if err := tx.Model(ans).Updates(map[string]interface{}{
				"is_correct":     isCorrect,
				"points_awarded": pointsAwarded,
				"updated_at":     now,
			}).Error; err != nil {
				return err
			}
		}

		submittedAt := now
		if err := tx.Model(&models.ExamAttempt{}).Where("id = ?", attemptID).Updates(map[string]interface{}{
			"status":        "completed",
			"score":         totalScore,
			"correct_count": correctCount,
			"wrong_count":   wrongCount,
			"submitted_at":  submittedAt,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

// autoSubmitAttempt is used when time runs out during resume.
func autoSubmitAttempt(attemptID uint) {
	_ = gradeAttemptInTransaction(attemptID)
}
