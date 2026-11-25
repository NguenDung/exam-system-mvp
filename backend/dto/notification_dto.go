package dto

import (
	"time"

	"exam-system-mvp/models"
)

type NotificationResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	TargetURL string    `json:"target_url"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type NotificationListResponse struct {
	Notifications []models.Notification `json:"notifications"`
	UnreadCount   int64                 `json:"unread_count"`
}
