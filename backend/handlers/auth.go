package handlers

import (
	"errors"
	"exam-system-mvp/apiresponse"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"exam-system-mvp/models"
	"exam-system-mvp/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterInput is the expected body for POST /api/auth/register.
// Email is validated after normalize (trim + lower), not via binding:"email",
// so leading/trailing spaces do not reject a valid address.
type RegisterInput struct {
	FullName string `json:"full_name" binding:"required,min=2,max=120"`
	Email    string `json:"email" binding:"required,max=254"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

// LoginInput is the expected body for POST /api/auth/login
type LoginInput struct {
	Email    string `json:"email" binding:"required,max=254"`
	Password string `json:"password" binding:"required,max=72"`
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

// bcrypt rejects passwords longer than 72 bytes; enforce before hashing.
func passwordByteLenOK(password string) bool {
	return len(password) <= 72 && utf8.ValidString(password)
}

// Register handles student self-registration.
// POST /api/auth/register
func (h *UserHandler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiresponse.Validation(c, err)
		return
	}
	fullName := strings.TrimSpace(input.FullName)
	nameLength := utf8.RuneCountInString(fullName)
	if nameLength < 2 || nameLength > 120 {
		apiresponse.Error(c, http.StatusBadRequest, "Full name must contain between 2 and 120 characters")
		return
	}

	email := normalizeEmail(input.Email)
	if !isValidEmail(email) {
		apiresponse.Error(c, http.StatusBadRequest, "Invalid email format")
		return
	}
	if !passwordByteLenOK(input.Password) {
		apiresponse.Error(c, http.StatusBadRequest, "Password must be at most 72 bytes")
		return
	}

	// Check if email already exists.
	var existing models.User
	result := h.db.Where("email = ?", email).First(&existing)
	if result.Error == nil {
		apiresponse.Error(c, http.StatusConflict, "Email already registered")
		return
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to check whether email is registered", result.Error)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to hash password", err)
		return
	}

	user := models.User{
		FullName:     fullName,
		Email:        email,
		PasswordHash: string(hash),
		Role:         models.RoleStudent,
	}

	if err := h.db.Create(&user).Error; err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	apiresponse.JSON(c, http.StatusCreated, RegisterResponse{
		ResponseMeta: ResponseMeta{Code: "USER_REGISTERED"},
		Message:      apiresponse.SuccessMessage(c, "USER_REGISTERED"),
		User:         PublicUserDTO{ID: user.ID, FullName: user.FullName, Email: user.Email, Role: user.Role},
	})
}

// Login handles email/password login and returns a JWT.
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiresponse.Validation(c, err)
		return
	}

	email := normalizeEmail(input.Email)
	if !isValidEmail(email) {
		apiresponse.Error(c, http.StatusBadRequest, "Invalid email format")
		return
	}
	if !passwordByteLenOK(input.Password) {
		apiresponse.Error(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	var user models.User
	if err := h.db.Where("email = ?", email).First(&user).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			apiresponse.Error(c, http.StatusInternalServerError, "Failed to load user for login", err)
			return
		}
		apiresponse.Error(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		apiresponse.Error(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}
	if user.Role != models.RoleTeacher && user.Role != models.RoleStudent {
		apiresponse.Error(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := utils.GenerateJWTWithConfig(user.ID, user.Email, user.Role, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "Failed to generate token", err)
		return
	}

	apiresponse.JSON(c, http.StatusOK, LoginResponse{
		ResponseMeta: ResponseMeta{Code: "USER_AUTHENTICATED"},
		Token:        token,
		User:         PublicUserDTO{ID: user.ID, FullName: user.FullName, Email: user.Email, Role: user.Role},
		ExpiresAt:    time.Now().Add(h.jwtExpiry).Unix(),
	})
}
