package dto

import (
	"time"

	"exam-system-mvp/models"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Service string `json:"service" example:"exam-system-mvp"`
}

type StartExamRequest struct {
	Confirmed bool `json:"confirmed" example:"true"`
}

type SaveAnswersResponse struct {
	Message string    `json:"message"`
	SavedAt time.Time `json:"saved_at"`
}

type ExamEnvelope struct {
	Message string       `json:"message,omitempty"`
	Exam    *models.Exam `json:"exam"`
}

type ExamsListResponse struct {
	Exams []models.Exam `json:"exams"`
}

type QuestionsListResponse struct {
	Message   string            `json:"message"`
	Questions []models.Question `json:"questions"`
}

type QuestionEnvelope struct {
	Message  string           `json:"message"`
	Question *models.Question `json:"question"`
}

type AttemptEnvelope struct {
	Message string              `json:"message,omitempty"`
	Attempt *models.ExamAttempt `json:"attempt"`
}

type AuditLogsResponse struct {
	Data []models.AuditLog `json:"data"`
	Logs []models.AuditLog `json:"logs"`
}

type UserProfileEnvelope struct {
	Message string              `json:"message,omitempty"`
	User    UserProfileResponse `json:"user"`
}

type DirectoryResponse struct {
	Users []UserDirectoryItem `json:"users"`
}

type ClassesListResponse struct {
	Classes []ClassListItem `json:"classes"`
}

type ClassEnvelope struct {
	Message string            `json:"message,omitempty"`
	Class   *models.Classroom `json:"class"`
}

type AppealEnvelope struct {
	Message string         `json:"message,omitempty"`
	Appeal  *models.Appeal `json:"appeal"`
}

type PaginationMeta struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

type TeacherAppealsResponse struct {
	Items      []TeacherAppealItem `json:"items"`
	Appeals    []TeacherAppealItem `json:"appeals"`
	Pagination PaginationMeta      `json:"pagination"`
}
