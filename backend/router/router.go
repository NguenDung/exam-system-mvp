package router

import (
	"time"

	"exam-system-mvp/handlers"
	"exam-system-mvp/middleware"
	"exam-system-mvp/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures and returns the Gin engine with all routes.
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// ── CORS ──────────────────────────────────────────────────────────────────
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ── Health check ──────────────────────────────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "exam-system-mvp"})
	})

	api := r.Group("/api")

	// ── Auth routes (public) ──────────────────────────────────────────────────
	auth := api.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	// ── Teacher routes ────────────────────────────────────────────────────────
	teacher := api.Group("/teacher")
	teacher.Use(middleware.AuthRequired(), middleware.RequireRole(models.RoleTeacher))
	{
		teacher.GET("/exams", handlers.TeacherListExams)
		teacher.POST("/exams", handlers.TeacherCreateExam)
		teacher.GET("/exams/:id", handlers.TeacherGetExam)
		teacher.POST("/exams/:id/questions", handlers.TeacherAddQuestions)
		teacher.PATCH("/exams/:id/publish", handlers.TeacherPublishExam)
		teacher.GET("/exams/:id/results", handlers.TeacherGetResults)
		teacher.GET("/exams/:id/appeals", handlers.TeacherListExamAppeals)
		teacher.GET("/appeals/:appealId", handlers.TeacherGetAppeal)
		teacher.PATCH("/appeals/:appealId/reply", handlers.TeacherReplyAppeal)
		teacher.PATCH("/appeals/:appealId/reject", handlers.TeacherRejectAppeal)
		teacher.PATCH("/appeals/:appealId/resolve", handlers.TeacherResolveAppeal)
		teacher.DELETE("/questions/:question_id", handlers.TeacherDeleteQuestion)
	}

	// ── Student routes ────────────────────────────────────────────────────────
	student := api.Group("/student")
	student.Use(middleware.AuthRequired(), middleware.RequireRole(models.RoleStudent))
	{
		student.GET("/exams/lookup", handlers.StudentLookupExam)
		student.GET("/history", handlers.StudentGetHistory)
		student.GET("/history/:submissionId", handlers.StudentGetHistoryDetail)
		student.POST("/appeals", handlers.StudentCreateAppeal)
		student.GET("/attempts", handlers.StudentListAttempts)
		student.POST("/exams/:id/start", handlers.StudentStartExam)
		student.POST("/attempts/:attempt_id/save", handlers.StudentSaveAnswers)
		student.POST("/attempts/:attempt_id/submit", handlers.StudentSubmitExam)
		student.GET("/attempts/:attempt_id/result", handlers.StudentGetResult)
	}

	return r
}
