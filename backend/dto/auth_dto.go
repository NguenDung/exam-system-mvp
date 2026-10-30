package dto

type RegisterRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

type UserSummaryResponse struct {
	ID        uint   `json:"id"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
}

type RegisterResponse struct {
	Message string              `json:"message"`
	User    UserSummaryResponse `json:"user"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,max=72"`
}

type UserProfileResponse struct {
	ID          uint   `json:"id"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
	Department  string `json:"department"`
	StudentCode string `json:"student_code"`
}

type LoginResponse struct {
	Token     string              `json:"token"`
	User      UserProfileResponse `json:"user"`
	ExpiresAt int64               `json:"expires_at"`
}
