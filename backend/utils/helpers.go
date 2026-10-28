package utils

import (
	"fmt"
	"math/rand"
	"time"

	"exam-system-mvp/config"
	"exam-system-mvp/middleware"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateJWT creates a signed JWT for the given user.
func GenerateJWT(userID uint, email, role string) (string, error) {
	secret := config.GetEnv("JWT_SECRET", "exam_mvp_super_secret_key_2026")

	claims := middleware.Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateExamCode creates a random 6-character exam code like EX2026.
func GenerateExamCode() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := "EX"
	for i := 0; i < 4; i++ {
		code += string(letters[r.Intn(len(letters))])
	}
	return fmt.Sprintf("%s", code)
}
