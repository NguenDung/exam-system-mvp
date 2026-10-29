package handlers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"exam-system-mvp/config"
	"exam-system-mvp/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const saveGracePeriod = 5 * time.Second

// attemptDeadline returns the earlier of (startedAt + duration) and exam.EndTime.
func attemptDeadline(startedAt time.Time, exam models.Exam) time.Time {
	durationDeadline := startedAt.Add(time.Duration(exam.Duration) * time.Minute)
	if durationDeadline.Before(exam.EndTime) {
		return durationDeadline
	}
	return exam.EndTime
}

// gradeExpiredAttempts auto-grades all in_progress attempts for an exam that are past deadline.
func gradeExpiredAttempts(examID uint) {
	var exam models.Exam
	if err := config.DB.First(&exam, examID).Error; err != nil {
		return
	}

	var attempts []models.ExamAttempt
	config.DB.Where("exam_id = ? AND status = ?", examID, "in_progress").Find(&attempts)

	now := time.Now()
	for _, a := range attempts {
		if now.After(attemptDeadline(a.StartedAt, exam)) {
			_ = gradeAttemptInTransaction(a.ID)
		}
	}
}

// upsertAnswers validates question ownership and upserts answers inside a transaction.
func upsertAnswers(tx *gorm.DB, attempt models.ExamAttempt, answers []SaveAnswerInput) error {
	if len(answers) == 0 {
		return nil
	}

	var validIDs []uint
	if err := tx.Model(&models.Question{}).
		Where("exam_id = ?", attempt.ExamID).
		Pluck("id", &validIDs).Error; err != nil {
		return err
	}
	validSet := make(map[uint]struct{}, len(validIDs))
	for _, id := range validIDs {
		validSet[id] = struct{}{}
	}

	now := time.Now()
	for _, ans := range answers {
		if _, ok := validSet[ans.QuestionID]; !ok {
			return fmt.Errorf("question_id %d does not belong to this exam", ans.QuestionID)
		}

		chosen := ans.ChosenOption
		var chosenPtr *string
		if chosen != "" {
			chosenPtr = &chosen
		}

		if err := tx.Exec(`
			INSERT INTO attempt_answers (attempt_id, question_id, chosen_option, is_correct, points_awarded, updated_at)
			VALUES (?, ?, ?, false, 0, ?)
			ON CONFLICT (attempt_id, question_id)
			DO UPDATE SET chosen_option = EXCLUDED.chosen_option, updated_at = EXCLUDED.updated_at
		`, attempt.ID, ans.QuestionID, chosenPtr, now).Error; err != nil {
			return err
		}
	}
	return nil
}

// ─── GET /api/student/exams/lookup?code=XYZ ──────────────────────────────────

