package service

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"exam-system-mvp/dto"
	"exam-system-mvp/models"
	"exam-system-mvp/repository"
)

var (
	ErrAppealClosed = errors.New("appeal is already closed")
)

type AppealService interface {
	GetHistory(studentID uint, search, status string, page, pageSize int) ([]map[string]interface{}, int64, error)
	GetHistoryDetail(submissionID, studentID uint) (map[string]interface{}, error)
	CreateAppeal(studentID uint, req dto.CreateAppealRequest) (*models.Appeal, error)

	ListExamAppeals(examID, teacherID uint, status, query string, page, pageSize int) ([]dto.TeacherAppealItem, int64, error)
	GetAppeal(appealID, teacherID uint) (*models.Appeal, error)
	GetAppealDetail(appealID, teacherID uint) (map[string]interface{}, error)
	ReplyAppeal(appealID, teacherID uint, req dto.ReplyAppealRequest) (*models.Appeal, error)
	RejectAppeal(appealID, teacherID uint, req dto.ActionAppealRequest) (*models.Appeal, error)
	ResolveAppeal(appealID, teacherID uint, req dto.ActionAppealRequest) (*models.Appeal, error)
}

type appealService struct {
	appealRepo repository.AppealRepository
	examRepo   repository.ExamRepository
	notifSvc   NotificationService
}

func NewAppealService(
	appealRepo repository.AppealRepository,
	examRepo repository.ExamRepository,
	notifSvc NotificationService,
) AppealService {
	return &appealService{
		appealRepo: appealRepo,
		examRepo:   examRepo,
		notifSvc:   notifSvc,
	}
}

func (s *appealService) GetHistory(studentID uint, search, status string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	return s.appealRepo.GetStudentHistory(studentID, search, status, page, pageSize)
}

func (s *appealService) GetHistoryDetail(submissionID, studentID uint) (map[string]interface{}, error) {
	attempt, exam, answers, appeals, maxScore, err := s.appealRepo.GetSubmissionDetail(submissionID, studentID)
	if err != nil {
		return nil, errors.New("Submission not found")
	}

	res := map[string]interface{}{
		"submission_id": attempt.ID,
		"exam": map[string]interface{}{
			"id":        exam.ID,
			"title":     exam.Title,
			"exam_code": exam.ExamCode,
			"duration":  exam.Duration,
			"end_time":  exam.EndTime,
		},
		"score":        attempt.Score,
		"max_score":    maxScore,
		"started_at":   attempt.StartedAt,
		"submitted_at": attempt.SubmittedAt,
		"answers":      answers,
		"appeals":      appeals,
	}

	return res, nil
}

