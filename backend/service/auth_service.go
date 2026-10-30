package service

import (
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"exam-system-mvp/dto"
	"exam-system-mvp/models"
	"exam-system-mvp/repository"
	"exam-system-mvp/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidEmail       = errors.New("Invalid email format")
	ErrPasswordTooLong    = errors.New("Password must be at most 72 bytes")
	ErrEmailAlreadyExists = errors.New("Email already registered")
	ErrInvalidCredentials = errors.New("Invalid email or password")
	ErrFailedToHash       = errors.New("Failed to hash password")
	ErrFailedToCreateUser = errors.New("Failed to create user")
	ErrFailedToGenToken   = errors.New("Failed to generate token")
)

type AuthService interface {
	Register(req dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(req dto.LoginRequest) (*dto.LoginResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

func passwordByteLenOK(password string) bool {
	return len(password) <= 72 && utf8.ValidString(password)
}

func (s *authService) Register(req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	email := normalizeEmail(req.Email)
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}
	if !passwordByteLenOK(req.Password) {
		return nil, ErrPasswordTooLong
	}

	existing, err := s.userRepo.FindByEmail(email)
	if err == nil && existing != nil {
		return nil, ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrFailedToHash
	}

	user := models.User{
		FullName:     strings.TrimSpace(req.FullName),
		Email:        email,
		PasswordHash: string(hash),
		Role:         models.RoleStudent,
	}

	if err := s.userRepo.Create(&user); err != nil {
		return nil, ErrFailedToCreateUser
	}

	return &dto.RegisterResponse{
		Message: "Registration successful",
		User: dto.UserSummaryResponse{
			ID:       user.ID,
			FullName: user.FullName,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}

func (s *authService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	email := normalizeEmail(req.Email)
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}
	if !passwordByteLenOK(req.Password) {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	if user.Role != models.RoleTeacher && user.Role != models.RoleStudent {
		return nil, ErrInvalidCredentials
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, ErrFailedToGenToken
	}

	return &dto.LoginResponse{
		Token: token,
		User: dto.UserProfileResponse{
			ID:          user.ID,
			FullName:    user.FullName,
			Email:       user.Email,
			Role:        user.Role,
			AvatarURL:   user.AvatarURL,
			Bio:         user.Bio,
			Department:  user.Department,
			StudentCode: user.StudentCode,
		},
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}, nil
}
