package repository

import (
	"exam-system-mvp/models"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	GetLatestByUserID(userID uint, limit int) ([]models.Notification, error)
	CountUnreadByUserID(userID uint) (int64, error)
	FindByIDAndUserID(id, userID uint) (*models.Notification, error)
	MarkAsRead(id, userID uint) error
	MarkAllAsRead(userID uint) error
	Delete(id, userID uint) (int64, error)
	ClearRead(userID uint) error
	Create(notif *models.Notification) error
}

type gormNotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &gormNotificationRepository{db: db}
}

func (r *gormNotificationRepository) GetLatestByUserID(userID uint, limit int) ([]models.Notification, error) {
	var notifs []models.Notification
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&notifs).Error
	return notifs, err
}

func (r *gormNotificationRepository) CountUnreadByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return count, err
}

func (r *gormNotificationRepository) FindByIDAndUserID(id, userID uint) (*models.Notification, error) {
	var notif models.Notification
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&notif).Error
	if err != nil {
		return nil, err
	}
	return &notif, nil
}

func (r *gormNotificationRepository) MarkAsRead(id, userID uint) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

func (r *gormNotificationRepository) MarkAllAsRead(userID uint) error {
	return r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}

func (r *gormNotificationRepository) Delete(id, userID uint) (int64, error) {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Notification{})
	return res.RowsAffected, res.Error
}

func (r *gormNotificationRepository) ClearRead(userID uint) error {
	return r.db.Where("user_id = ? AND is_read = true", userID).Delete(&models.Notification{}).Error
}

func (r *gormNotificationRepository) Create(notif *models.Notification) error {
	return r.db.Create(notif).Error
}
