package handlers

import (
	"time"

	"exam-system-mvp/models"
)

type ResponseMeta struct {
	Code string `json:"code"`
}

type PublicUserDTO struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type StudentSummaryDTO struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type RegisterResponse struct {
	ResponseMeta
	Message string        `json:"message"`
	User    PublicUserDTO `json:"user"`
}

type LoginResponse struct {
	ResponseMeta
	Token     string        `json:"token"`
	User      PublicUserDTO `json:"user"`
	ExpiresAt int64         `json:"expires_at"`
}

type ExamListResponse struct {
	ResponseMeta
	Exams []models.Exam `json:"exams"`
}
type ExamResponse struct {
	ResponseMeta
	Exam models.Exam `json:"exam"`
}
type MessageResponse struct {
	ResponseMeta
	Message string `json:"message"`
}
type QuestionsCreatedResponse struct {
	ResponseMeta
	Message   string            `json:"message"`
	Questions []models.Question `json:"questions"`
}
type PublishedExamResponse struct {
	ResponseMeta
	Message string      `json:"message"`
	Exam    models.Exam `json:"exam"`
}
type ExamStatisticsDTO struct {
	TotalStudents int     `json:"total_students"`
	AvgScore      float64 `json:"avg_score"`
	MaxScore      float64 `json:"max_score"`
	MinScore      float64 `json:"min_score"`
}
type TeacherResultsResponse struct {
	ResponseMeta
	Exam       models.Exam       `json:"exam"`
	Statistics ExamStatisticsDTO `json:"statistics"`
	Results    []StudentResult   `json:"results"`
}

type ExamLookupResponse struct {
	ResponseMeta
	Exam          models.Exam         `json:"exam"`
	ExamStatus    string              `json:"exam_status"`
	AttemptStatus string              `json:"attempt_status"`
	Attempt       *models.ExamAttempt `json:"attempt,omitempty"`
}
type AttemptSummaryDTO struct {
	models.ExamAttempt
	ExamTitle    string `json:"exam_title"`
	ExamCode     string `json:"exam_code"`
	ExamDuration int    `json:"exam_duration"`
}
type AttemptsResponse struct {
	ResponseMeta
	Attempts []AttemptSummaryDTO `json:"attempts"`
}
type SavedAnswerDTO struct {
	QuestionID   uint    `json:"question_id"`
	ChosenOption *string `json:"chosen_option"`
}
type AutoSubmittedResponse struct {
	ResponseMeta
	Message          string             `json:"message"`
	RemainingSeconds int64              `json:"remaining_seconds"`
	Attempt          models.ExamAttempt `json:"attempt"`
	AutoSubmitted    bool               `json:"auto_submitted"`
}
type StartExamResponse struct {
	ResponseMeta
	Attempt          models.ExamAttempt    `json:"attempt"`
	RemainingSeconds int64                 `json:"remaining_seconds"`
	Questions        []models.QuestionSafe `json:"questions"`
	SavedAnswers     []SavedAnswerDTO      `json:"saved_answers"`
}
type SaveAnswersResponse struct {
	ResponseMeta
	Message string    `json:"message"`
	SavedAt time.Time `json:"saved_at"`
}
type SubmitExamResponse struct {
	ResponseMeta
	Message string             `json:"message"`
	Attempt models.ExamAttempt `json:"attempt"`
}
type AnswerDetailDTO struct {
	QuestionID    uint    `json:"question_id"`
	Prompt        string  `json:"prompt"`
	OptionA       string  `json:"option_a"`
	OptionB       string  `json:"option_b"`
	OptionC       string  `json:"option_c"`
	OptionD       string  `json:"option_d"`
	CorrectOption *string `json:"correct_option,omitempty"`
	ChosenOption  *string `json:"chosen_option"`
	IsCorrect     bool    `json:"is_correct"`
	PointsAwarded float64 `json:"points_awarded"`
	Points        float64 `json:"points"`
}
type ExamResultResponse struct {
	ResponseMeta
	Attempt            models.ExamAttempt `json:"attempt"`
	Exam               models.Exam        `json:"exam"`
	ShowCorrectAnswers bool               `json:"show_correct_answers"`
	Answers            []AnswerDetailDTO  `json:"answers"`
}

type ExamSummaryDTO struct {
	ID       uint   `json:"id"`
	Title    string `json:"title"`
	ExamCode string `json:"exam_code,omitempty"`
}
type HistoryResponse struct {
	ResponseMeta
	Items      []StudentHistoryItem `json:"items"`
	Pagination pagination           `json:"pagination"`
}
type HistoryDetailResponse struct {
	ResponseMeta
	SubmissionID uint            `json:"submission_id"`
	Exam         ExamSummaryDTO  `json:"exam"`
	Score        float64         `json:"score"`
	MaxScore     float64         `json:"max_score"`
	StartedAt    time.Time       `json:"started_at"`
	SubmittedAt  *time.Time      `json:"submitted_at"`
	Answers      []HistoryAnswer `json:"answers"`
	Appeals      []models.Appeal `json:"appeals"`
}
type AppealCreatedResponse struct {
	ResponseMeta
	Message string        `json:"message"`
	Appeal  models.Appeal `json:"appeal"`
}
type AppealsResponse struct {
	ResponseMeta
	Items      []TeacherAppealItem `json:"items"`
	Pagination pagination          `json:"pagination"`
}
type SubmissionSummaryDTO struct {
	ID          uint       `json:"id"`
	Score       float64    `json:"score"`
	MaxScore    float64    `json:"max_score"`
	StartedAt   time.Time  `json:"started_at"`
	SubmittedAt *time.Time `json:"submitted_at"`
}
type TeacherAppealDetailResponse struct {
	ResponseMeta
	Appeal     models.Appeal         `json:"appeal"`
	Student    StudentSummaryDTO     `json:"student"`
	Exam       ExamSummaryDTO        `json:"exam"`
	Submission SubmissionSummaryDTO  `json:"submission"`
	Answers    []TeacherAppealAnswer `json:"answers"`
}
type AppealResponse struct {
	ResponseMeta
	Appeal models.Appeal `json:"appeal"`
}
