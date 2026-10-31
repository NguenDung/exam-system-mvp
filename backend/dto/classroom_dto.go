package dto

import "time"

type CreateClassRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateClassRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type AddClassMemberRequest struct {
	StudentID   *uint   `json:"student_id"`
	Email       *string `json:"email"`
	StudentCode *string `json:"student_code"`
}

type JoinClassRequest struct {
	Code string `json:"code" binding:"required"`
}

type ClassListItem struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	TeacherID   uint      `json:"teacher_id"`
	TeacherName string    `json:"teacher_name,omitempty"`
	MemberCount int64     `json:"member_count"`
	ExamCount   int64     `json:"exam_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type MemberInfo struct {
	ID          uint      `json:"id"`
	StudentID   uint      `json:"student_id"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email"`
	AvatarURL   string    `json:"avatar_url"`
	Bio         string    `json:"bio"`
	Department  string    `json:"department"`
	StudentCode string    `json:"student_code"`
	JoinedAt    time.Time `json:"joined_at"`
}

type ClassDetailTeacher struct {
	ID         uint   `json:"id"`
	FullName   string `json:"full_name"`
	Email      string `json:"email"`
	AvatarURL  string `json:"avatar_url"`
	Bio        string `json:"bio"`
	Department string `json:"department"`
}

type ClassDetailClass struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	TeacherID   uint      `json:"teacher_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type ClassDetailResponse struct {
	Class   ClassDetailClass   `json:"class"`
	Teacher ClassDetailTeacher `json:"teacher"`
	Members []MemberInfo       `json:"members"`
	Exams   interface{}        `json:"exams"`
}

type StudentGradeEntry struct {
	AttemptID      *uint      `json:"attempt_id"`
	Score          *float64   `json:"score"`
	Status         string     `json:"status"` // "completed" | "in_progress" | "not_attempted"
	ViolationCount int        `json:"violation_count"`
	SubmittedAt    *time.Time `json:"submitted_at"`
}

type StudentGradebookRow struct {
	StudentID      uint                         `json:"student_id"`
	FullName       string                       `json:"full_name"`
	Email          string                       `json:"email"`
	StudentCode    string                       `json:"student_code"`
	Department     string                       `json:"department"`
	AvatarURL      string                       `json:"avatar_url"`
	ExamScores     map[uint]StudentGradeEntry   `json:"exam_scores"` // examID -> entry
	AverageScore   float64                      `json:"average_score"`
	CompletedCount int                          `json:"completed_count"`
	AcademicRank   string                       `json:"academic_rank"` // "Xuất sắc", "Giỏi", "Khá", "Trung bình", "Yếu"
}

type ClassGradebookExam struct {
	ID        uint      `json:"id"`
	ExamCode  string    `json:"exam_code"`
	Title     string    `json:"title"`
	Duration  int       `json:"duration"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
}

type ClassGradebookSummary struct {
	TotalStudents  int     `json:"total_students"`
	TotalExams     int     `json:"total_exams"`
	ClassAverage   float64 `json:"class_average"`
	HighestAverage float64 `json:"highest_average"`
	LowestAverage  float64 `json:"lowest_average"`
	AtRiskCount    int     `json:"at_risk_count"` // Average < 4.0
}

type ClassGradebookResponse struct {
	ClassID   uint                  `json:"class_id"`
	ClassName string                `json:"class_name"`
	ClassCode string                `json:"class_code"`
	Exams     []ClassGradebookExam  `json:"exams"`
	Students  []StudentGradebookRow `json:"students"`
	Summary   ClassGradebookSummary `json:"summary"`
}
