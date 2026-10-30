package models

import (
	"log"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	if err := db.AutoMigrate(
		&User{},
		&Classroom{},
		&ClassroomMember{},
		&Exam{},
		&Question{},
		&ExamAttempt{},
		&AttemptAnswer{},
		&Appeal{},
		&AuditLog{},
		&Notification{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	_ = db.Exec(`
		UPDATE exams
		SET published_at = updated_at
		WHERE status = 'published' AND published_at IS NULL
	`).Error

	_ = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_exam_attempts_exam_user
		ON exam_attempts (exam_id, user_id)
	`).Error

	_ = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_attempt_answers_attempt_question
		ON attempt_answers (attempt_id, question_id)
	`).Error

	_ = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_classroom_members_class_student
		ON classroom_members (classroom_id, student_id)
	`).Error
}
