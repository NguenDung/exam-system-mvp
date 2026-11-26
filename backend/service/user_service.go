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

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound               = errors.New("User not found")
	ErrCannotChangeOwnRole        = errors.New("Cannot change your own role")
	ErrCannotDemotePrimaryTeacher = errors.New("Cannot demote the primary teacher account")
	ErrTeacherHasOwnedResources   = errors.New("Cannot demote a teacher who owns existing exams or classrooms")
	ErrInvalidRole                = errors.New("Role must be either 'teacher' or 'student'")
)

type UserService interface {
	GetProfile(userID uint) (*models.User, error)
	UpdateProfile(userID uint, req dto.UpdateProfileRequest) (*models.User, error)
	GetDirectory(q, role, department string) ([]dto.UserDirectoryItem, error)
	UpdateUserRole(actorID, targetID uint, req dto.UpdateUserRoleRequest) (*models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
	notifSvc NotificationService
}

func NewUserService(userRepo repository.UserRepository, notifSvc ...NotificationService) UserService {
	var svc NotificationService
	if len(notifSvc) > 0 {
		svc = notifSvc[0]
	}
	return &userService{userRepo: userRepo, notifSvc: svc}
}

func (s *userService) GetProfile(userID uint) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *userService) UpdateProfile(userID uint, req dto.UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	updates := map[string]interface{}{}

	if req.FullName != nil {
		name := strings.TrimSpace(*req.FullName)
		if name == "" {
			return nil, errors.New("Full name cannot be empty")
		}
		updates["full_name"] = name
	}
	if req.AvatarURL != nil {
		updates["avatar_url"] = strings.TrimSpace(*req.AvatarURL)
	}
	if req.Bio != nil {
		updates["bio"] = strings.TrimSpace(*req.Bio)
	}
	if req.Department != nil {
		updates["department"] = strings.TrimSpace(*req.Department)
	}
	if req.StudentCode != nil {
		updates["student_code"] = strings.TrimSpace(*req.StudentCode)
	}
	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		pass := *req.Password
		if len(pass) < 6 || len(pass) > 72 || !utf8.ValidString(pass) {
			return nil, errors.New("Password must be between 6 and 72 bytes")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("Failed to hash password")
		}
		updates["password_hash"] = string(hash)
	}

	if len(updates) > 0 {
		if err := s.userRepo.UpdateFields(user, updates); err != nil {
			return nil, errors.New("Failed to update profile")
		}
	}

	return s.userRepo.FindByID(userID)
}

func (s *userService) GetDirectory(q, role, department string) ([]dto.UserDirectoryItem, error) {
	q = strings.ToLower(strings.TrimSpace(q))
	role = strings.ToLower(strings.TrimSpace(role))
	department = strings.ToLower(strings.TrimSpace(department))

	users, err := s.userRepo.FindDirectory(q, role, department, 100)
	if err != nil {
		return nil, err
	}

	items := make([]dto.UserDirectoryItem, 0, len(users))
	for _, u := range users {
		items = append(items, dto.UserDirectoryItem{
			ID:          u.ID,
			FullName:    u.FullName,
			Email:       u.Email,
			Role:        u.Role,
			AvatarURL:   u.AvatarURL,
			Bio:         u.Bio,
			Department:  u.Department,
			StudentCode: u.StudentCode,
		})
	}
	return items, nil
}

func (s *userService) UpdateUserRole(actorID, targetID uint, req dto.UpdateUserRoleRequest) (*models.User, error) {
	if actorID == targetID {
		return nil, ErrCannotChangeOwnRole
	}

	newRole := strings.ToLower(strings.TrimSpace(req.Role))
	if newRole != models.RoleTeacher && newRole != models.RoleStudent {
		return nil, ErrInvalidRole
	}

	targetUser, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if targetUser.Role == newRole {
		return targetUser, nil
	}

	if newRole == models.RoleStudent {
		if strings.EqualFold(targetUser.Email, "teacher@test.com") {
			return nil, ErrCannotDemotePrimaryTeacher
		}
		hasOwned, err := s.userRepo.HasOwnedResources(targetID)
		if err != nil {
			return nil, errors.New("Failed to verify teacher resources")
		}
		if hasOwned {
			return nil, ErrTeacherHasOwnedResources
		}
	}

	oldRole := targetUser.Role
	if err := s.userRepo.UpdateRole(targetID, newRole); err != nil {
		return nil, errors.New("Failed to update role")
	}

	_ = s.userRepo.CreateAuditLog(&models.AuditLog{
		ActorID:   actorID,
		Action:    "UPDATE_USER_ROLE",
		TargetID:  targetID,
		Details:   fmt.Sprintf("Changed role of %s from %s to %s", targetUser.Email, oldRole, newRole),
		CreatedAt: time.Now(),
	})

	if s.notifSvc != nil {
		roleLabel := "Sinh viên"
		if newRole == models.RoleTeacher {
			roleLabel = "Giảng viên"
		}
		s.notifSvc.CreateNotification(
			targetID,
			"Cập nhật vai trò tài khoản",
			fmt.Sprintf("Tài khoản của bạn đã được phân quyền thành %s. Vui lòng tải lại trang hoặc đăng nhập lại nếu giao diện chưa tự chuyển.", roleLabel),
			"/profile",
		)
	}

	targetUser.Role = newRole
	return targetUser, nil
}
