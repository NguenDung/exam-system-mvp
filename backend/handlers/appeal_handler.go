package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"exam-system-mvp/config"
	"exam-system-mvp/dto"
	"exam-system-mvp/models"
	"exam-system-mvp/repository"
	"exam-system-mvp/service"

	"github.com/gin-gonic/gin"
)

type AppealHandler struct {
	appealService service.AppealService
}

func NewAppealHandler(svc service.AppealService) *AppealHandler {
	return &AppealHandler{appealService: svc}
}

func getDefaultAppealHandler() *AppealHandler {
	appealRepo := repository.NewAppealRepository(config.DB)
	examRepo := repository.NewExamRepository(config.DB)
	notifRepo := repository.NewNotificationRepository(config.DB)
	notifSvc := service.NewNotificationService(notifRepo)
	svc := service.NewAppealService(appealRepo, examRepo, notifSvc)
	return NewAppealHandler(svc)
}

// StudentGetHistory godoc
// @Summary Lịch sử các bài thi đã làm của sinh viên
// @Tags Student Appeals & History
// @Produce json
// @Security BearerAuth
// @Param page query int false "Trang (mặc định 1)"
// @Param page_size query int false "Số dòng mỗi trang (mặc định 10)"
// @Param q query string false "Từ khóa tìm kiếm"
// @Param appeal_status query string false "Trạng thái phúc khảo (PENDING|REPLIED|REJECTED|RESOLVED)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/student/history [get]
func (h *AppealHandler) StudentGetHistory(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	studentID := userIDVal.(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	search := c.Query("q")
	status := strings.TrimSpace(c.Query("appeal_status"))
	if status != "" {
		if status != models.AppealStatusPending && status != models.AppealStatusReplied &&
			status != models.AppealStatusRejected && status != models.AppealStatusResolved {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appeal_status"})
			return
		}
	}

	items, total, err := h.appealService.GetHistory(studentID, search, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load exam history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":   items,
		"history": items,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// StudentGetHistoryDetail godoc
// @Summary Xem chi tiết bài làm trong lịch sử (chỉ hiện đáp án đúng sau khi đóng ca thi)
// @Tags Student Appeals & History
// @Produce json
// @Security BearerAuth
// @Param submissionId path int true "ID lượt thi"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/student/history/{submissionId} [get]
func (h *AppealHandler) StudentGetHistoryDetail(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	studentID := userIDVal.(uint)

	submissionID, ok := parseUintParam(c, "submissionId")
	if !ok {
		return
	}

	res, err := h.appealService.GetHistoryDetail(submissionID, studentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	c.JSON(http.StatusOK, res)
}

// StudentCreateAppeal godoc
// @Summary Gửi đơn phúc khảo điểm bài thi
// @Tags Student Appeals & History
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateAppealRequest true "Nội dung đơn phúc khảo"
// @Success 201 {object} dto.AppealEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/student/appeals [post]
func (h *AppealHandler) StudentCreateAppeal(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	studentID := userIDVal.(uint)

	var req dto.CreateAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.SubmissionID == 0 && req.SubmissionId != 0 {
		req.SubmissionID = req.SubmissionId
	}
	if req.SubmissionID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "submissionId is required"})
		return
	}

	appeal, err := h.appealService.CreateAppeal(studentID, req)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "Bạn đã có khiếu nại đang chờ xử lý cho bài thi này." || errMsg == "An appeal for this submission already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": errMsg})
		} else if errMsg == "Submission not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Appeal submitted successfully",
		"appeal":  appeal,
	})
}

// TeacherListExamAppeals godoc
// @Summary Danh sách đơn phúc khảo của một đề thi
// @Tags Teacher Appeals
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID đề thi"
// @Param page query int false "Trang"
// @Param page_size query int false "Số dòng mỗi trang"
// @Param status query string false "Lọc trạng thái"
// @Param q query string false "Từ khóa tìm kiếm"
// @Success 200 {object} dto.TeacherAppealsResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/exams/{id}/appeals [get]
func (h *AppealHandler) TeacherListExamAppeals(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	status := c.Query("status")
	query := c.Query("q")

	appeals, total, err := h.appealService.ListExamAppeals(examID, teacherID, status, query, page, pageSize)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":   appeals,
		"appeals": appeals,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// TeacherGetAppeal godoc
// @Summary Xem chi tiết đơn phúc khảo và bài làm liên quan
// @Tags Teacher Appeals
// @Produce json
// @Security BearerAuth
// @Param appealId path int true "ID đơn phúc khảo"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/appeals/{appealId} [get]
func (h *AppealHandler) TeacherGetAppeal(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	appealID, ok := parseUintParam(c, "appealId")
	if !ok {
		return
	}

	detail, err := h.appealService.GetAppealDetail(appealID, teacherID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appeal not found"})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// TeacherReplyAppeal godoc
// @Summary Giảng viên phản hồi đơn phúc khảo
// @Tags Teacher Appeals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param appealId path int true "ID đơn phúc khảo"
// @Param request body dto.ReplyAppealRequest true "Nội dung phản hồi"
// @Success 200 {object} dto.AppealEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/teacher/appeals/{appealId}/reply [patch]
func (h *AppealHandler) TeacherReplyAppeal(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	appealID, ok := parseUintParam(c, "appealId")
	if !ok {
		return
	}

	var req dto.ReplyAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appeal, err := h.appealService.ReplyAppeal(appealID, teacherID, req)
	if err != nil {
		if errors.Is(err, service.ErrAppealClosed) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appeal not found"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Response submitted",
		"appeal":  appeal,
	})
}

// TeacherRejectAppeal godoc
// @Summary Giảng viên từ chối đơn phúc khảo
// @Tags Teacher Appeals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param appealId path int true "ID đơn phúc khảo"
// @Param request body dto.ActionAppealRequest false "Lý do từ chối (tùy chọn)"
// @Success 200 {object} dto.AppealEnvelope
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/teacher/appeals/{appealId}/reject [patch]
func (h *AppealHandler) TeacherRejectAppeal(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	appealID, ok := parseUintParam(c, "appealId")
	if !ok {
		return
	}

	var req dto.ActionAppealRequest
	_ = c.ShouldBindJSON(&req)

	appeal, err := h.appealService.RejectAppeal(appealID, teacherID, req)
	if err != nil {
		if errors.Is(err, service.ErrAppealClosed) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appeal not found"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Appeal rejected",
		"appeal":  appeal,
	})
}

// TeacherResolveAppeal godoc
// @Summary Giảng viên chấp thuận đơn phúc khảo
// @Tags Teacher Appeals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param appealId path int true "ID đơn phúc khảo"
// @Param request body dto.ActionAppealRequest false "Ghi chú xử lý (tùy chọn)"
// @Success 200 {object} dto.AppealEnvelope
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/teacher/appeals/{appealId}/resolve [patch]
func (h *AppealHandler) TeacherResolveAppeal(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	appealID, ok := parseUintParam(c, "appealId")
	if !ok {
		return
	}

	var req dto.ActionAppealRequest
	_ = c.ShouldBindJSON(&req)

	appeal, err := h.appealService.ResolveAppeal(appealID, teacherID, req)
	if err != nil {
		if errors.Is(err, service.ErrAppealClosed) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appeal not found"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Appeal resolved",
		"appeal":  appeal,
	})
}
