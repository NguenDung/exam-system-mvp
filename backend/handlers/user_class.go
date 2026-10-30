package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"exam-system-mvp/config"
	"exam-system-mvp/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func generateClassCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	for attempt := 0; attempt < 20; attempt++ {
		b := make([]byte, 6)
		for i := range b {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				b[i] = charset[i%len(charset)]
			} else {
				b[i] = charset[n.Int64()]
			}
		}
		code := string(b)
		var count int64
		if err := config.DB.Model(&models.Classroom{}).Where("code = ?", code).Count(&count).Error; err != nil {
			return code
		}
		if count == 0 {
			return code
		}
	}
	return fmt.Sprintf("C%05d", time.Now().UnixNano()%100000)
}

type UpdateProfileInput struct {
	FullName    *string `json:"full_name"`
	AvatarURL   *string `json:"avatar_url"`
	Bio         *string `json:"bio"`
	Department  *string `json:"department"`
	StudentCode *string `json:"student_code"`
	Password    *string `json:"password"`
}

func GetProfile(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":           user.ID,
			"full_name":    user.FullName,
			"email":        user.Email,
			"role":         user.Role,
			"avatar_url":   user.AvatarURL,
			"bio":          user.Bio,
			"department":   user.Department,
			"student_code": user.StudentCode,
			"created_at":   user.CreatedAt,
		},
	})
}

func UpdateProfile(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var input UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}

	if input.FullName != nil {
		name := strings.TrimSpace(*input.FullName)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Full name cannot be empty"})
			return
		}
		updates["full_name"] = name
	}
	if input.AvatarURL != nil {
		updates["avatar_url"] = *input.AvatarURL
	}
	if input.Bio != nil {
		updates["bio"] = *input.Bio
	}
	if input.Department != nil {
		updates["department"] = strings.TrimSpace(*input.Department)
	}
	if input.StudentCode != nil {
		updates["student_code"] = strings.TrimSpace(*input.StudentCode)
	}
	if input.Password != nil && strings.TrimSpace(*input.Password) != "" {
		pwd := *input.Password
		if len(pwd) < 6 || len(pwd) > 72 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be between 6 and 72 characters"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		updates["password_hash"] = string(hash)
	}

	if len(updates) > 0 {
		if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}
	}

	config.DB.First(&user, userID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":           user.ID,
			"full_name":    user.FullName,
			"email":        user.Email,
			"role":         user.Role,
			"avatar_url":   user.AvatarURL,
			"bio":          user.Bio,
			"department":   user.Department,
			"student_code": user.StudentCode,
			"created_at":   user.CreatedAt,
		},
	})
}

