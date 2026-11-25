package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"exam-system-mvp/dto"
	"exam-system-mvp/models"
	"exam-system-mvp/repository"

	"gorm.io/gorm"
)

var (
	ErrClassroomNotFound     = errors.New("Classroom not found or you do not have permission")
	ErrClassroomAccessDenied = errors.New("You do not have access to this classroom")
	ErrStudentNotEnrolled    = errors.New("You are not enrolled in this classroom")
	ErrStudentNotFound       = errors.New("Student not found")
	ErrStudentAlreadyInClass = errors.New("Student is already in this class")
	ErrStudentNotInClass     = errors.New("Student is not in this class")
	ErrEmptyClassName        = errors.New("Class name cannot be empty")
	ErrEmptyClassCode        = errors.New("Class code is required")
	ErrMissingStudentParam   = errors.New("Please provide student_id, email, or student_code")
)

type ClassroomService interface {
	CreateClass(teacherID uint, req dto.CreateClassRequest) (*models.Classroom, error)
	UpdateClass(teacherID, classID uint, req dto.UpdateClassRequest) (*models.Classroom, error)
	DeleteClass(teacherID, classID uint) error
	AddMember(teacherID, classID uint, req dto.AddClassMemberRequest) (*dto.MemberInfo, error)
	RemoveMember(teacherID, classID, studentID uint) error
	JoinClass(studentID uint, req dto.JoinClassRequest) (*models.Classroom, error)
	ListClasses(userID uint, role string) ([]dto.ClassListItem, error)
	GetClassDetail(userID uint, role string, classID uint) (*dto.ClassDetailResponse, error)
	GetClassGradebook(teacherID, classID uint) (*dto.ClassGradebookResponse, error)
}

type classroomService struct {
	classRepo repository.ClassroomRepository
	userRepo  repository.UserRepository
	notifSvc  NotificationService
}

func NewClassroomService(
	classRepo repository.ClassroomRepository,
	userRepo repository.UserRepository,
	notifSvc NotificationService,
) ClassroomService {
	return &classroomService{
		classRepo: classRepo,
		userRepo:  userRepo,
		notifSvc:  notifSvc,
	}
}

func generateClassCode(repo repository.ClassroomRepository) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	for attempt := 0; attempt < 20; attempt++ {
		b := make([]byte, 6)
		for i := range b {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				b[i] = charset[i%len(charset)]
			} else {
				b[i] = charset[n.Int64()]
			}
		}
		code := string(b)
		exists, err := repo.IsCodeExists(code)
		if err == nil && !exists {
			return code
		}
	}
	return fmt.Sprintf("C%05d", time.Now().UnixNano()%100000)
}

func (s *classroomService) CreateClass(teacherID uint, req dto.CreateClassRequest) (*models.Classroom, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrEmptyClassName
	}

	class := models.Classroom{
		Name:        name,
		Code:        generateClassCode(s.classRepo),
		Description: strings.TrimSpace(req.Description),
		TeacherID:   teacherID,
	}

	if err := s.classRepo.Create(&class); err != nil {
		return nil, err
	}
	return &class, nil
}

func (s *classroomService) UpdateClass(teacherID, classID uint, req dto.UpdateClassRequest) (*models.Classroom, error) {
	class, err := s.classRepo.FindByID(classID)
	if err != nil || class.TeacherID != teacherID {
		return nil, ErrClassroomNotFound
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrEmptyClassName
		}
		updates["name"] = name
	}
	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
	}

	if len(updates) > 0 {
		if err := s.classRepo.Update(class, updates); err != nil {
			return nil, err
		}
	}

	return s.classRepo.FindByID(classID)
}

func (s *classroomService) DeleteClass(teacherID, classID uint) error {
	class, err := s.classRepo.FindByID(classID)
	if err != nil || class.TeacherID != teacherID {
		return ErrClassroomNotFound
	}
	return s.classRepo.Delete(classID)
}

