package service

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"exam-system-mvp/dto"
	"exam-system-mvp/models"
	"exam-system-mvp/repository"
)

var ErrParticipationNotConfirmed = errors.New("Confirm participation before starting the exam")

const saveGracePeriod = 5 * time.Second

type ExamTakingService interface {
	GetAvailableExams(studentID uint) ([]map[string]interface{}, error)
	LookupExam(studentID uint, code string) (map[string]interface{}, error)
	StartExam(studentID, examID uint, confirmed bool) (*dto.StartExamResponse, error)
	SaveAnswers(studentID, attemptID uint, req dto.SaveAnswersRequest) error
	RecordViolation(studentID, attemptID uint, req dto.ViolationRequest) (map[string]interface{}, error)
	SubmitExam(studentID, attemptID uint, req dto.SubmitExamRequest) (*models.ExamAttempt, error)
	GetResult(studentID, attemptID uint) (map[string]interface{}, error)
	ListAttempts(studentID uint) ([]map[string]interface{}, error)
	AutoSubmitExpiredAttempts() error
	GetExamSchedule(userID uint, role string) (map[string]interface{}, error)
}

type examTakingService struct {
	examRepo  repository.ExamRepository
	classRepo repository.ClassroomRepository
	userRepo  repository.UserRepository
}

func NewExamTakingService(
	examRepo repository.ExamRepository,
	classRepo repository.ClassroomRepository,
	userRepo repository.UserRepository,
) ExamTakingService {
	return &examTakingService{
		examRepo:  examRepo,
		classRepo: classRepo,
		userRepo:  userRepo,
	}
}

func attemptDeadline(startedAt time.Time, exam models.Exam) time.Time {
	durationDeadline := startedAt.Add(time.Duration(exam.Duration) * time.Minute)
	if durationDeadline.Before(exam.EndTime) {
		return durationDeadline
	}
	return exam.EndTime
}

func (s *examTakingService) gradeAttemptInTransaction(attemptID uint) error {
	attempt, err := s.examRepo.FindAttemptByID(attemptID)
	if err != nil {
		return err
	}
	if attempt.Status == "completed" {
		return nil
	}

	questions, err := s.examRepo.GetQuestionsByExamID(attempt.ExamID)
	if err != nil {
		return err
	}

	answers, err := s.examRepo.GetAttemptAnswers(attemptID)
	if err != nil {
		return err
	}

	ansMap := make(map[uint]models.AttemptAnswer, len(answers))
	for _, a := range answers {
		ansMap[a.QuestionID] = a
	}

	var totalScore float64
	var correctCount, wrongCount int

	var updatedAnswers []models.AttemptAnswer
	now := time.Now()

	for _, q := range questions {
		userAns, hasAns := ansMap[q.ID]
		isCorrect := false
		var pts float64 = 0

		if hasAns && userAns.ChosenOption != nil {
			if strings.EqualFold(strings.TrimSpace(*userAns.ChosenOption), strings.TrimSpace(q.CorrectOption)) {
				isCorrect = true
				pts = q.Points
				correctCount++
			} else {
				wrongCount++
			}
			userAns.IsCorrect = isCorrect
			userAns.PointsAwarded = pts
			userAns.UpdatedAt = now
			updatedAnswers = append(updatedAnswers, userAns)
		} else {
			wrongCount++
			updatedAnswers = append(updatedAnswers, models.AttemptAnswer{
				AttemptID:     attemptID,
				QuestionID:    q.ID,
				ChosenOption:  nil,
				IsCorrect:     false,
				PointsAwarded: 0,
				UpdatedAt:     now,
			})
		}
		totalScore += pts
	}

	attempt.Score = totalScore
	attempt.CorrectCount = correctCount
	attempt.WrongCount = wrongCount
	attempt.Status = "completed"
	attempt.SubmittedAt = &now

	return s.examRepo.GradeAttemptAtomic(attemptID, updatedAnswers, attempt)
}