type DirectoryUserItem struct {
	ID          uint      `json:"id"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	AvatarURL   string    `json:"avatar_url"`
	Bio         string    `json:"bio"`
	Department  string    `json:"department"`
	StudentCode string    `json:"student_code"`
	CreatedAt   time.Time `json:"created_at"`
}

func GetDirectory(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	role := strings.TrimSpace(c.Query("role"))
	dept := strings.TrimSpace(c.Query("department"))

	db := config.DB.Model(&models.User{})

	if role == models.RoleTeacher || role == models.RoleStudent {
		db = db.Where("role = ?", role)
	}
	if dept != "" {
		db = db.Where("LOWER(department) LIKE ?", "%"+strings.ToLower(dept)+"%")
	}
	if q != "" {
		term := "%" + strings.ToLower(q) + "%"
		db = db.Where("LOWER(full_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(student_code) LIKE ? OR LOWER(department) LIKE ?", term, term, term, term)
	}

	var users []DirectoryUserItem
	if err := db.Select("id, full_name, email, role, avatar_url, bio, department, student_code, created_at").
		Order("full_name ASC").
		Limit(100).
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch directory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

type UpdateUserRoleInput struct {
	Role string `json:"role" binding:"required"`
}

func TeacherUpdateUserRole(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	targetID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if targetID == teacherID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot change your own role"})
		return
	}

	var input UpdateUserRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newRole := strings.ToLower(strings.TrimSpace(input.Role))
	if newRole != models.RoleTeacher && newRole != models.RoleStudent {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role must be either 'teacher' or 'student'"})
		return
	}

	var targetUser models.User
	if err := config.DB.First(&targetUser, targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := config.DB.Model(&targetUser).Update("role", newRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
		"user": gin.H{
			"id":        targetUser.ID,
			"full_name": targetUser.FullName,
			"email":     targetUser.Email,
			"role":      newRole,
		},
	})
}

type CreateClassInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func TeacherCreateClass(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	var input CreateClassInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Class name cannot be empty"})
		return
	}

	class := models.Classroom{
		Name:        name,
		Code:        generateClassCode(),
		Description: strings.TrimSpace(input.Description),
		TeacherID:   teacherID,
	}

	if err := config.DB.Create(&class).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create classroom"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Class created successfully",
		"class":   class,
	})
}

type UpdateClassInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func TeacherUpdateClass(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	classID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var class models.Classroom
	if err := config.DB.Where("id = ? AND teacher_id = ?", classID, teacherID).First(&class).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found or you do not have permission"})
		return
	}

	var input UpdateClassInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Class name cannot be empty"})
			return
		}
		updates["name"] = name
	}
	if input.Description != nil {
		updates["description"] = strings.TrimSpace(*input.Description)
	}

	if len(updates) > 0 {
		if err := config.DB.Model(&class).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update classroom"})
			return
		}
	}

	config.DB.First(&class, classID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Class updated successfully",
		"class":   class,
	})
}

func TeacherDeleteClass(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	classID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var class models.Classroom
	if err := config.DB.Where("id = ? AND teacher_id = ?", classID, teacherID).First(&class).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found or you do not have permission"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("classroom_id = ?", classID).Delete(&models.ClassroomMember{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Exam{}).Where("classroom_id = ?", classID).Update("classroom_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&class).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete classroom"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Classroom deleted successfully"})
}

type AddClassMemberInput struct {
	StudentID   *uint   `json:"student_id"`
	Email       *string `json:"email"`
	StudentCode *string `json:"student_code"`
}

func TeacherAddClassMember(c *gin.Context) {
	teacherIDVal, _ := c.Get("user_id")
	teacherID := teacherIDVal.(uint)

	classID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var class models.Classroom
	if err := config.DB.Where("id = ? AND teacher_id = ?", classID, teacherID).First(&class).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found or you do not have permission"})
		return
	}

	var input AddClassMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var student models.User
	var findErr error

	if input.StudentID != nil && *input.StudentID > 0 {
		findErr = config.DB.Where("id = ? AND role = ?", *input.StudentID, models.RoleStudent).First(&student).Error
	} else if input.Email != nil && strings.TrimSpace(*input.Email) != "" {
		email := strings.ToLower(strings.TrimSpace(*input.Email))
		findErr = config.DB.Where("email = ? AND role = ?", email, models.RoleStudent).First(&student).Error
	} else if input.StudentCode != nil && strings.TrimSpace(*input.StudentCode) != "" {
		code := strings.TrimSpace(*input.StudentCode)
		findErr = config.DB.Where("student_code = ? AND role = ?", code, models.RoleStudent).First(&student).Error
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide student_id, email, or student_code"})
		return
	}

	if findErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	var existing models.ClassroomMember
	if err := config.DB.Where("classroom_id = ? AND student_id = ?", class.ID, student.ID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Student is already in this class"})
		return
	}

	member := models.ClassroomMember{
		ClassroomID: class.ID,
		StudentID:   student.ID,
		JoinedAt:    time.Now(),
	}

	if err := config.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member to classroom"})
		return
	}

	CreateNotification(
		student.ID,
		"Đã được thêm vào lớp học",
		fmt.Sprintf("Bạn đã được thêm vào lớp học \"%s\".", class.Name),
		fmt.Sprintf("/classes/%d", class.ID),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Student added to class successfully",
		"member": gin.H{
			"id":           member.ID,
			"classroom_id": member.ClassroomID,
			"student_id":   member.StudentID,
			"joined_at":    member.JoinedAt,
			"student": gin.H{
				"id":           student.ID,
				"full_name":    student.FullName,
				"email":        student.Email,
				"student_code": student.StudentCode,
				"department":   student.Department,
				"avatar_url":   student.AvatarURL,
			},
		},
	})
}

func TeacherRemoveClassMember(c *gin.Context) {
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

	var class models.Classroom
	if err := config.DB.Where("id = ? AND teacher_id = ?", classID, teacherID).First(&class).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found or you do not have permission"})
		return
	}

	res := config.DB.Where("classroom_id = ? AND student_id = ?", classID, studentID).Delete(&models.ClassroomMember{})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student is not in this class"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed from class successfully"})
}

type JoinClassInput struct {
	Code string `json:"code" binding:"required"`
}

func StudentJoinClass(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	var input JoinClassInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code := strings.ToUpper(strings.TrimSpace(input.Code))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Class code is required"})
		return
	}

	var class models.Classroom
	if err := config.DB.Where("code = ?", code).First(&class).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found with this code"})
		return
	}

	var existing models.ClassroomMember
	if err := config.DB.Where("classroom_id = ? AND student_id = ?", class.ID, userID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You have already joined this classroom"})
		return
	}

	member := models.ClassroomMember{
		ClassroomID: class.ID,
		StudentID:   userID,
		JoinedAt:    time.Now(),
	}

	if err := config.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join classroom"})
		return
	}

	var student models.User
	config.DB.First(&student, userID)

	CreateNotification(
		class.TeacherID,
		"Thành viên mới tham gia lớp",
		fmt.Sprintf("Sinh viên %s đã tham gia lớp \"%s\".", student.FullName, class.Name),
		fmt.Sprintf("/classes/%d", class.ID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Joined classroom successfully",
		"class":   class,
	})
}

type ClassListItem struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	TeacherID   uint      `json:"teacher_id"`
	TeacherName string    `json:"teacher_name,omitempty"`
	MemberCount int64     `json:"member_count"`
	ExamCount   int64     `json:"exam_count"`
	CreatedAt   time.Time `json:"created_at"`
}

func ListClasses(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)
	roleVal, _ := c.Get("role")
	role := roleVal.(string)

	var classes []models.Classroom
	if role == models.RoleTeacher {
		if err := config.DB.Where("teacher_id = ?", userID).Order("created_at DESC").Find(&classes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch classes"})
			return
		}
	} else {
		var classIDs []uint
		config.DB.Model(&models.ClassroomMember{}).Where("student_id = ?", userID).Pluck("classroom_id", &classIDs)
		if len(classIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{"classes": []ClassListItem{}})
			return
		}
		if err := config.DB.Where("id IN ?", classIDs).Preload("Teacher").Order("created_at DESC").Find(&classes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch classes"})
			return
		}
	}

	items := make([]ClassListItem, 0, len(classes))
	for _, cl := range classes {
		var memberCount int64
		config.DB.Model(&models.ClassroomMember{}).Where("classroom_id = ?", cl.ID).Count(&memberCount)

		var examCount int64
		config.DB.Model(&models.Exam{}).Where("classroom_id = ?", cl.ID).Count(&examCount)

		item := ClassListItem{
			ID:          cl.ID,
			Name:        cl.Name,
			Code:        cl.Code,
			Description: cl.Description,
			TeacherID:   cl.TeacherID,
			TeacherName: cl.Teacher.FullName,
			MemberCount: memberCount,
			ExamCount:   examCount,
			CreatedAt:   cl.CreatedAt,
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{"classes": items})
}

func GetClassDetail(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)
	roleVal, _ := c.Get("role")
	role := roleVal.(string)

	classID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var class models.Classroom
	if err := config.DB.Preload("Teacher").First(&class, classID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		return
	}

	if role == models.RoleTeacher {
		if class.TeacherID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this classroom"})
			return
		}
	} else {
		var count int64
		config.DB.Model(&models.ClassroomMember{}).Where("classroom_id = ? AND student_id = ?", classID, userID).Count(&count)
		if count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not enrolled in this classroom"})
			return
		}
	}

	var rawMembers []models.ClassroomMember
	config.DB.Where("classroom_id = ?", classID).
		Preload("Student").
		Order("joined_at DESC").
		Find(&rawMembers)

	type MemberInfo struct {
		ID          uint      `json:"id"`
		StudentID   uint      `json:"student_id"`
		FullName    string    `json:"full_name"`
		Email       string    `json:"email"`
		AvatarURL   string    `json:"avatar_url"`
		Bio         string    `json:"bio"`
		Department  string    `json:"department"`
		StudentCode string    `json:"student_code"`
		JoinedAt    time.Time `json:"joined_at"`
	}

	members := make([]MemberInfo, 0, len(rawMembers))
	for _, m := range rawMembers {
		members = append(members, MemberInfo{
			ID:          m.ID,
			StudentID:   m.StudentID,
			FullName:    m.Student.FullName,
			Email:       m.Student.Email,
			AvatarURL:   m.Student.AvatarURL,
			Bio:         m.Student.Bio,
			Department:  m.Student.Department,
			StudentCode: m.Student.StudentCode,
			JoinedAt:    m.JoinedAt,
		})
	}

	var exams []models.Exam
	config.DB.Where("classroom_id = ?", classID).Order("start_time DESC, id DESC").Find(&exams)

	c.JSON(http.StatusOK, gin.H{
		"class": gin.H{
			"id":          class.ID,
			"name":        class.Name,
			"code":        class.Code,
			"description": class.Description,
			"teacher_id":  class.TeacherID,
			"created_at":  class.CreatedAt,
		},
		"teacher": gin.H{
			"id":         class.Teacher.ID,
			"full_name":  class.Teacher.FullName,
			"email":      class.Teacher.Email,
			"avatar_url": class.Teacher.AvatarURL,
			"bio":        class.Teacher.Bio,
			"department": class.Teacher.Department,
		},
		"members": members,
		"exams":   exams,
	})
}
