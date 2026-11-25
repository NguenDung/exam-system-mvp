package repository

import (
	"exam-system-mvp/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	FindByStudentCode(code string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	UpdateFields(user *models.User, updates map[string]interface{}) error
	UpdateRole(id uint, newRole string) error
	FindDirectory(q, role, department string, limit int) ([]models.User, error)
}

type gormUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) FindByStudentCode(code string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("student_code = ?", code).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *gormUserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *gormUserRepository) UpdateFields(user *models.User, updates map[string]interface{}) error {
	return r.db.Model(user).Updates(updates).Error
}

func (r *gormUserRepository) UpdateRole(id uint, newRole string) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("role", newRole).Error
}

func (r *gormUserRepository) FindDirectory(q, role, department string, limit int) ([]models.User, error) {
	db := r.db.Model(&models.User{})

	if role == models.RoleTeacher || role == models.RoleStudent {
		db = db.Where("role = ?", role)
	}
	if department != "" {
		db = db.Where("LOWER(department) LIKE ?", "%"+department+"%")
	}
	if q != "" {
		term := "%" + q + "%"
		db = db.Where("LOWER(full_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(student_code) LIKE ? OR LOWER(department) LIKE ?", term, term, term, term)
	}

	var users []models.User
	err := db.Order("full_name ASC").Limit(limit).Find(&users).Error
	return users, err
}