func StudentLookupExam(c *gin.Context) {
	userID, _ := c.Get("user_id")
	code := strings.ToUpper(strings.TrimSpace(c.Query("code")))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'code' is required"})
		return
	}

	var exam models.Exam
	if err := config.DB.Where("exam_code = ? AND status = 'published'", code).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found or not published"})
		return
	}

	gradeExpiredAttempts(exam.ID)

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

	// Check student's attempt (reload after possible auto-grade)
	var attempt models.ExamAttempt
	hasAttempt := config.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&attempt).Error == nil

	var attemptStatus string
	if hasAttempt {
		if attempt.Status == "completed" {
			attemptStatus = "submitted"
		} else {
			deadline := attemptDeadline(attempt.StartedAt, exam)
			if !now.Before(deadline) {
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

	// Grade expired attempts for all exams this student has started
	var examIDs []uint
	config.DB.Model(&models.ExamAttempt{}).
		Where("user_id = ? AND status = ?", userID, "in_progress").
		Distinct("exam_id").
		Pluck("exam_id", &examIDs)
	for _, eid := range examIDs {
		gradeExpiredAttempts(eid)
	}

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
	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

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

	// Check for existing attempt BEFORE blocking on exam.EndTime —
	// an in-progress attempt past its deadline must still be auto-graded & returned.
	var attempt models.ExamAttempt
	existingErr := config.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&attempt).Error
	hasAttempt := existingErr == nil

	if hasAttempt {
		if attempt.Status == "completed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already submitted this exam", "attempt": attempt})
			return
		}

		deadline := attemptDeadline(attempt.StartedAt, exam)
		if !now.Before(deadline) {
			_ = gradeAttemptInTransaction(attempt.ID)
			config.DB.First(&attempt, attempt.ID)
			c.JSON(http.StatusOK, gin.H{
				"message":           "Time is up. Your exam has been automatically submitted.",
				"remaining_seconds": 0,
				"attempt":           attempt,
				"auto_submitted":    true,
			})
			return
		}
	} else {
		// No attempt yet — block new starts after exam window ends
		if now.After(exam.EndTime) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Exam has expired"})
			return
		}
	}

	// Load questions (without correct_option)
	var questions []models.Question
	config.DB.Where("exam_id = ?", exam.ID).Order("sort_order ASC").Find(&questions)

	safeQuestions := make([]models.QuestionSafe, len(questions))
	for i, q := range questions {
		safeQuestions[i] = models.ToQuestionSafe(q)
	}

	var remainingSeconds int64

	if !hasAttempt {
		attempt = models.ExamAttempt{
			ExamID:    exam.ID,
			UserID:    userID.(uint),
			Status:    "in_progress",
			StartedAt: now,
		}
		// OnConflict DoNothing guards against concurrent duplicate creates
		if err := config.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&attempt).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create attempt"})
			return
		}
		// Re-fetch in case another request won the race
		if attempt.ID == 0 {
			if err := config.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&attempt).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create attempt"})
				return
			}
			if attempt.Status == "completed" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "You have already submitted this exam", "attempt": attempt})
				return
			}
		}
		deadline := attemptDeadline(attempt.StartedAt, exam)
		remainingSeconds = int64(deadline.Sub(now).Seconds())
		if remainingSeconds < 0 {
			remainingSeconds = 0
		}
	} else {
		deadline := attemptDeadline(attempt.StartedAt, exam)
		remaining := deadline.Sub(now)
		remainingSeconds = int64(remaining.Seconds())
		if remainingSeconds < 0 {
			remainingSeconds = 0
		}
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
	Answers []SaveAnswerInput `json:"answers" binding:"omitempty,dive"`
	// single answer convenience fields
	QuestionID   uint   `json:"question_id"`
	ChosenOption string `json:"chosen_option" binding:"omitempty,oneof=A B C D"`
}

func StudentSaveAnswers(c *gin.Context) {
	userID, _ := c.Get("user_id")
	attemptID, ok := parseUintParam(c, "attempt_id")
	if !ok {
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

	var expiredAutoGraded bool
	var gradedAttempt models.ExamAttempt

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var attempt models.ExamAttempt
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", attemptID, userID).
			First(&attempt).Error; err != nil {
			return err
		}
		if attempt.Status != "in_progress" {
			return errAlreadySubmitted
		}

		var exam models.Exam
		if err := tx.First(&exam, attempt.ExamID).Error; err != nil {
			return err
		}

		now := time.Now()
		deadline := attemptDeadline(attempt.StartedAt, exam)
		if now.After(deadline.Add(saveGracePeriod)) {
			// Past grace period — grade with previously saved answers only; reject new saves
			if err := gradeAttemptInTx(tx, attempt.ID); err != nil {
				return err
			}
			expiredAutoGraded = true
			tx.First(&gradedAttempt, attempt.ID)
			return errTimeExpired
		}

		if err := upsertAnswers(tx, attempt, answers); err != nil {
			return err
		}

		return tx.Model(&attempt).Update("last_saved_at", now).Error
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}
	if errors.Is(err, errAlreadySubmitted) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exam already submitted"})
		return
	}
	if errors.Is(err, errTimeExpired) {
		c.JSON(http.StatusConflict, gin.H{
			"error":          "Time is up. Your exam has been automatically submitted.",
			"auto_submitted": true,
			"attempt":        gradedAttempt,
		})
		return
	}
	if err != nil {
		// Invalid question_id (IDOR / cross-exam) → 400
		if strings.Contains(err.Error(), "does not belong") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save answers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Answers saved successfully", "saved_at": time.Now()})
	_ = expiredAutoGraded
}

// ─── POST /api/student/attempts/:attempt_id/submit ───────────────────────────

