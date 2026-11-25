package handlers

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"exam-system-mvp/apiresponse"
	"exam-system-mvp/models"
	"exam-system-mvp/observability"

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
func gradeExpiredAttempts(db *gorm.DB, examID uint) error {
	var exam models.Exam
	if err := db.First(&exam, examID).Error; err != nil {
		return fmt.Errorf("auto-submit failed to load exam %d: %w", examID, err)
	}

	var attempts []models.ExamAttempt
	if err := db.Where("exam_id = ? AND status = ?", examID, "in_progress").Find(&attempts).Error; err != nil {
		return fmt.Errorf("auto-submit failed to load attempts for exam %d: %w", examID, err)
	}

	now := time.Now()
	for _, a := range attempts {
		if !now.Before(attemptDeadline(a.StartedAt, exam)) {
			if err := gradeAttemptInTransaction(db, a.ID); err != nil {
				return fmt.Errorf("auto-submit failed to grade attempt %d: %w", a.ID, err)
			}
		}
	}
	return nil
}

// StartAttemptAutoSubmitWorker grades expired attempts even when a student has
// left the exam page or disconnected.
func StartAttemptAutoSubmitWorker(db *gorm.DB) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := autoSubmitExpiredAttempts(db); err != nil {
			observability.Error("auto-submit worker failed", "error", err)
		}
	}
}

func autoSubmitExpiredAttempts(db *gorm.DB) error {
	var attempts []models.ExamAttempt
	if err := db.Where("status = ?", "in_progress").Find(&attempts).Error; err != nil {
		return fmt.Errorf("failed to load in-progress attempts: %w", err)
	}

	examIDs := make([]uint, 0, len(attempts))
	seenExamIDs := make(map[uint]struct{}, len(attempts))
	for _, attempt := range attempts {
		if _, seen := seenExamIDs[attempt.ExamID]; !seen {
			seenExamIDs[attempt.ExamID] = struct{}{}
			examIDs = append(examIDs, attempt.ExamID)
		}
	}

	var exams []models.Exam
	if len(examIDs) > 0 {
		if err := db.Where("id IN ?", examIDs).Find(&exams).Error; err != nil {
			return fmt.Errorf("failed to load exams for in-progress attempts: %w", err)
		}
	}
	examsByID := make(map[uint]models.Exam, len(exams))
	for _, exam := range exams {
		examsByID[exam.ID] = exam
	}

	var submitErrors []error
	now := time.Now()
	for _, attempt := range attempts {
		exam, exists := examsByID[attempt.ExamID]
		if !exists {
			submitErrors = append(submitErrors, fmt.Errorf("exam %d for attempt %d not found", attempt.ExamID, attempt.ID))
			continue
		}
		if now.Before(attemptDeadline(attempt.StartedAt, exam)) {
			continue
		}
		if err := gradeAttemptInTransaction(db, attempt.ID); err != nil {
			submitErrors = append(submitErrors, fmt.Errorf("failed to grade attempt %d: %w", attempt.ID, err))
		}
	}
	return errors.Join(submitErrors...)
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

// â”€â”€â”€ GET /api/student/exams/lookup?code=XYZ â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

func (h *StudentHandler) StudentLookupExam(c *gin.Context) {
	userID, _ := c.Get("user_id")
	code := strings.ToUpper(strings.TrimSpace(c.Query("code")))
	if code == "" {
		apiresponse.Error(c, http.StatusBadRequest, "Query parameter 'code' is required")
		return
	}

	var exam models.Exam
	if err := h.db.Where("exam_code = ? AND status = 'published'", code).First(&exam).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Exam not found or not published", "Failed to look up exam")
		return
	}

	if err := gradeExpiredAttempts(h.db, exam.ID); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to submit expired attempts", err)
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

	// Check student's attempt (reload after possible auto-grade)
	var attempt models.ExamAttempt
	attemptErr := h.db.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&attempt).Error
	if attemptErr != nil && !errors.Is(attemptErr, gorm.ErrRecordNotFound) {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to look up exam attempt", attemptErr)
		return
	}
	hasAttempt := attemptErr == nil

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

	response := ExamLookupResponse{ResponseMeta: ResponseMeta{Code: "EXAM_LOOKUP_COMPLETED"}, Exam: exam, ExamStatus: examTimeStatus, AttemptStatus: attemptStatus}
	if hasAttempt {
		response.Attempt = &attempt
	}

	apiresponse.JSON(c, http.StatusOK, response)
}

// â”€â”€â”€ GET /api/student/attempts â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

func (h *StudentHandler) StudentListAttempts(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// Grade expired attempts for all exams this student has started
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

	var attempts []AttemptSummaryDTO
	if err := h.db.Table("exam_attempts ea").
		Select("ea.*, e.title as exam_title, e.exam_code, e.duration as exam_duration").
		Joins("JOIN exams e ON e.id = ea.exam_id").
		Where("ea.user_id = ?", userID).
		Order("ea.started_at DESC").
		Scan(&attempts).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load exam attempts", err)
		return
	}

	apiresponse.JSON(c, http.StatusOK, AttemptsResponse{ResponseMeta: ResponseMeta{Code: "ATTEMPTS_LISTED"}, Attempts: attempts})
}

