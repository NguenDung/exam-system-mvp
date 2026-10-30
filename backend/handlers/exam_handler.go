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

func (h *ExamHandler) TeacherBulkAddQuestions(c *gin.Context) {
	h.TeacherAddQuestions(c)
}

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
