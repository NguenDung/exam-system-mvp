package dto

import "time"

type CreateAppealRequest struct {
	SubmissionID uint   `json:"submission_id"`
	SubmissionId uint   `json:"submissionId"`
	Title        string `json:"title" binding:"max=255"`
	Content      string `json:"content" binding:"required"`
}

type ReplyAppealRequest struct {
	Response string `json:"response" binding:"required,min=1"`
}

type ActionAppealRequest struct {
	Response *string `json:"response"`
}

type TeacherAppealItem struct {
	ID           uint      `json:"id"`
	SubmissionID uint      `json:"submission_id"`
	StudentID    uint      `json:"student_id"`
	StudentName  string    `json:"student_name"`
	StudentEmail string    `json:"student_email"`
	ExamTitle    string    `json:"exam_title"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	Score        float64   `json:"score"`
}
