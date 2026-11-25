package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config groups all settings so the application loads environment values once
// at startup and passes explicit configuration to the components that need it.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	HTTP     HTTPConfig
	Logging  LoggingConfig
}

type ServerConfig struct{ Port, Mode string }
type DatabaseConfig struct{ Host, Port, User, Password, Name, SSLMode, TimeZone string }
type AuthConfig struct {
	JWTSecret      string
	JWTExpiryHours int
}
type HTTPConfig struct{ AllowedOrigins []string }
type LoggingConfig struct{ Level, Format, File string }

// DB remains available for seed and migration utilities. Request handlers
// receive their database dependency through constructors instead of reading it.
var DB *gorm.DB

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading environment variables directly")
	}
	return Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8081"),
			Mode: getEnv("GIN_MODE", "release"),
		},
		Database: DatabaseConfig{
			Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
			User: getEnv("DB_USER", "postgres"), Password: getEnv("DB_PASSWORD", "postgres"),
			Name: getEnv("DB_NAME", "exam_mvp"), SSLMode: getEnv("DB_SSLMODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "Asia/Ho_Chi_Minh"),
		},
		Auth:    AuthConfig{JWTSecret: getEnv("JWT_SECRET", "exam_mvp_super_secret_key_2026"), JWTExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 24)},
		HTTP:    HTTPConfig{AllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000"))},
		Logging: LoggingConfig{Level: getEnv("LOG_LEVEL", "info"), Format: getEnv("LOG_FORMAT", "json"), File: getEnv("LOG_FILE", "logs/app.log")},
	}
}

func (c Config) ConnectDatabase() *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password,
		c.Database.Name, c.Database.SSLMode, c.Database.TimeZone)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Error)})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	DB = db
	log.Println("Database connection established successfully")
	return db
}

// ConnectDatabase preserves the original startup entry point.
func ConnectDatabase() { Load().ConnectDatabase() }

func GetEnv(key, fallback string) string { return getEnv(key, fallback) }

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value, err := strconv.Atoi(getEnv(key, strconv.Itoa(fallback)))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}
