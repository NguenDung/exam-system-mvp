package router

import (
	"net/http"
	"time"

	"exam-system-mvp/apiresponse"
	"exam-system-mvp/config"
	"exam-system-mvp/handlers"
	"exam-system-mvp/middleware"
	"exam-system-mvp/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type healthResponse struct {
	Code    string `json:"code"`
	Status  string `json:"status"`
	Service string `json:"service"`
}

// SetupRouter preserves the default startup and integration-test entry point.
func SetupRouter() *gin.Engine {
	return SetupRouterWithConfig(config.DB, config.Load())
}

// SetupRouterWithConfig wires dependencies once. Individual endpoint handlers
// are methods on objects created here, so they never construct themselves.
func SetupRouterWithConfig(db *gorm.DB, cfg config.Config) *gin.Engine {
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(middleware.RequestLogging(), middleware.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.HTTP.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "Accept-Language", "X-Trace-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Trace-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.NoRoute(middleware.NotFound)
	r.NoMethod(middleware.MethodNotAllowed)

	r.GET("/health", func(c *gin.Context) {
		apiresponse.JSON(c, http.StatusOK, healthResponse{Code: "SERVICE_HEALTHY", Status: "ok", Service: "exam-system-mvp"})
	})

	authHandler := handlers.NewAuthHandler(db, cfg.Auth.JWTSecret, time.Duration(cfg.Auth.JWTExpiryHours)*time.Hour)
	userHandler := handlers.NewUserHandler(db)
	studentHandler := handlers.NewStudentHandler(db)
	teacherHandler := handlers.NewTeacherHandler(db)
	api := r.Group("/api")

	auth := api.Group("/auth")
	auth.POST("/register", userHandler.Register)
	auth.POST("/login", authHandler.Login)

	teacher := api.Group("/teacher")
	teacher.Use(middleware.AuthRequiredWithSecret(cfg.Auth.JWTSecret), middleware.RequireRole(models.RoleTeacher))
	{
		teacher.GET("/exams", teacherHandler.TeacherListExams)
		teacher.POST("/exams", teacherHandler.TeacherCreateExam)
		teacher.GET("/exams/:id", teacherHandler.TeacherGetExam)
		teacher.POST("/exams/:id/questions", teacherHandler.TeacherAddQuestions)
		teacher.PATCH("/exams/:id/publish", teacherHandler.TeacherPublishExam)
		teacher.GET("/exams/:id/results", teacherHandler.TeacherGetResults)
		teacher.GET("/exams/:id/appeals", teacherHandler.TeacherListExamAppeals)
		teacher.GET("/appeals/:appealId", teacherHandler.TeacherGetAppeal)
		teacher.PATCH("/appeals/:appealId/reply", teacherHandler.TeacherReplyAppeal)
		teacher.PATCH("/appeals/:appealId/reject", teacherHandler.TeacherRejectAppeal)
		teacher.PATCH("/appeals/:appealId/resolve", teacherHandler.TeacherResolveAppeal)
		teacher.DELETE("/questions/:question_id", teacherHandler.TeacherDeleteQuestion)
	}

	student := api.Group("/student")
	student.Use(middleware.AuthRequiredWithSecret(cfg.Auth.JWTSecret), middleware.RequireRole(models.RoleStudent))
	{
		student.GET("/exams/lookup", studentHandler.StudentLookupExam)
		student.GET("/history", studentHandler.StudentGetHistory)
		student.GET("/history/:submissionId", studentHandler.StudentGetHistoryDetail)
		student.POST("/appeals", studentHandler.StudentCreateAppeal)
		student.GET("/attempts", studentHandler.StudentListAttempts)
		student.POST("/exams/:id/start", studentHandler.StudentStartExam)
		student.POST("/attempts/:attempt_id/save", studentHandler.StudentSaveAnswers)
		student.POST("/attempts/:attempt_id/submit", studentHandler.StudentSubmitExam)
		student.GET("/attempts/:attempt_id/result", studentHandler.StudentGetResult)
	}

	return r
}
