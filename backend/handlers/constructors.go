package handlers

import (
	"time"

	"gorm.io/gorm"
)

type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
	jwtExpiry time.Duration
}
type UserHandler struct{ db *gorm.DB }
type StudentHandler struct{ db *gorm.DB }
type TeacherHandler struct{ db *gorm.DB }

func NewAuthHandler(db *gorm.DB, jwtSecret string, jwtExpiry time.Duration) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}
func NewUserHandler(db *gorm.DB) *UserHandler       { return &UserHandler{db: db} }
func NewStudentHandler(db *gorm.DB) *StudentHandler { return &StudentHandler{db: db} }
func NewTeacherHandler(db *gorm.DB) *TeacherHandler { return &TeacherHandler{db: db} }