func (s *classroomService) AddMember(teacherID, classID uint, req dto.AddClassMemberRequest) (*dto.MemberInfo, error) {
	class, err := s.classRepo.FindByID(classID)
	if err != nil || class.TeacherID != teacherID {
		return nil, ErrClassroomNotFound
	}

	var student *models.User
	var findErr error

	if req.StudentID != nil && *req.StudentID > 0 {
		st, err := s.userRepo.FindByID(*req.StudentID)
		if err == nil && st.Role == models.RoleStudent {
			student = st
		} else {
			findErr = ErrStudentNotFound
		}
	} else if req.Email != nil && strings.TrimSpace(*req.Email) != "" {
		email := strings.ToLower(strings.TrimSpace(*req.Email))
		st, err := s.userRepo.FindByEmail(email)
		if err == nil && st.Role == models.RoleStudent {
			student = st
		} else {
			findErr = ErrStudentNotFound
		}
	} else if req.StudentCode != nil && strings.TrimSpace(*req.StudentCode) != "" {
		code := strings.TrimSpace(*req.StudentCode)
		st, err := s.userRepo.FindByStudentCode(code)
		if err == nil && st.Role == models.RoleStudent {
			student = st
		} else {
			findErr = ErrStudentNotFound
		}
	} else {
		return nil, ErrMissingStudentParam
	}

	if findErr != nil || student == nil {
		return nil, ErrStudentNotFound
	}

	existing, err := s.classRepo.FindMember(class.ID, student.ID)
	if err == nil && existing != nil {
		return nil, ErrStudentAlreadyInClass
	}

	member := models.ClassroomMember{
		ClassroomID: class.ID,
		StudentID:   student.ID,
		JoinedAt:    time.Now(),
	}

	if err := s.classRepo.AddMember(&member); err != nil {
		return nil, err
	}

	if s.notifSvc != nil {
		s.notifSvc.CreateNotification(
			student.ID,
			"Đã được thêm vào lớp học",
			fmt.Sprintf("Bạn đã được thêm vào lớp học \"%s\".", class.Name),
			fmt.Sprintf("/classes/%d", class.ID),
		)
	}

	return &dto.MemberInfo{
		ID:          member.ID,
		StudentID:   student.ID,
		FullName:    student.FullName,
		Email:       student.Email,
		AvatarURL:   student.AvatarURL,
		Bio:         student.Bio,
		Department:  student.Department,
		StudentCode: student.StudentCode,
		JoinedAt:    member.JoinedAt,
	}, nil
}

func (s *classroomService) RemoveMember(teacherID, classID, studentID uint) error {
	class, err := s.classRepo.FindByID(classID)
	if err != nil || class.TeacherID != teacherID {
		return ErrClassroomNotFound
	}

	rows, err := s.classRepo.RemoveMember(classID, studentID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrStudentNotInClass
	}
	return nil
}

func (s *classroomService) JoinClass(studentID uint, req dto.JoinClassRequest) (*models.Classroom, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		return nil, ErrEmptyClassCode
	}

	class, err := s.classRepo.FindByCode(code)
	if err != nil {
		return nil, errors.New("Classroom not found with this code")
	}

	existing, err := s.classRepo.FindMember(class.ID, studentID)
	if err == nil && existing != nil {
		return nil, errors.New("You have already joined this classroom")
	}

	member := models.ClassroomMember{
		ClassroomID: class.ID,
		StudentID:   studentID,
		JoinedAt:    time.Now(),
	}

	if err := s.classRepo.AddMember(&member); err != nil {
		return nil, err
	}

	student, _ := s.userRepo.FindByID(studentID)
	studentName := "Sinh viên"
	if student != nil {
		studentName = student.FullName
	}

	if s.notifSvc != nil {
		s.notifSvc.CreateNotification(
			class.TeacherID,
			"Thành viên mới tham gia lớp",
			fmt.Sprintf("Sinh viên %s đã tham gia lớp \"%s\".", studentName, class.Name),
			fmt.Sprintf("/classes/%d", class.ID),
		)
	}

	return class, nil
}

