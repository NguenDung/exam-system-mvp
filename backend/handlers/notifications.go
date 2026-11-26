package handlers

import (
	"errors"
	"net/http"

	"exam-system-mvp/config"
	"exam-system-mvp/repository"
	"exam-system-mvp/service"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service service.NotificationService
}

func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: svc}
}

func getDefaultNotificationService() service.NotificationService {
	repo := repository.NewNotificationRepository(config.DB)
	return service.NewNotificationService(repo)
}

func getDefaultNotificationHandler() *NotificationHandler {
	return NewNotificationHandler(getDefaultNotificationService())
}

func CreateNotification(userID uint, title, message string, targetURL ...string) {
	getDefaultNotificationService().CreateNotification(userID, title, message, targetURL...)
}

// GetNotifications godoc
// @Summary Lấy danh sách thông báo và số lượng chưa đọc
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.NotificationListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/notifications [get]
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	res, err := h.service.GetNotifications(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load notifications"})
		return
	}

	c.JSON(http.StatusOK, res)
}

// MarkNotificationAsRead godoc
// @Summary Đánh dấu một thông báo là đã đọc
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID thông báo"
// @Success 200 {object} dto.MessageResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/notifications/{id}/read [patch]
func (h *NotificationHandler) MarkNotificationAsRead(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	notifID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	err := h.service.MarkAsRead(notifID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllNotificationsAsRead godoc
// @Summary Đánh dấu toàn bộ thông báo là đã đọc
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.MessageResponse
// @Router /api/notifications/read-all [patch]
func (h *NotificationHandler) MarkAllNotificationsAsRead(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	if err := h.service.MarkAllAsRead(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// DeleteNotification godoc
// @Summary Xóa một thông báo
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID thông báo"
// @Success 200 {object} dto.MessageResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	notifID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	err := h.service.DeleteNotification(notifID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// ClearReadNotifications godoc
// @Summary Xóa toàn bộ thông báo đã đọc
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.MessageResponse
// @Router /api/notifications/clear-read [delete]
func (h *NotificationHandler) ClearReadNotifications(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	if err := h.service.ClearReadNotifications(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear read notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Read notifications cleared successfully"})
}

func GetNotifications(c *gin.Context) {
	getDefaultNotificationHandler().GetNotifications(c)
}

func MarkNotificationAsRead(c *gin.Context) {
	getDefaultNotificationHandler().MarkNotificationAsRead(c)
}

func MarkAllNotificationsAsRead(c *gin.Context) {
	getDefaultNotificationHandler().MarkAllNotificationsAsRead(c)
}

func DeleteNotification(c *gin.Context) {
	getDefaultNotificationHandler().DeleteNotification(c)
}

func ClearReadNotifications(c *gin.Context) {
	getDefaultNotificationHandler().ClearReadNotifications(c)
}
