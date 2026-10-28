package models

import (
	"time"

	"gorm.io/gorm"
)

// ─── User ────────────────────────────────────────────────────────────────────

type User struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"`
	FullName     string         `gorm:"not null;size:255" json:"full_name"`
	Role         string         `gorm:"not null;default:'student';size:20" json:"role"` // 'teacher' | 'student'
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	Exams        []Exam        `gorm:"foreignKey:TeacherID" json:"-"`
	ExamAttempts []ExamAttempt `gorm:"foreignKey:UserID" json:"-"`
}

// ─── Exam ────────────────────────────────────────────────────────────────────

type Exam struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	TeacherID uint           `gorm:"not null;index" json:"teacher_id"`
	ExamCode  string         `gorm:"uniqueIndex;size:20" json:"exam_code"`
	Title     string         `gorm:"not null;size:500" json:"title"`
	Duration  int            `gorm:"not null" json:"duration"` // minutes
	StartTime time.Time      `json:"start_time"`
	EndTime   time.Time      `json:"end_time"`
	Status    string         `gorm:"not null;default:'draft';size:20" json:"status"` // 'draft' | 'published'
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	Teacher   User       `gorm:"foreignKey:TeacherID" json:"-"`
	Questions []Question `gorm:"foreignKey:ExamID;constraint:OnDelete:CASCADE" json:"questions,omitempty"`
}

// ─── Question ────────────────────────────────────────────────────────────────

type Question struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ExamID        uint      `gorm:"not null;index" json:"exam_id"`
	SortOrder     int       `gorm:"not null;default:0" json:"sort_order"`
	Prompt        string    `gorm:"not null;type:text" json:"prompt"`
	OptionA       string    `gorm:"not null;type:text" json:"option_a"`
	OptionB       string    `gorm:"not null;type:text" json:"option_b"`
	OptionC       string    `gorm:"not null;type:text" json:"option_c"`
	OptionD       string    `gorm:"not null;type:text" json:"option_d"`
	CorrectOption string    `gorm:"not null;size:1" json:"correct_option"` // 'A'|'B'|'C'|'D'
	Points        float64   `gorm:"not null;default:1.0;type:decimal(5,2)" json:"points"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// QuestionSafe is returned to students — without correct_option
type QuestionSafe struct {
	ID        uint    `json:"id"`
	ExamID    uint    `json:"exam_id"`
	SortOrder int     `json:"sort_order"`
	Prompt    string  `json:"prompt"`
	OptionA   string  `json:"option_a"`
	OptionB   string  `json:"option_b"`
	OptionC   string  `json:"option_c"`
	OptionD   string  `json:"option_d"`
	Points    float64 `json:"points"`
}

func ToQuestionSafe(q Question) QuestionSafe {
	return QuestionSafe{
		ID:        q.ID,
		ExamID:    q.ExamID,
		SortOrder: q.SortOrder,
		Prompt:    q.Prompt,
		OptionA:   q.OptionA,
		OptionB:   q.OptionB,
		OptionC:   q.OptionC,
		OptionD:   q.OptionD,
		Points:    q.Points,
	}
}

// ─── ExamAttempt ─────────────────────────────────────────────────────────────

type ExamAttempt struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	ExamID       uint       `gorm:"not null;index" json:"exam_id"`
	UserID       uint       `gorm:"not null;index" json:"user_id"`
	Status       string     `gorm:"not null;default:'in_progress';size:20" json:"status"` // 'in_progress' | 'completed'
	Score        float64    `gorm:"not null;default:0;type:decimal(8,2)" json:"score"`
	CorrectCount int        `gorm:"not null;default:0" json:"correct_count"`
	WrongCount   int        `gorm:"not null;default:0" json:"wrong_count"`
	StartedAt    time.Time  `gorm:"not null" json:"started_at"`
	LastSavedAt  *time.Time `json:"last_saved_at"`
	SubmittedAt  *time.Time `json:"submitted_at"`

	// Associations
	Exam            Exam            `gorm:"foreignKey:ExamID" json:"-"`
	User            User            `gorm:"foreignKey:UserID" json:"-"`
	AttemptAnswers  []AttemptAnswer `gorm:"foreignKey:AttemptID;constraint:OnDelete:CASCADE" json:"-"`
}

// ─── AttemptAnswer ───────────────────────────────────────────────────────────

type AttemptAnswer struct {
	ID            uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	AttemptID     uint    `gorm:"not null;index" json:"attempt_id"`
	QuestionID    uint    `gorm:"not null;index" json:"question_id"`
	ChosenOption  *string `gorm:"size:1" json:"chosen_option"` // nullable
	IsCorrect     bool    `gorm:"not null;default:false" json:"is_correct"`
	PointsAwarded float64 `gorm:"not null;default:0;type:decimal(5,2)" json:"points_awarded"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Associations
	Question Question `gorm:"foreignKey:QuestionID" json:"-"`
}
