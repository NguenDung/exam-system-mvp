package service

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
	"time"

	"exam-system-mvp/dto"
	"exam-system-mvp/models"
	"exam-system-mvp/repository"
	"exam-system-mvp/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrExamNotFound      = errors.New("Exam not found or you do not have permission")
	ErrExamNotDraft      = errors.New("Exam is already published and questions cannot be modified")
	ErrInvalidTimeRange  = errors.New("end_time must be after start_time")
	ErrNoQuestionsInExam = errors.New("Exam must have at least 1 question before publishing")
)

type ExamService interface {
	CreateExam(teacherID uint, req dto.CreateExamRequest) (*models.Exam, error)
	GetExam(teacherID, examID uint) (*models.Exam, error)
	ListExams(teacherID uint) ([]models.Exam, error)
	UpdateExam(teacherID, examID uint, req dto.UpdateExamRequest) (*models.Exam, error)
	DeleteExam(teacherID, examID uint) error
	PublishExam(teacherID, examID uint) (*models.Exam, error)

	AddQuestions(teacherID, examID uint, questions []dto.QuestionDTO) ([]models.Question, error)
	UpdateQuestion(teacherID, questionID uint, q dto.QuestionDTO) (*models.Question, error)
	DeleteQuestion(teacherID, questionID uint) error

	GetResults(teacherID, examID uint) (ginH map[string]interface{}, err error)
	ExportResultsCSV(teacherID, examID uint) ([]byte, string, error)

	UpdateAttemptScore(teacherID, attemptID uint, req dto.UpdateScoreRequest) (*models.ExamAttempt, error)
	ResetAttempt(teacherID, attemptID uint) error

	ListStudents(teacherID uint, query string) ([]map[string]interface{}, error)
	CreateStudent(teacherID uint, req dto.CreateStudentRequest) (*models.User, error)
	UpdateStudent(teacherID, studentID uint, req dto.UpdateStudentRequest) (*models.User, error)
	DeleteStudent(teacherID, studentID uint) error

	GetAuditLogs(teacherID uint) ([]models.AuditLog, error)
}

type examService struct {
	examRepo repository.ExamRepository
	userRepo repository.UserRepository
	notifSvc NotificationService
}

func NewExamService(
	examRepo repository.ExamRepository,
	userRepo repository.UserRepository,
	notifSvc NotificationService,
) ExamService {
	return &examService{
		examRepo: examRepo,
		userRepo: userRepo,
		notifSvc: notifSvc,
	}
}

func (s *examService) CreateExam(teacherID uint, req dto.CreateExamRequest) (*models.Exam, error) {
	if !req.EndTime.After(req.StartTime) {
		return nil, ErrInvalidTimeRange
	}

	var examCode string
	for {
		examCode = utils.GenerateExamCode()
		existing, err := s.examRepo.FindExamByCode(examCode)
		if (err != nil && errors.Is(err, gorm.ErrRecordNotFound)) || existing == nil {
			break
		}
	}

	duration := req.Duration
	if duration <= 0 && req.DurationMinutes > 0 {
		duration = req.DurationMinutes
	}

	exam := models.Exam{
		TeacherID:   teacherID,
		ExamCode:    examCode,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Duration:    duration,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Status:      "draft",
		ClassroomID: req.ClassroomID,
	}

	if err := s.examRepo.CreateExam(&exam); err != nil {
		return nil, err
	}
	return &exam, nil
}

func (s *examService) GetExam(teacherID, examID uint) (*models.Exam, error) {
	exam, err := s.examRepo.FindExamByIDWithQuestions(examID)
	if err != nil || exam.TeacherID != teacherID {
		return nil, ErrExamNotFound
	}
	return exam, nil
}

func (s *examService) ListExams(teacherID uint) ([]models.Exam, error) {
	return s.examRepo.ListExamsByTeacher(teacherID)
}