type FinalAnswerInput struct {
	Answers []SaveAnswerInput `json:"answers" binding:"omitempty,dive"`
}

func StudentSubmitExam(c *gin.Context) {
	userID, _ := c.Get("user_id")
	attemptID, ok := parseUintParam(c, "attempt_id")
	if !ok {
		return
	}

	var input FinalAnswerInput
	_ = c.ShouldBindJSON(&input) // optional body

	var finalAttempt models.ExamAttempt

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var attempt models.ExamAttempt
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", attemptID, userID).
			First(&attempt).Error; err != nil {
			return err
		}
		if attempt.Status != "in_progress" {
			return errAlreadySubmitted
		}

		var exam models.Exam
		if err := tx.First(&exam, attempt.ExamID).Error; err != nil {
			return err
		}

		now := time.Now()
		deadline := attemptDeadline(attempt.StartedAt, exam)
		pastGrace := now.After(deadline.Add(saveGracePeriod))

		// If still within grace (or before deadline), accept body answers.
		// If past grace, ignore body answers and grade only previously saved ones.
		if !pastGrace && len(input.Answers) > 0 {
			if err := upsertAnswers(tx, attempt, input.Answers); err != nil {
				return err
			}
		}

		if err := gradeAttemptInTx(tx, attempt.ID); err != nil {
			return err
		}
		return tx.First(&finalAttempt, attempt.ID).Error
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}
	if errors.Is(err, errAlreadySubmitted) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exam already submitted"})
		return
	}
	if err != nil {
		if strings.Contains(err.Error(), "does not belong") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to grade exam"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Exam submitted and graded successfully",
		"attempt": finalAttempt,
	})
}

// ─── GET /api/student/attempts/:attempt_id/result ────────────────────────────

func StudentGetResult(c *gin.Context) {
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
	if attempt.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exam not yet submitted"})
		return
	}

	var exam models.Exam
	config.DB.First(&exam, attempt.ExamID)

	now := time.Now()
	showCorrectAnswers := now.After(exam.EndTime)

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

// ─── Internal grading helpers ────────────────────────────────────────────────

var (
	errAlreadySubmitted = errors.New("already submitted")
	errTimeExpired      = errors.New("time expired")
)

// gradeAttemptInTransaction wraps gradeAttemptInTx in its own DB transaction.
func gradeAttemptInTransaction(attemptID uint) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		return gradeAttemptInTx(tx, attemptID)
	})
}

// gradeAttemptInTx grades an attempt using an existing transaction (supports nested use).
func gradeAttemptInTx(tx *gorm.DB, attemptID uint) error {
	var attempt models.ExamAttempt
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&attempt, attemptID).Error; err != nil {
		return err
	}
	if attempt.Status == "completed" {
		return nil // idempotent
	}

	var answers []models.AttemptAnswer
	if err := tx.Where("attempt_id = ?", attemptID).Find(&answers).Error; err != nil {
		return err
	}

	var questions []models.Question
	if err := tx.Where("exam_id = ?", attempt.ExamID).Find(&questions).Error; err != nil {
		return err
	}

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
			// Unanswered — create blank record and count as wrong
			if err := tx.Exec(`
				INSERT INTO attempt_answers (attempt_id, question_id, chosen_option, is_correct, points_awarded, updated_at)
				VALUES (?, ?, NULL, false, 0, ?)
				ON CONFLICT (attempt_id, question_id) DO NOTHING
			`, attemptID, q.ID, now).Error; err != nil {
				return err
			}
			wrongCount++
			continue
		}

		isCorrect := false
		pointsAwarded := 0.0
		if ans.ChosenOption != nil && *ans.ChosenOption == q.CorrectOption {
			isCorrect = true
			pointsAwarded = q.Points
			totalScore += pointsAwarded
			correctCount++
		} else {
			// Wrong answer OR blank (nil chosen_option) both count as wrong
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

	totalScore = math.Round(totalScore*100) / 100
	submittedAt := now
	return tx.Model(&models.ExamAttempt{}).Where("id = ?", attemptID).Updates(map[string]interface{}{
		"status":        "completed",
		"score":         totalScore,
		"correct_count": correctCount,
		"wrong_count":   wrongCount,
		"submitted_at":  submittedAt,
	}).Error
}
