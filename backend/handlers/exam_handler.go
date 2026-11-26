package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"exam-system-mvp/config"
	"exam-system-mvp/dto"
	"exam-system-mvp/repository"
	"exam-system-mvp/service"

	"github.com/gin-gonic/gin"
)

type ExamHandler struct {
	examService service.ExamService
}

func NewExamHandler(examService service.ExamService) *ExamHandler {
	return &ExamHandler{examService: examService}
}

func getDefaultExamHandler() *ExamHandler {
	examRepo := repository.NewExamRepository(config.DB)
	userRepo := repository.NewUserRepository(config.DB)
	notifRepo := repository.NewNotificationRepository(config.DB)
	notifSvc := service.NewNotificationService(notifRepo)
	svc := service.NewExamService(examRepo, userRepo, notifSvc)
	return NewExamHandler(svc)
}

// TeacherCreateExam godoc
// @Summary Giảng viên tạo đề thi mới
// @Tags Teacher Exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateExamRequest true "Thông tin đề thi"
// @Success 201 {object} dto.ExamEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/teacher/exams [post]
func (h *ExamHandler) TeacherCreateExam(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	var req dto.CreateExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exam, err := h.examService.CreateExam(teacherID, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTimeRange) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exam"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Exam created successfully",
		"exam":    exam,
	})
}

// TeacherListExams godoc
// @Summary Danh sách đề thi của giảng viên
// @Tags Teacher Exams
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ExamsListResponse
// @Router /api/teacher/exams [get]
func (h *ExamHandler) TeacherListExams(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	exams, err := h.examService.ListExams(teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load exams"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exams": exams})
}

// TeacherGetExam godoc
// @Summary Xem chi tiết đề thi và danh sách câu hỏi
// @Tags Teacher Exams
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID đề thi"
// @Success 200 {object} dto.ExamEnvelope
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/exams/{id} [get]
func (h *ExamHandler) TeacherGetExam(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	exam, err := h.examService.GetExam(teacherID, examID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exam": exam})
}

// TeacherUpdateExam godoc
// @Summary Cập nhật thông tin đề thi
// @Tags Teacher Exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID đề thi"
// @Param request body dto.UpdateExamRequest true "Thông tin cập nhật"
// @Success 200 {object} dto.ExamEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/exams/{id} [put]
func (h *ExamHandler) TeacherUpdateExam(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exam, err := h.examService.UpdateExam(teacherID, examID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Exam updated successfully",
		"exam":    exam,
	})
}

// TeacherDeleteExam godoc
// @Summary Xóa đề thi và dữ liệu liên quan
// @Tags Teacher Exams
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID đề thi"
// @Success 200 {object} dto.MessageResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/exams/{id} [delete]
func (h *ExamHandler) TeacherDeleteExam(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := h.examService.DeleteExam(teacherID, examID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exam deleted successfully"})
}

// TeacherPublishExam godoc
// @Summary Xuất bản đề thi cho sinh viên làm bài
// @Tags Teacher Exams
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID đề thi"
// @Success 200 {object} dto.ExamEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/exams/{id}/publish [patch]
func (h *ExamHandler) TeacherPublishExam(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	exam, err := h.examService.PublishExam(teacherID, examID)
	if err != nil {
		if errors.Is(err, service.ErrNoQuestionsInExam) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Exam published successfully",
		"exam":    exam,
	})
}

// TeacherAddQuestions godoc
// @Summary Thêm câu hỏi vào đề thi
// @Tags Teacher Questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID đề thi"
// @Param request body dto.AddQuestionsRequest true "Danh sách câu hỏi"
// @Success 201 {object} dto.QuestionsListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/teacher/exams/{id}/questions [post]
func (h *ExamHandler) TeacherAddQuestions(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.AddQuestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdQuestions, err := h.examService.AddQuestions(teacherID, examID, req.Questions)
	if err != nil {
		if errors.Is(err, service.ErrExamNotDraft) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add questions"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Questions added successfully",
		"questions": createdQuestions,
	})
}

