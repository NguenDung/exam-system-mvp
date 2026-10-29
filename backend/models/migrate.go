package models

import (
	"log"

	"gorm.io/gorm"
)

// Migrate runs AutoMigrate for all models and adds unique composite indexes.
func Migrate(db *gorm.DB) {
	log.Println("Running database migrations...")

	if err := db.AutoMigrate(
		&User{},
		&Exam{},
		&Question{},
		&ExamAttempt{},
		&AttemptAnswer{},
		&Appeal{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	if err := db.Exec(`
		UPDATE exams
		SET published_at = updated_at
		WHERE status = 'published' AND published_at IS NULL
	`).Error; err != nil {
		log.Fatalf("Failed to backfill exam publication timestamps: %v", err)
	}

	// Unique composite index: (exam_id, user_id) on exam_attempts
	// Note: exam_attempts has no deleted_at column — do not use a partial WHERE.
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_exam_attempts_exam_user
		ON exam_attempts (exam_id, user_id)
	`).Error; err != nil {
		log.Printf("Warning: could not create idx_exam_attempts_exam_user: %v", err)
	}

	// Unique composite index: (attempt_id, question_id) on attempt_answers
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_attempt_answers_attempt_question
		ON attempt_answers (attempt_id, question_id)
	`).Error; err != nil {
		log.Printf("Warning: could not create idx_attempt_answers_attempt_question: %v", err)
	}

	log.Println("Database migrations completed successfully")
}
