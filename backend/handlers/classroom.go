package handlers

import (
	"errors"
	"net/http"

	"exam-system-mvp/config"
	"exam-system-mvp/dto"
	"exam-system-mvp/repository"
	"exam-system-mvp/service"

	"github.com/gin-gonic/gin"
)

type ClassroomHandler struct {
	service service.ClassroomService
}

func NewClassroomHandler(svc service.ClassroomService) *ClassroomHandler {
	return &ClassroomHandler{service: svc}
}

func getDefaultClassroomHandler() *ClassroomHandler {
	classRepo := repository.NewClassroomRepository(config.DB)
	userRepo := repository.NewUserRepository(config.DB)
	notifRepo := repository.NewNotificationRepository(config.DB)
	notifSvc := service.NewNotificationService(notifRepo)
	svc := service.NewClassroomService(classRepo, userRepo, notifSvc)
	return NewClassroomHandler(svc)
}

// TeacherCreateClass godoc
// @Summary Giảng viên tạo lớp học phần mới
// @Tags Teacher Classrooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateClassRequest true "Thông tin lớp học phần"
// @Success 201 {object} dto.ClassEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/teacher/classes [post]
func (h *ClassroomHandler) TeacherCreateClass(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	var req dto.CreateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	class, err := h.service.CreateClass(teacherID, req)
	if err != nil {
		if errors.Is(err, service.ErrEmptyClassName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create classroom"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Class created successfully",
		"class":   class,
	})
}

// TeacherUpdateClass godoc
// @Summary Cập nhật thông tin lớp học phần
// @Tags Teacher Classrooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID lớp học phần"
// @Param request body dto.UpdateClassRequest true "Thông tin cập nhật"
// @Success 200 {object} dto.ClassEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/classes/{id} [put]
func (h *ClassroomHandler) TeacherUpdateClass(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	classID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	class, err := h.service.UpdateClass(teacherID, classID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyClassName):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrClassroomNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update classroom"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Class updated successfully",
		"class":   class,
	})
}

// TeacherDeleteClass godoc
// @Summary Xóa lớp học phần
// @Tags Teacher Classrooms
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID lớp học phần"
// @Success 200 {object} dto.MessageResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/classes/{id} [delete]
func (h *ClassroomHandler) TeacherDeleteClass(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	classID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	err := h.service.DeleteClass(teacherID, classID)
	if err != nil {
		if errors.Is(err, service.ErrClassroomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete classroom"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Classroom deleted successfully"})
}

// TeacherAddClassMember godoc
// @Summary Thêm sinh viên vào lớp học phần
// @Tags Teacher Classrooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID lớp học phần"
// @Param request body dto.AddClassMemberRequest true "Thông tin sinh viên (student_id, email hoặc student_code)"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/teacher/classes/{id}/members [post]
func (h *ClassroomHandler) TeacherAddClassMember(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	classID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.AddClassMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.service.AddMember(teacherID, classID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMissingStudentParam):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrStudentNotFound), errors.Is(err, service.ErrClassroomNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrStudentAlreadyInClass):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member to classroom"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Student added to class successfully",
		"member": gin.H{
			"id":           member.ID,
			"classroom_id": classID,
			"student_id":   member.StudentID,
			"joined_at":    member.JoinedAt,
			"student": gin.H{
				"id":           member.StudentID,
				"full_name":    member.FullName,
				"email":        member.Email,
				"student_code": member.StudentCode,
				"department":   member.Department,
				"avatar_url":   member.AvatarURL,
			},
		},
	})
}

// TeacherRemoveClassMember godoc
// @Summary Xóa sinh viên khỏi lớp học phần
// @Tags Teacher Classrooms
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID lớp học phần"
// @Param studentId path int true "ID sinh viên"
// @Success 200 {object} dto.MessageResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/classes/{id}/members/{studentId} [delete]
func (h *ClassroomHandler) TeacherRemoveClassMember(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	classID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	studentID, ok := parseUintParam(c, "studentId")
	if !ok {
		return
	}

	err := h.service.RemoveMember(teacherID, classID, studentID)
	if err != nil {
		if errors.Is(err, service.ErrClassroomNotFound) || errors.Is(err, service.ErrStudentNotInClass) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed from class successfully"})
}

// StudentJoinClass godoc
// @Summary Sinh viên tham gia lớp học phần bằng mã lớp
// @Tags Classrooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.JoinClassRequest true "Mã lớp học phần"
// @Success 200 {object} dto.ClassEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/student/classes/join [post]
func (h *ClassroomHandler) StudentJoinClass(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	var req dto.JoinClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	class, err := h.service.JoinClass(userID, req)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "Class code is required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		} else if errMsg == "Classroom not found with this code" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else if errMsg == "You have already joined this classroom" {
			c.JSON(http.StatusConflict, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join classroom"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Joined classroom successfully",
		"class":   class,
	})
}

// ListClasses godoc
// @Summary Danh sách lớp học phần của người dùng hiện tại
// @Tags Classrooms
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ClassesListResponse
// @Router /api/classes [get]
func (h *ClassroomHandler) ListClasses(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)
	roleVal, _ := c.Get("role")
	role := roleVal.(string)

	items, err := h.service.ListClasses(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch classes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"classes": items})
}

// GetClassDetail godoc
// @Summary Xem chi tiết lớp học phần
// @Tags Classrooms
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID lớp học phần"
// @Success 200 {object} dto.ClassDetailResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/classes/{id} [get]
func (h *ClassroomHandler) GetClassDetail(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)
	roleVal, _ := c.Get("role")
	role := roleVal.(string)

	classID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	res, err := h.service.GetClassDetail(userID, role, classID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrClassroomNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		case errors.Is(err, service.ErrClassroomAccessDenied), errors.Is(err, service.ErrStudentNotEnrolled):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load class detail"})
		}
		return
	}

	c.JSON(http.StatusOK, res)
}

// TeacherGetClassGradebook godoc
// @Summary Lấy sổ điểm tổng hợp của lớp học phần
// @Tags Teacher Classrooms
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID lớp học phần"
// @Success 200 {object} dto.ClassGradebookResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/teacher/classes/{id}/gradebook [get]
func (h *ClassroomHandler) TeacherGetClassGradebook(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	classID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	gradebook, err := h.service.GetClassGradebook(teacherID, classID)
	if err != nil {
		if errors.Is(err, service.ErrClassroomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found or permission denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load class gradebook"})
		return
	}

	c.JSON(http.StatusOK, gradebook)
}
