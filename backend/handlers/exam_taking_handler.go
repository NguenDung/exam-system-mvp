package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"exam-system-mvp/config"
	"exam-system-mvp/dto"
	"exam-system-mvp/repository"
	"exam-system-mvp/service"

	"github.com/gin-gonic/gin"
)

type ExamTakingHandler struct {
	takingService service.ExamTakingService
}

func NewExamTakingHandler(takingService service.ExamTakingService) *ExamTakingHandler {
	return &ExamTakingHandler{takingService: takingService}
}

func getDefaultExamTakingHandler() *ExamTakingHandler {
	examRepo := repository.NewExamRepository(config.DB)
	classRepo := repository.NewClassroomRepository(config.DB)
	userRepo := repository.NewUserRepository(config.DB)
	svc := service.NewExamTakingService(examRepo, classRepo, userRepo)
	return NewExamTakingHandler(svc)
}

func (h *ExamTakingHandler) StudentGetAvailableExams(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	exams, err := h.takingService.GetAvailableExams(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load exams"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exams": exams})
}

func (h *ExamTakingHandler) StudentLookupExam(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'code' is required"})
		return
	}

	userIDVal, exists := c.Get("user_id")
	var userID uint
	if exists {
		userID = userIDVal.(uint)
	}

	res, err := h.takingService.LookupExam(userID, code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *ExamTakingHandler) StudentStartExam(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var input struct {
		Confirmed bool `json:"confirmed"`
	}
	_ = c.ShouldBindJSON(&input)

	res, err := h.takingService.StartExam(userID, examID, input.Confirmed)
	if err != nil {
		if errors.Is(err, service.ErrParticipationNotConfirmed) {
			c.JSON(http.StatusPreconditionRequired, gin.H{"error": err.Error()})
			return
		}
		errMsg := err.Error()
		if errMsg == "Exam not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else if errMsg == "You have already submitted this exam" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		} else if errMsg == "You are not enrolled in the classroom for this exam" {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		}
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *ExamTakingHandler) StudentSaveAnswers(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	attemptID, ok := parseUintParam(c, "attempt_id")
	if !ok {
		return
	}

	var req dto.SaveAnswersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.takingService.SaveAnswers(userID, attemptID, req)
	if err != nil {
		if err.Error() == "Submission window closed" {
			c.JSON(http.StatusConflict, gin.H{
				"error":          "Time is up. Your exam has been automatically submitted.",
				"auto_submitted": true,
			})
		} else if err.Error() == "Attempt not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Answers saved successfully",
		"saved_at": time.Now(),
	})
}

func (h *ExamTakingHandler) StudentRecordViolation(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	attemptID, ok := parseUintParam(c, "attempt_id")
	if !ok {
		return
	}

	var req dto.ViolationRequest
	_ = c.ShouldBindJSON(&req)

	res, err := h.takingService.RecordViolation(userID, attemptID, req)
	if err != nil {
		if err.Error() == "Attempt not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *ExamTakingHandler) StudentSubmitExam(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	attemptID, ok := parseUintParam(c, "attempt_id")
	if !ok {
		return
	}

	var req dto.SubmitExamRequest
	_ = c.ShouldBindJSON(&req)

	attempt, err := h.takingService.SubmitExam(userID, attemptID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Exam submitted successfully",
		"attempt": attempt,
	})
}

func (h *ExamTakingHandler) StudentGetResult(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	attemptID, ok := parseUintParam(c, "attempt_id")
	if !ok {
		return
	}

	res, err := h.takingService.GetResult(userID, attemptID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *ExamTakingHandler) StudentListAttempts(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	attempts, err := h.takingService.ListAttempts(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load attempts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"attempts": attempts})
}

func (h *ExamTakingHandler) GetExamSchedule(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	roleVal, _ := c.Get("role")
	role := roleVal.(string)

	res, err := h.takingService.GetExamSchedule(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load schedule"})
		return
	}

	c.JSON(http.StatusOK, res)
}

func StartAttemptAutoSubmitWorker() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if config.DB != nil {
			svc := getDefaultExamTakingHandler().takingService
			if err := svc.AutoSubmitExpiredAttempts(); err != nil {
				log.Printf("Auto-submit worker error: %v", err)
			}
		}
	}
}
