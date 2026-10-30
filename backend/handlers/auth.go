package handlers

import (
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"exam-system-mvp/config"
	"exam-system-mvp/models"
	"exam-system-mvp/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// RegisterInput is the expected body for POST /api/auth/register.
// Email is validated after normalize (trim + lower), not via binding:"email",
// so leading/trailing spaces do not reject a valid address.
type RegisterInput struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

// LoginInput is the expected body for POST /api/auth/login
type LoginInput struct {
	Email    string `json:"email" binding:"required"`
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
func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := normalizeEmail(input.Email)
	if !isValidEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}
	if !passwordByteLenOK(input.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at most 72 bytes"})
		return
	}

	// Check if email already exists
	var existing models.User
	if result := config.DB.Where("email = ?", email).First(&existing); result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		FullName:     strings.TrimSpace(input.FullName),
		Email:        email,
		PasswordHash: string(hash),
		Role:         models.RoleStudent,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"user": gin.H{
			"id":        user.ID,
			"full_name": user.FullName,
			"email":     user.Email,
			"role":      user.Role,
		},
	})
}

// Login handles email/password login and returns a JWT.
// POST /api/auth/login
func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := normalizeEmail(input.Email)
	if !isValidEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}
	if !passwordByteLenOK(input.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if user.Role != models.RoleTeacher && user.Role != models.RoleStudent {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":           user.ID,
			"full_name":    user.FullName,
			"email":        user.Email,
			"role":         user.Role,
			"avatar_url":   user.AvatarURL,
			"bio":          user.Bio,
			"department":   user.Department,
			"student_code": user.StudentCode,
		},
		"expires_at": time.Now().Add(24 * time.Hour).Unix(),
	})
}