func (s *examService) UpdateExam(teacherID, examID uint, req dto.UpdateExamRequest) (*models.Exam, error) {
	exam, err := s.examRepo.FindExamByID(examID)
	if err != nil || exam.TeacherID != teacherID {
		return nil, ErrExamNotFound
	}

	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
	}
	if req.Duration != nil {
		updates["duration"] = *req.Duration
	}
	if req.DurationMinutes != nil {
		updates["duration"] = *req.DurationMinutes
	}
	if req.StartTime != nil {
		updates["start_time"] = *req.StartTime
	}
	if req.EndTime != nil {
		updates["end_time"] = *req.EndTime
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.ClassroomID != nil {
		updates["classroom_id"] = req.ClassroomID
	}

	if err := s.examRepo.UpdateExam(exam, updates); err != nil {
		return nil, err
	}
	return s.examRepo.FindExamByID(examID)
}

func (s *examService) DeleteExam(teacherID, examID uint) error {
	exam, err := s.examRepo.FindExamByID(examID)
	if err != nil || exam.TeacherID != teacherID {
		return ErrExamNotFound
	}
	return s.examRepo.DeleteExam(examID)
}

func (s *examService) PublishExam(teacherID, examID uint) (*models.Exam, error) {
	exam, err := s.examRepo.FindExamByID(examID)
	if err != nil || exam.TeacherID != teacherID {
		return nil, ErrExamNotFound
	}

	count, _, err := s.examRepo.GetQuestionStats(examID)
	if err != nil || count == 0 {
		return nil, ErrNoQuestionsInExam
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       "published",
		"published_at": &now,
	}
	if err := s.examRepo.UpdateExam(exam, updates); err != nil {
		return nil, err
	}
	return s.examRepo.FindExamByID(examID)
}

func (s *examService) AddQuestions(teacherID, examID uint, questionsDTO []dto.QuestionDTO) ([]models.Question, error) {
	exam, err := s.examRepo.FindExamByID(examID)
	if err != nil || exam.TeacherID != teacherID {
		return nil, ErrExamNotFound
	}
	if exam.Status != "draft" {
		return nil, ErrExamNotDraft
	}

	questions := make([]models.Question, 0, len(questionsDTO))
	for i, q := range questionsDTO {
		pts := q.Points
		if pts <= 0 {
			pts = 1.0
		}
		questions = append(questions, models.Question{
			ExamID:        examID,
			SortOrder:     i + 1,
			Prompt:        q.Prompt,
			OptionA:       q.OptionA,
			OptionB:       q.OptionB,
			OptionC:       q.OptionC,
			OptionD:       q.OptionD,
			CorrectOption: strings.ToUpper(strings.TrimSpace(q.CorrectOption)),
			Points:        pts,
		})
	}

	return s.examRepo.AddQuestions(examID, questions)
}

func (s *examService) UpdateQuestion(teacherID, questionID uint, qDTO dto.QuestionDTO) (*models.Question, error) {
	q, err := s.examRepo.FindQuestionByID(questionID)
	if err != nil {
		return nil, errors.New("Question not found")
	}

	exam, err := s.examRepo.FindExamByID(q.ExamID)
	if err != nil || exam.TeacherID != teacherID {
		return nil, errors.New("Permission denied")
	}
	if exam.Status != "draft" {
		return nil, ErrExamNotDraft
	}

	q.Prompt = qDTO.Prompt
	q.OptionA = qDTO.OptionA
	q.OptionB = qDTO.OptionB
	q.OptionC = qDTO.OptionC
	q.OptionD = qDTO.OptionD
	q.CorrectOption = strings.ToUpper(strings.TrimSpace(qDTO.CorrectOption))
	if qDTO.Points > 0 {
		q.Points = qDTO.Points
	}

	if err := s.examRepo.UpdateQuestion(q); err != nil {
		return nil, err
	}
	return q, nil
}

func (s *examService) DeleteQuestion(teacherID, questionID uint) error {
	q, err := s.examRepo.FindQuestionByID(questionID)
	if err != nil {
		return errors.New("Question not found")
	}

	exam, err := s.examRepo.FindExamByID(q.ExamID)
	if err != nil || exam.TeacherID != teacherID {
		return errors.New("Permission denied")
	}
	if exam.Status != "draft" {
		return ErrExamNotDraft
	}

	return s.examRepo.DeleteQuestion(questionID)
}

