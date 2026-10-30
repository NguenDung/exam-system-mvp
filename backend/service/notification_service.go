package service

import (
	"errors"
	"log"
	"time"

	"exam-system-mvp/dto"
	"exam-system-mvp/models"
	"exam-system-mvp/repository"

	"gorm.io/gorm"
)

var (
	ErrNotificationNotFound = errors.New("Notification not found")
)

type NotificationService interface {
	GetNotifications(userID uint) (*dto.NotificationListResponse, error)
	MarkAsRead(id, userID uint) error
	MarkAllAsRead(userID uint) error
	DeleteNotification(id, userID uint) error
	ClearReadNotifications(userID uint) error
	CreateNotification(userID uint, title, message string, targetURL ...string)
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) CreateNotification(userID uint, title, message string, targetURL ...string) {
	if userID == 0 || title == "" || message == "" {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("CreateNotification panic: %v", r)
		}
	}()

	var target string
	if len(targetURL) > 0 {
		target = targetURL[0]
	}

	notif := models.Notification{
		UserID:    userID,
		Title:     title,
		Message:   message,
		TargetURL: target,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(&notif); err != nil {
		log.Printf("CreateNotification error: %v", err)
	}
}

func (s *notificationService) GetNotifications(userID uint) (*dto.NotificationListResponse, error) {
	notifs, err := s.repo.GetLatestByUserID(userID, 20)
	if err != nil {
		return nil, err
	}

	unreadCount, err := s.repo.CountUnreadByUserID(userID)
	if err != nil {
		return nil, err
	}

	return &dto.NotificationListResponse{
		Notifications: notifs,
		UnreadCount:   unreadCount,
	}, nil
}

func (s *notificationService) MarkAsRead(id, userID uint) error {
	notif, err := s.repo.FindByIDAndUserID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotificationNotFound
		}
		return err
	}

	if !notif.IsRead {
		return s.repo.MarkAsRead(id, userID)
	}
	return nil
}

func (s *notificationService) MarkAllAsRead(userID uint) error {
	return s.repo.MarkAllAsRead(userID)
}

func (s *notificationService) DeleteNotification(id, userID uint) error {
	rows, err := s.repo.Delete(id, userID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

func (s *notificationService) ClearReadNotifications(userID uint) error {
	return s.repo.ClearRead(userID)
}