func (s *examTakingService) GetAvailableExams(studentID uint) ([]map[string]interface{}, error) {
	// Auto submit expired attempts for student
	inProgress, _ := s.examRepo.FindAllInProgressAttempts()
	for _, a := range inProgress {
		if a.UserID == studentID {
			exam, err := s.examRepo.FindExamByID(a.ExamID)
			if err == nil {
				if !time.Now().Before(attemptDeadline(a.StartedAt, *exam)) {
					_ = s.gradeAttemptInTransaction(a.ID)
				}
			}
		}
	}

	enrolledClasses, _ := s.classRepo.FindEnrolledByStudentID(studentID)
	var classIDs []uint
	for _, c := range enrolledClasses {
		classIDs = append(classIDs, c.ID)
	}

	exams, err := s.examRepo.GetPublishedExams(classIDs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	items := make([]map[string]interface{}, 0, len(exams))

	for _, exam := range exams {
		var examTimeStatus string
		switch {
		case now.Before(exam.StartTime):
			examTimeStatus = "not_started_yet"
		case now.After(exam.EndTime):
			examTimeStatus = "expired"
		default:
			examTimeStatus = "open"
		}

		attempt, err := s.examRepo.FindAttemptByExamAndUser(exam.ID, studentID)
		hasAttempt := (err == nil && attempt != nil)

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

		qCount, qPoints, _ := s.examRepo.GetQuestionStats(exam.ID)

		item := map[string]interface{}{
			"id":             exam.ID,
			"exam_code":      exam.ExamCode,
			"title":          exam.Title,
			"description":    exam.Description,
			"duration":       exam.Duration,
			"start_time":     exam.StartTime,
			"end_time":       exam.EndTime,
			"question_count": qCount,
			"total_points":   qPoints,
			"exam_status":    examTimeStatus,
			"attempt_status": attemptStatus,
		}
		if hasAttempt {
			item["attempt"] = attempt
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *examTakingService) LookupExam(studentID uint, code string) (map[string]interface{}, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, errors.New("Query parameter 'code' is required")
	}
	exam, err := s.examRepo.FindExamByCode(code)
	if err != nil || exam.Status != "published" {
		return nil, errors.New("Exam not found or not published")
	}

	now := time.Now()
	var examTimeStatus string
	switch {
	case now.Before(exam.StartTime):
		examTimeStatus = "not_started_yet"
	case now.After(exam.EndTime):
		examTimeStatus = "expired"
	default:
		examTimeStatus = "open"
	}

	attempt, err := s.examRepo.FindAttemptByExamAndUser(exam.ID, studentID)
	hasAttempt := (err == nil && attempt != nil)

	var attemptStatus string
	if hasAttempt {
		if attempt.Status == "completed" {
			attemptStatus = "submitted"
		} else {
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

	res := map[string]interface{}{
		"exam":           exam,
		"exam_status":    examTimeStatus,
		"attempt_status": attemptStatus,
	}
	if hasAttempt {
		res["attempt"] = attempt
	}
	return res, nil
}

func (s *examTakingService) StartExam(studentID, examID uint, confirmed bool) (*dto.StartExamResponse, error) {
	exam, err := s.examRepo.FindExamByID(examID)
	if err != nil {
		return nil, errors.New("Exam not found")
	}
	if exam.Status != "published" {
		return nil, errors.New("Exam is not published")
	}

	now := time.Now()
	if now.Before(exam.StartTime) {
		return nil, errors.New("Exam has not started yet")
	}
	if now.After(exam.EndTime) {
		return nil, errors.New("Exam has already ended")
	}

	// Verify classroom enrollment if restricted
	if exam.ClassroomID != nil {
		member, err := s.classRepo.FindMember(*exam.ClassroomID, studentID)
		if err != nil || member == nil {
			return nil, errors.New("You are not enrolled in the classroom for this exam")
		}
	}

	attempt, err := s.examRepo.FindAttemptByExamAndUser(examID, studentID)
	hasAttempt := (err == nil && attempt != nil)
	if hasAttempt {
		deadline := attemptDeadline(attempt.StartedAt, *exam)
		if !now.Before(deadline) {
			if attempt.Status != "completed" {
				if err := s.gradeAttemptInTransaction(attempt.ID); err != nil {
					return nil, errors.New("Failed to automatically submit exam")
				}
				attempt, _ = s.examRepo.FindAttemptByID(attempt.ID)
			}
			return &dto.StartExamResponse{
				Attempt:          *attempt,
				RemainingSeconds: 0,
				AutoSubmitted:    true,
				Message:          "Time is up. Your exam has been automatically submitted.",
			}, nil
		}
		if attempt.Status == "completed" {
			return nil, errors.New("You have already submitted this exam")
		}
	} else {
		if !confirmed {
			return nil, ErrParticipationNotConfirmed
		}
		newAttempt := models.ExamAttempt{
			ExamID:         examID,
			UserID:         studentID,
			Status:         "in_progress",
			StartedAt:      now,
			ViolationCount: 0,
		}
		if err := s.examRepo.CreateAttempt(&newAttempt); err != nil {
			return nil, errors.New("Failed to start exam")
		}
		attempt = &newAttempt
	}

	deadline := attemptDeadline(attempt.StartedAt, *exam)
	if !now.Before(deadline) {
		if err := s.gradeAttemptInTransaction(attempt.ID); err != nil {
			return nil, errors.New("Failed to automatically submit exam")
		}
		updatedAttempt, _ := s.examRepo.FindAttemptByID(attempt.ID)
		return &dto.StartExamResponse{
			Attempt:          *updatedAttempt,
			RemainingSeconds: 0,
			AutoSubmitted:    true,
			Message:          "Time is up. Your exam has been automatically submitted.",
		}, nil
	}

	remainingSeconds := int64(deadline.Sub(now).Seconds())
	if remainingSeconds < 0 {
		remainingSeconds = 0
	}

	questions, err := s.examRepo.GetQuestionsByExamID(examID)
	if err != nil {
		return nil, err
	}

	// Deterministic shuffling with attempt.ID as random seed
	r := rand.New(rand.NewSource(int64(attempt.ID)))
	r.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})

	safeQuestions := make([]models.QuestionSafe, 0, len(questions))
	for _, q := range questions {
		safeQuestions = append(safeQuestions, models.ToQuestionSafe(q))
	}

	savedAnswers, _ := s.examRepo.GetAttemptAnswers(attempt.ID)

	return &dto.StartExamResponse{
		Attempt:          *attempt,
		Questions:        safeQuestions,
		Duration:         exam.Duration,
		StartTime:        exam.StartTime,
		EndTime:          exam.EndTime,
		RemainingSeconds: remainingSeconds,
		SavedAnswers:     savedAnswers,
	}, nil
}

func (s *examTakingService) SaveAnswers(studentID, attemptID uint, req dto.SaveAnswersRequest) error {
	attempt, err := s.examRepo.FindAttemptByID(attemptID)
	if err != nil || attempt.UserID != studentID {
		return errors.New("Attempt not found")
	}
	if attempt.Status != "in_progress" {
		return errors.New("Attempt is already submitted")
	}

	exam, err := s.examRepo.FindExamByID(attempt.ExamID)
	if err != nil {
		return errors.New("Exam not found")
	}

	now := time.Now()
	deadline := attemptDeadline(attempt.StartedAt, *exam)
	if now.After(deadline.Add(saveGracePeriod)) {
		_ = s.gradeAttemptInTransaction(attemptID)
		return errors.New("Submission window closed")
	}

	answersInput := req.Answers
	if len(answersInput) == 0 && req.QuestionID != 0 {
		answersInput = []dto.AnswerInputDTO{{
			QuestionID:   req.QuestionID,
			ChosenOption: req.ChosenOption,
		}}
	}
	if len(answersInput) == 0 {
		return errors.New("No answers provided")
	}

	// Verify all question IDs belong to this exam
	questions, err := s.examRepo.GetQuestionsByExamID(attempt.ExamID)
	if err != nil {
		return err
	}
	validMap := make(map[uint]bool, len(questions))
	for _, q := range questions {
		validMap[q.ID] = true
	}
	for _, a := range answersInput {
		if !validMap[a.QuestionID] {
			return errors.New("Question does not belong to this exam")
		}
		if a.ChosenOption != nil && *a.ChosenOption != "" {
			opt := strings.ToUpper(strings.TrimSpace(*a.ChosenOption))
			if opt != "A" && opt != "B" && opt != "C" && opt != "D" {
				return errors.New("Invalid chosen option: must be A, B, C, or D")
			}
		}
	}

	var answers []models.AttemptAnswer
	for _, a := range answersInput {
		answers = append(answers, models.AttemptAnswer{
			AttemptID:    attemptID,
			QuestionID:   a.QuestionID,
			ChosenOption: a.ChosenOption,
		})
	}

	if err := s.examRepo.SaveAnswersAtomic(attemptID, answers); err != nil {
		return err
	}

	attempt.LastSavedAt = &now
	return s.examRepo.UpdateAttempt(attempt)
}

func (s *examTakingService) RecordViolation(studentID, attemptID uint, req dto.ViolationRequest) (map[string]interface{}, error) {
	attempt, err := s.examRepo.FindAttemptByID(attemptID)
	if err != nil || attempt.UserID != studentID {
		return nil, errors.New("Attempt not found")
	}
	if attempt.Status != "in_progress" {
		return nil, errors.New("Exam already submitted")
	}

	attempt.ViolationCount++
	autoSubmitted := false

	if attempt.ViolationCount >= 3 {
		_ = s.examRepo.UpdateAttempt(attempt)
		_ = s.gradeAttemptInTransaction(attemptID)
		autoSubmitted = true
		attempt, _ = s.examRepo.FindAttemptByID(attemptID)

		_ = s.examRepo.CreateAuditLog(&models.AuditLog{
			ActorID:   studentID,
			Action:    "CHEATING_VIOLATION_AUTO_SUBMIT",
			TargetID:  attemptID,
			Details:   fmt.Sprintf("Tự động thu bài attempt #%d do vi phạm quy chế phòng thi %d lần", attemptID, attempt.ViolationCount),
			CreatedAt: time.Now(),
		})
	} else {
		_ = s.examRepo.UpdateAttempt(attempt)
	}

	return map[string]interface{}{
		"message":         "Violation recorded",
		"violation_count": attempt.ViolationCount,
		"auto_submitted":  autoSubmitted,
		"attempt":         *attempt,
		"max_violations":  3,
	}, nil
}

func (s *examTakingService) SubmitExam(studentID, attemptID uint, req dto.SubmitExamRequest) (*models.ExamAttempt, error) {
	attempt, err := s.examRepo.FindAttemptByID(attemptID)
	if err != nil || attempt.UserID != studentID {
		return nil, errors.New("Attempt not found")
	}
	if attempt.Status == "completed" {
		return nil, errors.New("Exam already submitted")
	}

	exam, err := s.examRepo.FindExamByID(attempt.ExamID)
	if err != nil {
		return nil, errors.New("Exam not found")
	}

	now := time.Now()
	deadline := attemptDeadline(attempt.StartedAt, *exam)
	isExpired := now.After(deadline.Add(saveGracePeriod))

	// Chống gian lận: Chỉ chấp nhận cập nhật đáp án mới nếu chưa quá hạn nộp bài
	if len(req.Answers) > 0 && !isExpired {
		questions, err := s.examRepo.GetQuestionsByExamID(attempt.ExamID)
		if err == nil {
			validMap := make(map[uint]bool, len(questions))
			for _, q := range questions {
				validMap[q.ID] = true
			}
			var validAnswers []models.AttemptAnswer
			for _, a := range req.Answers {
				if validMap[a.QuestionID] {
					if a.ChosenOption != nil && *a.ChosenOption != "" {
						opt := strings.ToUpper(strings.TrimSpace(*a.ChosenOption))
						if opt == "A" || opt == "B" || opt == "C" || opt == "D" {
							validAnswers = append(validAnswers, models.AttemptAnswer{
								AttemptID:    attemptID,
								QuestionID:   a.QuestionID,
								ChosenOption: &opt,
							})
						}
					} else {
						validAnswers = append(validAnswers, models.AttemptAnswer{
							AttemptID:    attemptID,
							QuestionID:   a.QuestionID,
							ChosenOption: nil,
						})
					}
				}
			}
			if len(validAnswers) > 0 {
				_ = s.examRepo.SaveAnswersAtomic(attemptID, validAnswers)
			}
		}
	}

	if err := s.gradeAttemptInTransaction(attemptID); err != nil {
		return nil, err
	}

	updated, err := s.examRepo.FindAttemptByID(attemptID)
	if err == nil && updated != nil {
		_ = s.examRepo.CreateAuditLog(&models.AuditLog{
			ActorID:   studentID,
			Action:    "STUDENT_SUBMIT_EXAM",
			TargetID:  attemptID,
			Details:   fmt.Sprintf("Sinh viên nộp bài thi %s (Điểm: %.2f, Đúng: %d, Sai: %d)", exam.ExamCode, updated.Score, updated.CorrectCount, updated.WrongCount),
			CreatedAt: now,
		})
	}

	return updated, nil
}

func (s *examTakingService) GetResult(studentID, attemptID uint) (map[string]interface{}, error) {
	attempt, err := s.examRepo.FindAttemptByID(attemptID)
	if err != nil || attempt.UserID != studentID {
		return nil, errors.New("Attempt not found")
	}
	if attempt.Status != "completed" {
		return nil, errors.New("Exam not yet submitted")
	}

	exam, err := s.examRepo.FindExamByID(attempt.ExamID)
	if err != nil {
		return nil, errors.New("Exam not found")
	}

	now := time.Now()
	showCorrectAnswers := now.After(exam.EndTime)

	questions, err := s.examRepo.GetQuestionsByExamID(attempt.ExamID)
	if err != nil {
		return nil, err
	}

	savedAnswers, err := s.examRepo.GetAttemptAnswers(attemptID)
	if err != nil {
		return nil, err
	}
	ansMap := make(map[uint]models.AttemptAnswer, len(savedAnswers))
	for _, a := range savedAnswers {
		ansMap[a.QuestionID] = a
	}

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
	for _, q := range questions {
		ans, has := ansMap[q.ID]
		detail := AnswerDetail{
			QuestionID: q.ID,
			Prompt:     q.Prompt,
			OptionA:    q.OptionA,
			OptionB:    q.OptionB,
			OptionC:    q.OptionC,
			OptionD:    q.OptionD,
			Points:     q.Points,
		}
		if showCorrectAnswers {
			cOpt := q.CorrectOption
			detail.CorrectOption = &cOpt
		}
		if has {
			detail.ChosenOption = ans.ChosenOption
			detail.IsCorrect = ans.IsCorrect
			detail.PointsAwarded = ans.PointsAwarded
		}
		answers = append(answers, detail)
	}

	return map[string]interface{}{
		"attempt":              attempt,
		"exam":                 exam,
		"show_correct_answers": showCorrectAnswers,
		"answers":              answers,
	}, nil
}

func (s *examTakingService) ListAttempts(studentID uint) ([]map[string]interface{}, error) {
	attempts, err := s.examRepo.ListAttemptsByUserID(studentID)
	if err != nil {
		return nil, err
	}
	res := make([]map[string]interface{}, 0, len(attempts))
	for _, a := range attempts {
		item := map[string]interface{}{
			"id":              a.ID,
			"exam_id":         a.ExamID,
			"user_id":         a.UserID,
			"status":          a.Status,
			"score":           a.Score,
			"teacher_note":    a.TeacherNote,
			"violation_count": a.ViolationCount,
			"correct_count":   a.CorrectCount,
			"wrong_count":     a.WrongCount,
			"started_at":      a.StartedAt,
			"last_saved_at":   a.LastSavedAt,
			"submitted_at":    a.SubmittedAt,
			"exam_title":      a.Exam.Title,
			"exam_code":       a.Exam.ExamCode,
			"exam_duration":   a.Exam.Duration,
		}
		res = append(res, item)
	}
	return res, nil
}

func (s *examTakingService) AutoSubmitExpiredAttempts() error {
	attempts, err := s.examRepo.FindAllInProgressAttempts()
	if err != nil {
		return err
	}
	now := time.Now()
	for _, a := range attempts {
		exam, err := s.examRepo.FindExamByID(a.ExamID)
		if err == nil {
			if !now.Before(attemptDeadline(a.StartedAt, *exam)) {
				_ = s.gradeAttemptInTransaction(a.ID)
			}
		}
	}
	return nil
}

func (s *examTakingService) GetExamSchedule(userID uint, role string) (map[string]interface{}, error) {
	now := time.Now()
	var exams []models.Exam
	var err error

	if role == models.RoleTeacher {
		exams, err = s.examRepo.ListExamsByTeacher(userID)
	} else {
		// Student: check enrolled classes
		enrolledClasses, _ := s.classRepo.FindEnrolledByStudentID(userID)
		var classIDs []uint
		for _, c := range enrolledClasses {
			classIDs = append(classIDs, c.ID)
		}
		exams, err = s.examRepo.GetPublishedExams(classIDs)
	}
	if err != nil {
		return nil, err
	}

	ongoing := make([]map[string]interface{}, 0)
	upcoming := make([]map[string]interface{}, 0)
	past := make([]map[string]interface{}, 0)

	classrooms, _ := s.examRepo.GetAllClassrooms()
	classMap := make(map[uint]models.Classroom)
	for _, c := range classrooms {
		classMap[c.ID] = c
	}

	for _, exam := range exams {
		var timeStatus string
		var countdownSeconds int64

		if now.Before(exam.StartTime) {
			timeStatus = "upcoming"
			countdownSeconds = int64(exam.StartTime.Sub(now).Seconds())
			if countdownSeconds < 0 {
				countdownSeconds = 0
			}
		} else if now.After(exam.EndTime) {
			timeStatus = "past"
			countdownSeconds = 0
		} else {
			timeStatus = "ongoing"
			countdownSeconds = int64(exam.EndTime.Sub(now).Seconds())
			if countdownSeconds < 0 {
				countdownSeconds = 0
			}
		}

		qCount, qPoints, _ := s.examRepo.GetQuestionStats(exam.ID)

		className := "Toàn trường"
		classCode := ""
		if exam.ClassroomID != nil {
			if cls, ok := classMap[*exam.ClassroomID]; ok {
				className = cls.Name
				classCode = cls.Code
			}
		}

		item := map[string]interface{}{
			"id":                exam.ID,
			"exam_code":         exam.ExamCode,
			"title":             exam.Title,
			"description":       exam.Description,
			"duration":          exam.Duration,
			"start_time":        exam.StartTime,
			"end_time":          exam.EndTime,
			"status":            exam.Status,
			"classroom_id":      exam.ClassroomID,
			"classroom_name":    className,
			"classroom_code":    classCode,
			"teacher_id":        exam.TeacherID,
			"teacher_name":      "Giảng viên",
			"question_count":    qCount,
			"total_points":      qPoints,
			"countdown_seconds": countdownSeconds,
			"time_status":       timeStatus,
		}

		if role == models.RoleTeacher {
			attemptsCount, _ := s.examRepo.CountAttemptsByExamID(exam.ID)
			item["total_attempts"] = attemptsCount
		} else {
			attempt, err := s.examRepo.FindAttemptByExamAndUser(exam.ID, userID)
			if err == nil && attempt != nil {
				item["attempt_id"] = attempt.ID
				if attempt.Status == "completed" {
					item["attempt_status"] = "submitted"
					item["score"] = attempt.Score
				} else {
					deadline := attemptDeadline(attempt.StartedAt, exam)
					if !now.Before(deadline) {
						item["attempt_status"] = "time_up"
					} else {
						item["attempt_status"] = "in_progress"
					}
				}
			} else {
				item["attempt_status"] = "not_attempted"
			}
		}

		switch timeStatus {
		case "ongoing":
			ongoing = append(ongoing, item)
		case "upcoming":
			upcoming = append(upcoming, item)
		case "past":
			past = append(past, item)
		}
	}

	return map[string]interface{}{
		"role":     role,
		"ongoing":  ongoing,
		"upcoming": upcoming,
		"past":     past,
		"summary": map[string]interface{}{
			"ongoing_count":  len(ongoing),
			"upcoming_count": len(upcoming),
			"past_count":     len(past),
			"total":          len(ongoing) + len(upcoming) + len(past),
		},
	}, nil
}