func (s *examService) GetResults(teacherID, examID uint) (map[string]interface{}, error) {
	exam, err := s.examRepo.FindExamByID(examID)
	if err != nil || exam.TeacherID != teacherID {
		return nil, ErrExamNotFound
	}

	attempts, err := s.examRepo.ListAttemptsByExamID(examID)
	if err != nil {
		return nil, err
	}

	type Stats struct {
		TotalStudents int     `json:"total_students"`
		AvgScore      float64 `json:"avg_score"`
		MaxScore      float64 `json:"max_score"`
		MinScore      float64 `json:"min_score"`
	}
	stats := Stats{TotalStudents: len(attempts)}

	resultsList := make([]map[string]interface{}, 0, len(attempts))
	var completed []models.ExamAttempt
	for _, a := range attempts {
		if a.Status == "completed" {
			completed = append(completed, a)
		}

		resultsList = append(resultsList, map[string]interface{}{
			"attempt_id":      a.ID,
			"user_id":         a.UserID,
			"student_id":      a.UserID,
			"full_name":       a.User.FullName,
			"student_name":    a.User.FullName,
			"email":           a.User.Email,
			"student_email":   a.User.Email,
			"status":          a.Status,
			"score":           a.Score,
			"teacher_note":    a.TeacherNote,
			"violation_count": a.ViolationCount,
			"correct_count":   a.CorrectCount,
			"wrong_count":     a.WrongCount,
			"started_at":      a.StartedAt,
			"submitted_at":    a.SubmittedAt,
		})
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

	return map[string]interface{}{
		"exam":       exam,
		"statistics": stats,
		"stats": map[string]interface{}{
			"total_attempts": len(attempts),
			"completed":      len(completed),
			"average_score":  stats.AvgScore,
			"highest_score":  stats.MaxScore,
			"lowest_score":   stats.MinScore,
		},
		"results": resultsList,
	}, nil
}

func (s *examService) ExportResultsCSV(teacherID, examID uint) ([]byte, string, error) {
	exam, err := s.examRepo.FindExamByID(examID)
	if err != nil || exam.TeacherID != teacherID {
		return nil, "", ErrExamNotFound
	}

	attempts, err := s.examRepo.ListAttemptsByExamID(examID)
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"Họ tên", "Email", "Điểm số", "Trạng thái", "Số lần vi phạm", "Thời gian nộp"})

	for _, a := range attempts {
		statusStr := "Đang làm"
		if a.Status == "completed" {
			statusStr = "Đã nộp"
		}
		submittedAtStr := ""
		if a.SubmittedAt != nil {
			submittedAtStr = a.SubmittedAt.Format("2006-01-02 15:04:05")
		}
		_ = writer.Write([]string{
			a.User.FullName,
			a.User.Email,
			fmt.Sprintf("%.2f", a.Score),
			statusStr,
			fmt.Sprintf("%d", a.ViolationCount),
			submittedAtStr,
		})
	}
	writer.Flush()

	fileName := fmt.Sprintf("bang-diem-%s-%s.csv", exam.ExamCode, time.Now().Format("20060102-150405"))
	return buf.Bytes(), fileName, nil
}

func (s *examService) UpdateAttemptScore(teacherID, attemptID uint, req dto.UpdateScoreRequest) (*models.ExamAttempt, error) {
	attempt, err := s.examRepo.FindAttemptByID(attemptID)
	if err != nil {
		return nil, errors.New("Attempt not found")
	}

	exam, err := s.examRepo.FindExamByID(attempt.ExamID)
	if err != nil || exam.TeacherID != teacherID {
		return nil, errors.New("Permission denied")
	}

	oldScore := attempt.Score
	attempt.Score = req.Score
	note := req.TeacherNote
	if note == "" && req.Note != "" {
		note = req.Note
	}
	attempt.TeacherNote = note

	if err := s.examRepo.UpdateAttempt(attempt); err != nil {
		return nil, err
	}

	s.examRepo.CreateAuditLog(&models.AuditLog{
		ActorID:   teacherID,
		Action:    "UPDATE_SCORE",
		TargetID:  attempt.ID,
		Details:   fmt.Sprintf("Changed score from %.2f to %.2f. Note: %s", oldScore, req.Score, req.TeacherNote),
		CreatedAt: time.Now(),
	})

	if s.notifSvc != nil {
		s.notifSvc.CreateNotification(
			attempt.UserID,
			"Cập nhật điểm thi",
			fmt.Sprintf("Giảng viên đã hiệu chỉnh điểm thi của bạn trên đề thi \"%s\" thành %.2f điểm.", exam.Title, req.Score),
			fmt.Sprintf("/student/history/%d", attempt.ID),
		)
	}

	return attempt, nil
}

