package dto

import (
	"time"

	"exam-system-mvp/models"
)

type CreateExamRequest struct {
	Title           string    `json:"title" binding:"required"`
	Description     string    `json:"description"`
	Duration        int       `json:"duration"`
	DurationMinutes int       `json:"duration_minutes"`
	StartTime       time.Time `json:"start_time" binding:"required"`
	EndTime         time.Time `json:"end_time" binding:"required"`
	ClassroomID     *uint     `json:"classroom_id"`
}

type UpdateExamRequest struct {
	Title           *string    `json:"title"`
	Description     *string    `json:"description"`
	Duration        *int       `json:"duration"`
	DurationMinutes *int       `json:"duration_minutes"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	Status          *string    `json:"status"`
	ClassroomID     *uint      `json:"classroom_id"`
}

type QuestionDTO struct {
	Prompt        string  `json:"prompt" binding:"required"`
	OptionA       string  `json:"option_a" binding:"required"`
	OptionB       string  `json:"option_b" binding:"required"`
	OptionC       string  `json:"option_c" binding:"required"`
	OptionD       string  `json:"option_d" binding:"required"`
	CorrectOption string  `json:"correct_option" binding:"required"`
	Points        float64 `json:"points"`
}

type AddQuestionsRequest struct {
	Questions []QuestionDTO `json:"questions" binding:"required"`
}

type AnswerInputDTO struct {
	QuestionID   uint    `json:"question_id" binding:"required"`
	ChosenOption *string `json:"chosen_option" binding:"omitempty,oneof=A B C D"`
}

type SaveAnswersRequest struct {
	Answers      []AnswerInputDTO `json:"answers"`
	QuestionID   uint             `json:"question_id"`
	ChosenOption *string          `json:"chosen_option"`
}

type SubmitExamRequest struct {
	Answers []AnswerInputDTO `json:"answers"`
}

type ViolationRequest struct {
	Reason string `json:"reason"`
}

type UpdateScoreRequest struct {
	Score       float64 `json:"score" binding:"required"`
	TeacherNote string  `json:"teacher_note"`
	Note        string  `json:"note"`
}

type CreateStudentRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateStudentRequest struct {
	FullName string  `json:"full_name"`
	Email    string  `json:"email"`
	Password *string `json:"password"`
}

type StartExamResponse struct {
	Attempt          models.ExamAttempt     `json:"attempt"`
	Questions        []models.QuestionSafe  `json:"questions"`
	Duration         int                    `json:"duration"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          time.Time              `json:"end_time"`
	RemainingSeconds int64                  `json:"remaining_seconds"`
	SavedAnswers     []models.AttemptAnswer `json:"saved_answers,omitempty"`
	AutoSubmitted    bool                   `json:"auto_submitted,omitempty"`
	Message          string                 `json:"message,omitempty"`
}
