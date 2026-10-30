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
