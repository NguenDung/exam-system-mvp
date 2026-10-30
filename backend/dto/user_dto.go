package dto

type UpdateProfileRequest struct {
	FullName    *string `json:"full_name"`
	AvatarURL   *string `json:"avatar_url"`
	Bio         *string `json:"bio"`
	Department  *string `json:"department"`
	StudentCode *string `json:"student_code"`
	Password    *string `json:"password"`
}

type UserDirectoryItem struct {
	ID          uint   `json:"id"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
	Department  string `json:"department"`
	StudentCode string `json:"student_code"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}
