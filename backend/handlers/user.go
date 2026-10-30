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

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{userService: svc}
}

func getDefaultUserHandler() *UserHandler {
	userRepo := repository.NewUserRepository(config.DB)
	svc := service.NewUserService(userRepo)
	return NewUserHandler(svc)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":           user.ID,
			"full_name":    user.FullName,
			"email":        user.Email,
			"role":         user.Role,
			"avatar_url":   user.AvatarURL,
			"bio":          user.Bio,
			"department":   user.Department,
			"student_code": user.StudentCode,
			"created_at":   user.CreatedAt,
		},
	})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateProfile(userID, req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":           user.ID,
			"full_name":    user.FullName,
			"email":        user.Email,
			"role":         user.Role,
			"avatar_url":   user.AvatarURL,
			"bio":          user.Bio,
			"department":   user.Department,
			"student_code": user.StudentCode,
			"created_at":   user.CreatedAt,
		},
	})
}

func (h *UserHandler) GetDirectory(c *gin.Context) {
	q := c.Query("q")
	role := c.Query("role")
	dept := c.Query("department")

	users, err := h.userService.GetDirectory(q, role, dept)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch directory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *UserHandler) TeacherUpdateUserRole(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	targetID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateUserRole(teacherID, targetID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCannotChangeOwnRole), errors.Is(err, service.ErrInvalidRole):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
		"user": gin.H{
			"id":        user.ID,
			"full_name": user.FullName,
			"email":     user.Email,
			"role":      user.Role,
		},
	})
}