func (s *examService) ResetAttempt(teacherID, attemptID uint) error {
	attempt, err := s.examRepo.FindAttemptByID(attemptID)
	if err != nil {
		return errors.New("Attempt not found")
	}

	exam, err := s.examRepo.FindExamByID(attempt.ExamID)
	if err != nil || exam.TeacherID != teacherID {
		return errors.New("Permission denied")
	}

	if err := s.examRepo.DeleteAttempt(attemptID); err != nil {
		return err
	}

	s.examRepo.CreateAuditLog(&models.AuditLog{
		ActorID:   teacherID,
		Action:    "RESET_ATTEMPT",
		TargetID:  attemptID,
		Details:   fmt.Sprintf("Reset attempt for student ID %d on exam %s", attempt.UserID, exam.ExamCode),
		CreatedAt: time.Now(),
	})

	if s.notifSvc != nil {
		s.notifSvc.CreateNotification(
			attempt.UserID,
			"Được phép thi lại",
			fmt.Sprintf("Lượt thi của bạn trên đề thi \"%s\" đã được giảng viên thiết lập lại. Bạn có thể bắt đầu thi lại từ đầu.", exam.Title),
			"/student",
		)
	}

	return nil
}

func (s *examService) ListStudents(teacherID uint, query string) ([]map[string]interface{}, error) {
	users, err := s.userRepo.FindDirectory(query, models.RoleStudent, "", 200)
	if err != nil {
		return nil, err
	}

	res := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		attempts, _ := s.examRepo.ListAttemptsByUserID(u.ID)
		res = append(res, map[string]interface{}{
			"id":           u.ID,
			"full_name":    u.FullName,
			"email":        u.Email,
			"student_code": u.StudentCode,
			"department":   u.Department,
			"avatar_url":   u.AvatarURL,
			"exams_taken":  len(attempts),
			"created_at":   u.CreatedAt,
		})
	}
	return res, nil
}

func (s *examService) CreateStudent(teacherID uint, req dto.CreateStudentRequest) (*models.User, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	existing, _ := s.userRepo.FindByEmail(email)
	if existing != nil {
		return nil, errors.New("Email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("Failed to hash password")
	}

	user := models.User{
		FullName:     strings.TrimSpace(req.FullName),
		Email:        email,
		PasswordHash: string(hash),
		Role:         models.RoleStudent,
	}

	if err := s.userRepo.Create(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *examService) UpdateStudent(teacherID, studentID uint, req dto.UpdateStudentRequest) (*models.User, error) {
	student, err := s.userRepo.FindByID(studentID)
	if err != nil || student.Role != models.RoleStudent {
		return nil, errors.New("Student not found")
	}

	updates := map[string]interface{}{}
	if req.FullName != "" {
		updates["full_name"] = strings.TrimSpace(req.FullName)
	}
	if req.Email != "" {
		email := strings.ToLower(strings.TrimSpace(req.Email))
		if email != student.Email {
			ex, _ := s.userRepo.FindByEmail(email)
			if ex != nil {
				return nil, errors.New("Email already in use")
			}
			updates["email"] = email
		}
	}
	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("Failed to hash password")
		}
		updates["password_hash"] = string(hash)
	}

	if len(updates) > 0 {
		if err := s.userRepo.UpdateFields(student, updates); err != nil {
			return nil, err
		}
	}

	return s.userRepo.FindByID(studentID)
}

func (s *examService) DeleteStudent(teacherID, studentID uint) error {
	student, err := s.userRepo.FindByID(studentID)
	if err != nil || student.Role != models.RoleStudent {
		return errors.New("Student not found")
	}
	// Soft delete student via repository
	return s.userRepo.UpdateFields(student, map[string]interface{}{"deleted_at": time.Now()})
}

func (s *examService) GetAuditLogs(teacherID uint) ([]models.AuditLog, error) {
	return s.examRepo.ListAuditLogs(100)
}
