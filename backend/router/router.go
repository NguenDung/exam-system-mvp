package router

import (
	"time"

	"exam-system-mvp/handlers"
	"exam-system-mvp/middleware"
	"exam-system-mvp/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(middleware.SecurityMiddleware(2 * 1024 * 1024))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "exam-system-mvp"})
	})

	api := r.Group("/api")

	auth := api.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", middleware.IPRateLimit(15, time.Minute), handlers.Login)
	}

	teacher := api.Group("/teacher")
	teacher.Use(middleware.AuthRequired(), middleware.RequireRole(models.RoleTeacher))
	{
		teacher.GET("/exams", handlers.TeacherListExams)
		teacher.POST("/exams", handlers.TeacherCreateExam)
		teacher.GET("/exams/:id", handlers.TeacherGetExam)
		teacher.PUT("/exams/:id", handlers.TeacherUpdateExam)
		teacher.DELETE("/exams/:id", handlers.TeacherDeleteExam)
		teacher.POST("/exams/:id/questions", handlers.TeacherAddQuestions)
		teacher.POST("/exams/:id/questions/bulk", handlers.TeacherBulkAddQuestions)
		teacher.PATCH("/exams/:id/publish", handlers.TeacherPublishExam)
		teacher.GET("/exams/:id/results", handlers.TeacherGetResults)
		teacher.GET("/exams/:id/results/export", handlers.TeacherExportResultsCSV)
		teacher.GET("/exams/:id/appeals", handlers.TeacherListExamAppeals)
		teacher.GET("/appeals/:appealId", handlers.TeacherGetAppeal)
		teacher.PATCH("/appeals/:appealId/reply", handlers.TeacherReplyAppeal)
		teacher.PATCH("/appeals/:appealId/reject", handlers.TeacherRejectAppeal)
		teacher.PATCH("/appeals/:appealId/resolve", handlers.TeacherResolveAppeal)
		teacher.PUT("/questions/:id", handlers.TeacherUpdateQuestion)
		teacher.DELETE("/questions/:question_id", handlers.TeacherDeleteQuestion)
		teacher.PATCH("/attempts/:id/score", handlers.TeacherUpdateAttemptScore)
		teacher.DELETE("/attempts/:id", handlers.TeacherResetAttempt)
		teacher.GET("/audit-logs", handlers.TeacherGetAuditLogs)
		teacher.GET("/students", handlers.TeacherListStudents)
		teacher.POST("/students", handlers.TeacherCreateStudent)
		teacher.PUT("/students/:id", handlers.TeacherUpdateStudent)
		teacher.DELETE("/students/:id", handlers.TeacherDeleteStudent)
		teacher.PATCH("/users/:id/role", handlers.TeacherUpdateUserRole)
		teacher.POST("/classes", handlers.TeacherCreateClass)
		teacher.PUT("/classes/:id", handlers.TeacherUpdateClass)
		teacher.DELETE("/classes/:id", handlers.TeacherDeleteClass)
		teacher.POST("/classes/:id/members", handlers.TeacherAddClassMember)
		teacher.DELETE("/classes/:id/members/:studentId", handlers.TeacherRemoveClassMember)
	}

	student := api.Group("/student")
	student.Use(middleware.AuthRequired(), middleware.RequireRole(models.RoleStudent))
	{
		student.GET("/exams/available", handlers.StudentGetAvailableExams)
		student.GET("/exams/lookup", middleware.IPRateLimit(15, time.Minute), handlers.StudentLookupExam)
		student.GET("/history", handlers.StudentGetHistory)
		student.GET("/history/:submissionId", handlers.StudentGetHistoryDetail)
		student.POST("/appeals", handlers.StudentCreateAppeal)
		student.GET("/attempts", handlers.StudentListAttempts)
		student.POST("/exams/:id/start", handlers.StudentStartExam)
		student.POST("/attempts/:attempt_id/save", handlers.StudentSaveAnswers)
		student.POST("/attempts/:attempt_id/submit", handlers.StudentSubmitExam)
		student.POST("/attempts/:attempt_id/violation", handlers.StudentRecordViolation)
		student.GET("/attempts/:attempt_id/result", handlers.StudentGetResult)
		student.POST("/classes/join", handlers.StudentJoinClass)
	}

	common := api.Group("")
	common.Use(middleware.AuthRequired())
	{
		common.GET("/profile", handlers.GetProfile)
		common.PUT("/profile", handlers.UpdateProfile)
		common.GET("/users/directory", handlers.GetDirectory)
		common.GET("/classes", handlers.ListClasses)
		common.GET("/classes/:id", handlers.GetClassDetail)
		common.GET("/exams/schedule", handlers.GetExamSchedule)
	}

	notifications := api.Group("/notifications")
	notifications.Use(middleware.AuthRequired())
	{
		notifications.GET("", handlers.GetNotifications)
		notifications.PATCH("/:id/read", handlers.MarkNotificationAsRead)
		notifications.PATCH("/read-all", handlers.MarkAllNotificationsAsRead)
		notifications.DELETE("/clear-read", handlers.ClearReadNotifications)
		notifications.DELETE("/:id", handlers.DeleteNotification)
	}

	return r
}
