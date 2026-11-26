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

// StudentGetAvailableExams godoc
// @Summary Danh sách đề thi khả dụng cho sinh viên
// @Tags Student Exam
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/student/exams/available [get]
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

// StudentLookupExam godoc
// @Summary Tra cứu đề thi bằng mã đề
// @Tags Student Exam
// @Produce json
// @Security BearerAuth
// @Param code query string true "Mã đề thi (VD: IT2026)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 429 {object} dto.ErrorResponse
// @Router /api/student/exams/lookup [get]
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

// StudentStartExam godoc
// @Summary Bắt đầu hoặc tiếp tục làm bài thi
// @Tags Student Exam
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID đề thi"
// @Param request body dto.StartExamRequest false "Xác nhận tham gia thi"
// @Success 200 {object} dto.StartExamResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 428 {object} dto.ErrorResponse
// @Router /api/student/exams/{id}/start [post]
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

// StudentSaveAnswers godoc
// @Summary Tự động lưu đáp án (Atomic UPSERT)
// @Tags Student Exam
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param attempt_id path int true "ID lượt thi"
// @Param request body dto.SaveAnswersRequest true "Danh sách đáp án"
// @Success 200 {object} dto.SaveAnswersResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/student/attempts/{attempt_id}/save [post]
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

// StudentRecordViolation godoc
// @Summary Ghi nhận vi phạm chuyển tab (tự động thu bài khi đủ 3 lần)
// @Tags Anti Cheat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param attempt_id path int true "ID lượt thi"
// @Param request body dto.ViolationRequest false "Lý do vi phạm"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/student/attempts/{attempt_id}/violation [post]
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

// StudentSubmitExam godoc
// @Summary Nộp bài thi chính thức và chấm điểm tự động
// @Tags Student Exam
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param attempt_id path int true "ID lượt thi"
// @Param request body dto.SubmitExamRequest false "Đáp án cuối (tùy chọn)"
// @Success 200 {object} dto.AttemptEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/student/attempts/{attempt_id}/submit [post]
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

// StudentGetResult godoc
// @Summary Xem kết quả bài thi sau khi nộp
// @Tags Student Exam
// @Produce json
// @Security BearerAuth
// @Param attempt_id path int true "ID lượt thi"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/student/attempts/{attempt_id}/result [get]
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

// StudentListAttempts godoc
// @Summary Danh sách các lượt thi của sinh viên
// @Tags Student Exam
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/student/attempts [get]
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

// GetExamSchedule godoc
// @Summary Lấy thời khóa biểu & lịch thi phân nhóm theo thời gian
// @Tags Schedule
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/exams/schedule [get]
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
