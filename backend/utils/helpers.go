package utils

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultJWTSecret = "exam_mvp_super_secret_key_2026"

// Claims is the JWT payload used by auth middleware and token generation.
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GetJWTSecret returns JWT_SECRET from the environment, with a stable
// fallback for local development and tests so go test does not fail.
func GetJWTSecret() string {
	if v, ok := os.LookupEnv("JWT_SECRET"); ok && v != "" {
		return v
	}
	return defaultJWTSecret
}

// GenerateJWT creates a signed JWT for the given user.
func GenerateJWT(userID uint, email, role string) (string, error) {
	secret := GetJWTSecret()

	claims := Claims{
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