// TeacherGetResults godoc
// @Summary Xem bảng điểm và thống kê kết quả đề thi
// @Tags Teacher Results & Audit
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID đề thi"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/exams/{id}/results [get]
func (h *ExamHandler) TeacherGetResults(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	res, err := h.examService.GetResults(teacherID, examID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	c.JSON(http.StatusOK, res)
}

// TeacherExportResultsCSV godoc
// @Summary Xuất bảng điểm đề thi ra file CSV UTF-8 BOM
// @Tags Teacher Results & Audit
// @Produce text/csv
// @Security BearerAuth
// @Param id path int true "ID đề thi"
// @Success 200 {file} file
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/exams/{id}/results/export [get]
func (h *ExamHandler) TeacherExportResultsCSV(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	bytesData, fileName, err := h.examService.ExportResultsCSV(teacherID, examID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", bytesData)
}

// TeacherUpdateAttemptScore godoc
// @Summary Điều chỉnh điểm bài làm kèm ghi chú và audit log
// @Tags Teacher Results & Audit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID lượt thi"
// @Param request body dto.UpdateScoreRequest true "Điểm mới và ghi chú"
// @Success 200 {object} dto.AttemptEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/attempts/{id}/score [patch]
func (h *ExamHandler) TeacherUpdateAttemptScore(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	attemptID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	attempt, err := h.examService.UpdateAttemptScore(teacherID, attemptID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Attempt score updated successfully",
		"attempt": attempt,
	})
}

// TeacherResetAttempt godoc
// @Summary Xóa lượt làm bài để cho phép sinh viên thi lại
// @Tags Teacher Results & Audit
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID lượt thi"
// @Success 200 {object} dto.MessageResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/attempts/{id} [delete]
func (h *ExamHandler) TeacherResetAttempt(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	attemptID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := h.examService.ResetAttempt(teacherID, attemptID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Attempt reset successfully"})
}

// TeacherBulkAddQuestions godoc
// @Summary Nhập câu hỏi hàng loạt trong một transaction
// @Tags Teacher Questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID đề thi"
// @Param request body dto.AddQuestionsRequest true "Danh sách câu hỏi"
// @Success 201 {object} dto.QuestionsListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/teacher/exams/{id}/questions/bulk [post]
func (h *ExamHandler) TeacherBulkAddQuestions(c *gin.Context) {
	h.TeacherAddQuestions(c)
}

// TeacherUpdateQuestion godoc
// @Summary Cập nhật câu hỏi trong đề nháp
// @Tags Teacher Questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID câu hỏi"
// @Param request body dto.QuestionDTO true "Thông tin câu hỏi"
// @Success 200 {object} dto.QuestionEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/questions/{id} [put]
func (h *ExamHandler) TeacherUpdateQuestion(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	questionID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.QuestionDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	q, err := h.examService.UpdateQuestion(teacherID, questionID, req)
	if err != nil {
		if errors.Is(err, service.ErrExamNotDraft) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Question updated successfully",
		"question": q,
	})
}

// TeacherDeleteQuestion godoc
// @Summary Xóa câu hỏi khỏi đề nháp
// @Tags Teacher Questions
// @Produce json
// @Security BearerAuth
// @Param question_id path int true "ID câu hỏi"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/questions/{question_id} [delete]
func (h *ExamHandler) TeacherDeleteQuestion(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	questionID, ok := parseUintParam(c, "question_id")
	if !ok {
		questionID, ok = parseUintParam(c, "id")
		if !ok {
			return
		}
	}

	if err := h.examService.DeleteQuestion(teacherID, questionID); err != nil {
		if errors.Is(err, service.ErrExamNotDraft) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Question deleted successfully"})
}

// TeacherListStudents godoc
// @Summary Danh sách sinh viên (hỗ trợ tìm kiếm ?q=)
// @Tags Teacher Users
// @Produce json
// @Security BearerAuth
// @Param q query string false "Từ khóa tìm kiếm"
// @Success 200 {object} map[string]interface{}
// @Router /api/teacher/students [get]
func (h *ExamHandler) TeacherListStudents(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	q := c.Query("q")
	students, err := h.examService.ListStudents(teacherID, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load students"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"students": students})
}

// TeacherCreateStudent godoc
// @Summary Giảng viên tạo tài khoản sinh viên mới
// @Tags Teacher Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateStudentRequest true "Thông tin sinh viên"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/teacher/students [post]
func (h *ExamHandler) TeacherCreateStudent(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	var req dto.CreateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	student, err := h.examService.CreateStudent(teacherID, req)
	if err != nil {
		if err.Error() == "Email already registered" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Student created successfully",
		"student": student,
	})
}

// TeacherUpdateStudent godoc
// @Summary Cập nhật thông tin hoặc đặt lại mật khẩu sinh viên
// @Tags Teacher Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID sinh viên"
// @Param request body dto.UpdateStudentRequest true "Thông tin cập nhật"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/teacher/students/{id} [put]
func (h *ExamHandler) TeacherUpdateStudent(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	studentID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	student, err := h.examService.UpdateStudent(teacherID, studentID, req)
	if err != nil {
		if err.Error() == "Email already in use" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Student updated successfully",
		"student": student,
	})
}

// TeacherDeleteStudent godoc
// @Summary Xóa tài khoản sinh viên
// @Tags Teacher Users
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID sinh viên"
// @Success 200 {object} dto.MessageResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/students/{id} [delete]
func (h *ExamHandler) TeacherDeleteStudent(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	studentID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := h.examService.DeleteStudent(teacherID, studentID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student deleted successfully"})
}

// TeacherGetAuditLogs godoc
// @Summary Xem nhật ký kiểm toán các thao tác điểm số và phòng thi
// @Tags Teacher Results & Audit
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.AuditLogsResponse
// @Router /api/teacher/audit-logs [get]
func (h *ExamHandler) TeacherGetAuditLogs(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	logs, err := h.examService.GetAuditLogs(teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"logs": logs,
	})
}