// â”€â”€â”€ POST /api/student/exams/:id/start â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

type StartExamInput struct {
	Confirmed bool `json:"confirmed"`
}

func (h *StudentHandler) StudentStartExam(c *gin.Context) {
	userID, _ := c.Get("user_id")
	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var exam models.Exam
	if err := h.db.Where("id = ? AND status = 'published'", examID).First(&exam).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Exam not found or not published", "Failed to load exam")
		return
	}

	now := time.Now()
	if now.Before(exam.StartTime) {
		apiresponse.Error(c, http.StatusForbidden, "Exam has not started yet")
		return
	}

	// Check for existing attempt BEFORE blocking on exam.EndTime â€”
	// an in-progress attempt past its deadline must still be auto-graded & returned.
	var attempt models.ExamAttempt
	existingErr := h.db.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&attempt).Error
	if existingErr != nil && !errors.Is(existingErr, gorm.ErrRecordNotFound) {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load existing exam attempt", existingErr)
		return
	}
	hasAttempt := existingErr == nil

	if hasAttempt {
		if attempt.Status == "completed" {
			apiresponse.ErrorWithData(c, http.StatusBadRequest, "You have already submitted this exam", apiresponse.ErrorData{Attempt: attempt})
			return
		}

		deadline := attemptDeadline(attempt.StartedAt, exam)
		if !now.Before(deadline) {
			if err := gradeAttemptInTransaction(h.db, attempt.ID); err != nil {
				apiresponse.Error(c, http.StatusInternalServerError, "Failed to automatically submit exam", err)
				return
			}
			if err := h.db.First(&attempt, attempt.ID).Error; err != nil {
				apiresponse.Error(c, http.StatusInternalServerError, "Failed to load submitted exam", err)
				return
			}
			apiresponse.JSON(c, http.StatusOK, AutoSubmittedResponse{
				ResponseMeta:     ResponseMeta{Code: "EXAM_AUTOSUBMITTED"},
				Message:          apiresponse.SuccessMessage(c, "EXAM_AUTOSUBMITTED"),
				RemainingSeconds: 0, Attempt: attempt, AutoSubmitted: true,
			})
			return
		}
	} else {
		// No attempt yet â€” block new starts after exam window ends
		if now.After(exam.EndTime) {
			apiresponse.Error(c, http.StatusForbidden, "Exam has expired")
			return
		}

		var input StartExamInput
		if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
			apiresponse.Validation(c, err)
			return
		}
		if !input.Confirmed {
			apiresponse.Error(c, http.StatusPreconditionRequired, "Confirm participation before starting the exam")
			return
		}
	}

	// Load questions (without correct_option)
	var questions []models.Question
	if err := h.db.Where("exam_id = ?", exam.ID).Order("sort_order ASC").Find(&questions).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load exam questions", err)
		return
	}

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
		if err := h.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&attempt).Error; err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "Failed to create attempt", err)
			return
		}
		// Re-fetch in case another request won the race
		if attempt.ID == 0 {
			if err := h.db.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&attempt).Error; err != nil {
				apiresponse.Error(c, http.StatusInternalServerError, "Failed to create attempt", err)
				return
			}
			if attempt.Status == "completed" {
				apiresponse.ErrorWithData(c, http.StatusBadRequest, "You have already submitted this exam", apiresponse.ErrorData{Attempt: attempt})
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
	var savedAnswers []SavedAnswerDTO
	if err := h.db.Table("attempt_answers").
		Select("question_id, chosen_option").
		Where("attempt_id = ?", attempt.ID).
		Scan(&savedAnswers).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load saved answers", err)
		return
	}

	apiresponse.JSON(c, http.StatusOK, StartExamResponse{
		ResponseMeta: ResponseMeta{Code: "EXAM_STARTED"},
		Attempt:      attempt, RemainingSeconds: remainingSeconds, Questions: safeQuestions, SavedAnswers: savedAnswers,
	})
}

// â”€â”€â”€ POST /api/student/attempts/:attempt_id/save â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

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

