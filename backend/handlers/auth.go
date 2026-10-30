package handlers

import (
	"errors"
	"net/http"

	"exam-system-mvp/config"
	"exam-system-mvp/dto"
	"exam-system-mvp/repository"
	"exam-system-mvp/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func getDefaultAuthHandler() *AuthHandler {
	userRepo := repository.NewUserRepository(config.DB)
	authSvc := service.NewAuthService(userRepo)
	return NewAuthHandler(authSvc)
}

// Register handles student registration.
// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.authService.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail), errors.Is(err, service.ErrPasswordTooLong):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, res)
}

// Login handles user authentication and issuance of JWT.
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.authService.Login(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, res)
}

func Register(c *gin.Context) {
	getDefaultAuthHandler().Register(c)
}

func Login(c *gin.Context) {
	getDefaultAuthHandler().Login(c)
}
