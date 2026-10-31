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