func (s *classroomService) ListClasses(userID uint, role string) ([]dto.ClassListItem, error) {
	var classes []models.Classroom
	var err error

	if role == models.RoleTeacher {
		classes, err = s.classRepo.FindByTeacherID(userID)
	} else {
		classes, err = s.classRepo.FindEnrolledByStudentID(userID)
	}
	if err != nil {
		return nil, err
	}

	items := make([]dto.ClassListItem, 0, len(classes))
	for _, cl := range classes {
		memberCount, _ := s.classRepo.CountMembers(cl.ID)
		examCount, _ := s.classRepo.CountExams(cl.ID)

		items = append(items, dto.ClassListItem{
			ID:          cl.ID,
			Name:        cl.Name,
			Code:        cl.Code,
			Description: cl.Description,
			TeacherID:   cl.TeacherID,
			TeacherName: cl.Teacher.FullName,
			MemberCount: memberCount,
			ExamCount:   examCount,
			CreatedAt:   cl.CreatedAt,
		})
	}
	return items, nil
}

func (s *classroomService) GetClassDetail(userID uint, role string, classID uint) (*dto.ClassDetailResponse, error) {
	class, err := s.classRepo.FindByIDWithTeacher(classID)
	if err != nil {
		return nil, ErrClassroomNotFound
	}

	if role == models.RoleTeacher {
		if class.TeacherID != userID {
			return nil, ErrClassroomAccessDenied
		}
	} else {
		member, err := s.classRepo.FindMember(classID, userID)
		if err != nil || member == nil {
			return nil, ErrStudentNotEnrolled
		}
	}

	rawMembers, err := s.classRepo.GetMembersWithStudent(classID)
	if err != nil {
		return nil, err
	}

	members := make([]dto.MemberInfo, 0, len(rawMembers))
	for _, m := range rawMembers {
		members = append(members, dto.MemberInfo{
			ID:          m.ID,
			StudentID:   m.StudentID,
			FullName:    m.Student.FullName,
			Email:       m.Student.Email,
			AvatarURL:   m.Student.AvatarURL,
			Bio:         m.Student.Bio,
			Department:  m.Student.Department,
			StudentCode: m.Student.StudentCode,
			JoinedAt:    m.JoinedAt,
		})
	}

	exams, err := s.classRepo.GetClassExams(classID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		exams = []models.Exam{}
	}

	return &dto.ClassDetailResponse{
		Class: dto.ClassDetailClass{
			ID:          class.ID,
			Name:        class.Name,
			Code:        class.Code,
			Description: class.Description,
			TeacherID:   class.TeacherID,
			CreatedAt:   class.CreatedAt,
		},
		Teacher: dto.ClassDetailTeacher{
			ID:         class.Teacher.ID,
			FullName:   class.Teacher.FullName,
			Email:      class.Teacher.Email,
			AvatarURL:  class.Teacher.AvatarURL,
			Bio:        class.Teacher.Bio,
			Department: class.Teacher.Department,
		},
		Members: members,
		Exams:   exams,
	}, nil
}