func (s *appealService) CreateAppeal(studentID uint, req dto.CreateAppealRequest) (*models.Appeal, error) {
	content := strings.TrimSpace(req.Content)
	if utf8.RuneCountInString(content) < 10 || utf8.RuneCountInString(content) > 1000 {
		return nil, errors.New("content must contain between 10 and 1000 characters")
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "Khiếu nại điểm thi"
	}
	if utf8.RuneCountInString(title) > 255 {
		return nil, errors.New("title must be at most 255 characters")
	}

	attempt, err := s.examRepo.FindAttemptByID(req.SubmissionID)
	if err != nil || attempt.UserID != studentID {
		return nil, errors.New("Submission not found")
	}
	if attempt.Status != "completed" {
		return nil, errors.New("Can only appeal a completed submission")
	}

	existing, _ := s.appealRepo.FindPendingAppealBySubmissionAndStudent(req.SubmissionID, studentID)
	if existing != nil {
		return nil, errors.New("Bạn đã có khiếu nại đang chờ xử lý cho bài thi này.")
	}

	exam, err := s.examRepo.FindExamByID(attempt.ExamID)
	if err != nil {
		return nil, errors.New("Exam not found")
	}

	// Thời hạn phúc khảo: chỉ tiếp nhận trong vòng 14 ngày sau khi kỳ thi kết thúc
	if !exam.EndTime.IsZero() && time.Now().After(exam.EndTime.Add(14*24*time.Hour)) {
		return nil, errors.New("Thời hạn nhận đơn phúc khảo cho kỳ thi này đã kết thúc (tối đa 14 ngày sau khi kết thúc kỳ thi).")
	}

	appeal := models.Appeal{
		ExamID:       exam.ID,
		SubmissionID: attempt.ID,
		StudentID:    studentID,
		TeacherID:    exam.TeacherID,
		Title:        title,
		Content:      content,
		Status:       models.AppealStatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.appealRepo.CreateAppeal(&appeal); err != nil {
		return nil, err
	}

	if s.notifSvc != nil {
		s.notifSvc.CreateNotification(
			exam.TeacherID,
			"Khiếu nại / Phúc khảo mới từ sinh viên",
			"Sinh viên đã gửi đơn phúc khảo cho bài thi "+exam.Title+": "+appeal.Title,
			fmt.Sprintf("/teacher/exams/%d", exam.ID),
		)
	}

	return &appeal, nil
}

func (s *appealService) ListExamAppeals(examID, teacherID uint, status, query string, page, pageSize int) ([]dto.TeacherAppealItem, int64, error) {
	exam, err := s.examRepo.FindExamByID(examID)
	if err != nil || exam.TeacherID != teacherID {
		return nil, 0, errors.New("Exam not found")
	}
	return s.appealRepo.ListTeacherAppealsByExam(examID, teacherID, status, query, page, pageSize)
}

func (s *appealService) GetAppealDetail(appealID, teacherID uint) (map[string]interface{}, error) {
	return s.appealRepo.GetTeacherAppealDetail(appealID, teacherID)
}

func (s *appealService) GetAppeal(appealID, teacherID uint) (*models.Appeal, error) {
	appeal, err := s.appealRepo.FindAppealByID(appealID)
	if err != nil || appeal.TeacherID != teacherID {
		return nil, errors.New("Appeal not found")
	}
	return appeal, nil
}

func (s *appealService) ReplyAppeal(appealID, teacherID uint, req dto.ReplyAppealRequest) (*models.Appeal, error) {
	appeal, err := s.appealRepo.FindAppealByID(appealID)
	if err != nil || appeal.TeacherID != teacherID {
		return nil, errors.New("Appeal not found")
	}
	if appeal.Status == models.AppealStatusRejected || appeal.Status == models.AppealStatusResolved {
		return nil, ErrAppealClosed
	}

	resp := strings.TrimSpace(req.Response)
	now := time.Now()
	appeal.TeacherResponse = &resp
	appeal.RespondedAt = &now
	appeal.Status = models.AppealStatusReplied
	appeal.UpdatedAt = now

	if err := s.appealRepo.UpdateAppeal(appeal); err != nil {
		return nil, err
	}

	if s.notifSvc != nil {
		s.notifSvc.CreateNotification(
			appeal.StudentID,
			"Cập nhật đơn phúc khảo",
			"Giảng viên đã phản hồi đơn phúc khảo của bạn: "+appeal.Title,
			fmt.Sprintf("/student/history/%d", appeal.SubmissionID),
		)
	}

	return appeal, nil
}

func (s *appealService) RejectAppeal(appealID, teacherID uint, req dto.ActionAppealRequest) (*models.Appeal, error) {
	appeal, err := s.appealRepo.FindAppealByID(appealID)
	if err != nil || appeal.TeacherID != teacherID {
		return nil, errors.New("Appeal not found")
	}
	if appeal.Status == models.AppealStatusRejected || appeal.Status == models.AppealStatusResolved {
		return nil, ErrAppealClosed
	}

	now := time.Now()
	if req.Response != nil && strings.TrimSpace(*req.Response) != "" {
		resp := strings.TrimSpace(*req.Response)
		appeal.TeacherResponse = &resp
		appeal.RespondedAt = &now
	}
	appeal.Status = models.AppealStatusRejected
	appeal.UpdatedAt = now

	if err := s.appealRepo.UpdateAppeal(appeal); err != nil {
		return nil, err
	}

	if s.notifSvc != nil {
		s.notifSvc.CreateNotification(
			appeal.StudentID,
			"Cập nhật đơn phúc khảo",
			"Đơn phúc khảo của bạn đã bị từ chối: "+appeal.Title,
			fmt.Sprintf("/student/history/%d", appeal.SubmissionID),
		)
	}

	return appeal, nil
}

func (s *appealService) ResolveAppeal(appealID, teacherID uint, req dto.ActionAppealRequest) (*models.Appeal, error) {
	appeal, err := s.appealRepo.FindAppealByID(appealID)
	if err != nil || appeal.TeacherID != teacherID {
		return nil, errors.New("Appeal not found")
	}
	if appeal.Status == models.AppealStatusRejected || appeal.Status == models.AppealStatusResolved {
		return nil, ErrAppealClosed
	}

	now := time.Now()
	if req.Response != nil && strings.TrimSpace(*req.Response) != "" {
		resp := strings.TrimSpace(*req.Response)
		appeal.TeacherResponse = &resp
		appeal.RespondedAt = &now
	}
	appeal.Status = models.AppealStatusResolved
	appeal.UpdatedAt = now

	if err := s.appealRepo.UpdateAppeal(appeal); err != nil {
		return nil, err
	}

	if s.notifSvc != nil {
		s.notifSvc.CreateNotification(
			appeal.StudentID,
			"Cập nhật đơn phúc khảo",
			"Đơn phúc khảo của bạn đã được chấp thuận: "+appeal.Title,
			fmt.Sprintf("/student/history/%d", appeal.SubmissionID),
		)
	}

	return appeal, nil
}