func (h *StudentHandler) StudentSaveAnswers(c *gin.Context) {
	userID, _ := c.Get("user_id")
	attemptID, ok := parseUintParam(c, "attempt_id")
	if !ok {
		return
	}

	var input SaveAnswersInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiresponse.Validation(c, err)
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
		apiresponse.Error(c, http.StatusBadRequest, "No answers provided")
		return
	}

	var expiredAutoGraded bool
	var gradedAttempt models.ExamAttempt

	err := h.db.Transaction(func(tx *gorm.DB) error {
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
			// Past grace period â€” grade with previously saved answers only; reject new saves
			if err := gradeAttemptInTx(tx, attempt.ID); err != nil {
				return err
			}
			expiredAutoGraded = true
			if err := tx.First(&gradedAttempt, attempt.ID).Error; err != nil {
				return err
			}
			return errTimeExpired
		}

		if err := upsertAnswers(tx, attempt, answers); err != nil {
			return err
		}

		return tx.Model(&attempt).Update("last_saved_at", now).Error
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		apiresponse.Error(c, http.StatusNotFound, "Attempt not found")
		return
	}
	if errors.Is(err, errAlreadySubmitted) {
		apiresponse.Error(c, http.StatusBadRequest, "Exam already submitted")
		return
	}
	if errors.Is(err, errTimeExpired) || expiredAutoGraded {
		apiresponse.ErrorWithData(c, http.StatusConflict, "Time is up. Your exam has been automatically submitted.", apiresponse.ErrorData{AutoSubmitted: true, Attempt: gradedAttempt})
		return
	}
	if err != nil {
		if strings.Contains(err.Error(), "does not belong") {
			apiresponse.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to save answers", err)
		return
	}

	apiresponse.JSON(c, http.StatusOK, SaveAnswersResponse{ResponseMeta: ResponseMeta{Code: "ANSWERS_SAVED"}, Message: apiresponse.SuccessMessage(c, "ANSWERS_SAVED"), SavedAt: time.Now()})
}

// â”€â”€â”€ POST /api/student/attempts/:attempt_id/submit â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

type FinalAnswerInput struct {
	Answers []SaveAnswerInput `json:"answers" binding:"omitempty,dive"`
}

func (h *StudentHandler) StudentSubmitExam(c *gin.Context) {
	userID, _ := c.Get("user_id")
	attemptID, ok := parseUintParam(c, "attempt_id")
	if !ok {
		return
	}

	var input FinalAnswerInput
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		apiresponse.Validation(c, err)
		return
	}

	var finalAttempt models.ExamAttempt

	err := h.db.Transaction(func(tx *gorm.DB) error {
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
		apiresponse.Error(c, http.StatusNotFound, "Attempt not found")
		return
	}
	if errors.Is(err, errAlreadySubmitted) {
		apiresponse.Error(c, http.StatusBadRequest, "Exam already submitted")
		return
	}
	if err != nil {
		if strings.Contains(err.Error(), "does not belong") {
			apiresponse.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to grade exam", err)
		return
	}

	apiresponse.JSON(c, http.StatusOK, SubmitExamResponse{ResponseMeta: ResponseMeta{Code: "EXAM_SUBMITTED"}, Message: apiresponse.SuccessMessage(c, "EXAM_SUBMITTED"), Attempt: finalAttempt})
}

// â”€â”€â”€ GET /api/student/attempts/:attempt_id/result â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

func (h *StudentHandler) StudentGetResult(c *gin.Context) {
	userID, _ := c.Get("user_id")
	attemptID, ok := parseUintParam(c, "attempt_id")
	if !ok {
		return
	}

	var attempt models.ExamAttempt
	if err := h.db.Where("id = ? AND user_id = ?", attemptID, userID).First(&attempt).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Attempt not found", "Failed to load exam attempt")
		return
	}
	if attempt.Status != "completed" {
		apiresponse.Error(c, http.StatusBadRequest, "Exam not yet submitted")
		return
	}

	var exam models.Exam
	if err := h.db.First(&exam, attempt.ExamID).Error; err != nil {
		lookupFailure(c, err, http.StatusNotFound, "Exam not found", "Failed to load exam result")
		return
	}

	now := time.Now()
	showCorrectAnswers := now.After(exam.EndTime)

	var answers []AnswerDetailDTO
	if err := h.db.Table("attempt_answers aa").
		Select("aa.question_id, q.prompt, q.option_a, q.option_b, q.option_c, q.option_d, q.correct_option, aa.chosen_option, aa.is_correct, aa.points_awarded, q.points").
		Joins("JOIN questions q ON q.id = aa.question_id").
		Where("aa.attempt_id = ?", attemptID).
		Order("q.sort_order ASC").
		Scan(&answers).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to load exam answers", err)
		return
	}

	if !showCorrectAnswers {
		for i := range answers {
			answers[i].CorrectOption = nil
		}
	}

	apiresponse.JSON(c, http.StatusOK, ExamResultResponse{
		ResponseMeta: ResponseMeta{Code: "EXAM_RESULT_RETRIEVED"},
		Attempt:      attempt, Exam: exam, ShowCorrectAnswers: showCorrectAnswers, Answers: answers,
	})
}

// â”€â”€â”€ Internal grading helpers â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

var (
	errAlreadySubmitted = errors.New("already submitted")
	errTimeExpired      = errors.New("time expired")
)

// gradeAttemptInTransaction wraps gradeAttemptInTx in its own DB transaction.
func gradeAttemptInTransaction(db *gorm.DB, attemptID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
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
			// Unanswered â€” create blank record and count as wrong
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
