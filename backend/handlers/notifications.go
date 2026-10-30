package handlers

import (
	"log"
	"net/http"
	"time"

	"exam-system-mvp/config"
	"exam-system-mvp/models"

	"github.com/gin-gonic/gin"
)

// CreateNotification safely records a notification for a user.
// It never panics or disrupts the calling request if an error occurs.
func CreateNotification(userID uint, title, message string) {
	if userID == 0 || title == "" || message == "" {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Notification Panic] Recovered from panic in CreateNotification: %v", r)
		}
	}()

	notif := models.Notification{
		UserID:    userID,
		Title:     title,
		Message:   message,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	if err := config.DB.Create(&notif).Error; err != nil {
		log.Printf("[Notification Error] Failed to create notification for user %d: %v", userID, err)
	}
}

// GetNotifications returns the latest 20 notifications for the current user and their unread count.
// GET /api/notifications
func GetNotifications(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	var notifications []models.Notification
	if err := config.DB.Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Limit(20).
		Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load notifications"})
		return
	}

	var unreadCount int64
	config.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&unreadCount)

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"unread_count":  unreadCount,
	})
}

// MarkNotificationAsRead marks a single notification as read.
// PATCH /api/notifications/:id/read
func MarkNotificationAsRead(c *gin.Context) {
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

	var notif models.Notification
	if err := config.DB.Where("id = ? AND user_id = ?", notifID, userID).First(&notif).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	if !notif.IsRead {
		if err := config.DB.Model(&notif).Update("is_read", true).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllNotificationsAsRead marks all unread notifications for the current user as read.
// PATCH /api/notifications/read-all
func MarkAllNotificationsAsRead(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	if err := config.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}