func (s *classroomService) GetClassGradebook(teacherID, classID uint) (*dto.ClassGradebookResponse, error) {
	class, err := s.classRepo.FindByID(classID)
	if err != nil || class.TeacherID != teacherID {
		return nil, ErrClassroomNotFound
	}

	rawMembers, err := s.classRepo.GetMembersWithStudent(classID)
	if err != nil {
		return nil, err
	}

	exams, err := s.classRepo.GetClassExams(classID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		exams = []models.Exam{}
	}

	var examIDs []uint
	gradebookExams := make([]dto.ClassGradebookExam, 0, len(exams))
	for _, ex := range exams {
		examIDs = append(examIDs, ex.ID)
		gradebookExams = append(gradebookExams, dto.ClassGradebookExam{
			ID:        ex.ID,
			ExamCode:  ex.ExamCode,
			Title:     ex.Title,
			Duration:  ex.Duration,
			StartTime: ex.StartTime,
			EndTime:   ex.EndTime,
			Status:    ex.Status,
		})
	}

	attempts, err := s.classRepo.GetClassAttempts(examIDs)
	if err != nil {
		attempts = []models.ExamAttempt{}
	}

	// Map: userID -> examID -> ExamAttempt
	attemptMap := make(map[uint]map[uint]models.ExamAttempt)
	for _, a := range attempts {
		if _, ok := attemptMap[a.UserID]; !ok {
			attemptMap[a.UserID] = make(map[uint]models.ExamAttempt)
		}
		attemptMap[a.UserID][a.ExamID] = a
	}

	students := make([]dto.StudentGradebookRow, 0, len(rawMembers))
	var totalScoresSum float64
	var totalScoresCount int
	highestAvg := 0.0
	lowestAvg := 10.0
	atRiskCount := 0

	for _, m := range rawMembers {
		userAttempts := attemptMap[m.StudentID]
		examScores := make(map[uint]dto.StudentGradeEntry)
		var studentSum float64
		var studentCompleted int

		for _, ex := range exams {
			if att, exists := userAttempts[ex.ID]; exists {
				scoreVal := att.Score
				examScores[ex.ID] = dto.StudentGradeEntry{
					AttemptID:      &att.ID,
					Score:          &scoreVal,
					Status:         att.Status,
					ViolationCount: att.ViolationCount,
					SubmittedAt:    att.SubmittedAt,
				}
				if att.Status == "completed" {
					studentSum += att.Score
					studentCompleted++
				}
			} else {
				examScores[ex.ID] = dto.StudentGradeEntry{
					AttemptID:      nil,
					Score:          nil,
					Status:         "not_attempted",
					ViolationCount: 0,
					SubmittedAt:    nil,
				}
			}
		}

		var avgScore float64
		if studentCompleted > 0 {
			avgScore = studentSum / float64(studentCompleted)
		}

		var rank string
		switch {
		case studentCompleted == 0:
			rank = "Chưa có điểm"
		case avgScore >= 8.5:
			rank = "Xuất sắc"
		case avgScore >= 7.0:
			rank = "Giỏi"
		case avgScore >= 5.5:
			rank = "Khá"
		case avgScore >= 4.0:
			rank = "Trung bình"
		default:
			rank = "Yếu / Cảnh báo"
		}

		if studentCompleted > 0 {
			totalScoresSum += avgScore
			totalScoresCount++
			if avgScore > highestAvg {
				highestAvg = avgScore
			}
			if avgScore < lowestAvg {
				lowestAvg = avgScore
			}
			if avgScore < 4.0 {
				atRiskCount++
			}
		}

		students = append(students, dto.StudentGradebookRow{
			StudentID:      m.StudentID,
			FullName:       m.Student.FullName,
			Email:          m.Student.Email,
			StudentCode:    m.Student.StudentCode,
			Department:     m.Student.Department,
			AvatarURL:      m.Student.AvatarURL,
			ExamScores:     examScores,
			AverageScore:   avgScore,
			CompletedCount: studentCompleted,
			AcademicRank:   rank,
		})
	}

	classAvg := 0.0
	if totalScoresCount > 0 {
		classAvg = totalScoresSum / float64(totalScoresCount)
	} else {
		lowestAvg = 0.0
	}

	summary := dto.ClassGradebookSummary{
		TotalStudents:  len(rawMembers),
		TotalExams:     len(exams),
		ClassAverage:   classAvg,
		HighestAverage: highestAvg,
		LowestAverage:  lowestAvg,
		AtRiskCount:    atRiskCount,
	}

	return &dto.ClassGradebookResponse{
		ClassID:   class.ID,
		ClassName: class.Name,
		ClassCode: class.Code,
		Exams:     gradebookExams,
		Students:  students,
		Summary:   summary,
	}, nil
}
