package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"exam-system-mvp/config"
	"exam-system-mvp/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UpdateExamInput struct {
	Title           *string    `json:"title"`
	Description     *string    `json:"description"`
	Duration        *int       `json:"duration"`
	DurationMinutes *int       `json:"duration_minutes"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	Status          *string    `json:"status"`
	IsPublished     *bool      `json:"is_published"`
}

// TeacherUpdateExam updates an existing exam.
// PUT /api/teacher/exams/:id
func TeacherUpdateExam(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var exam models.Exam
	if err := config.DB.Where("id = ? AND teacher_id = ?", examID, teacherID).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found or you do not have permission"})
		return
	}

	var input UpdateExamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty"})
			return
		}
		updates["title"] = title
	}

	if input.Description != nil {
		updates["description"] = strings.TrimSpace(*input.Description)
	}

	if input.DurationMinutes != nil && *input.DurationMinutes > 0 {
		updates["duration"] = *input.DurationMinutes
	} else if input.Duration != nil && *input.Duration > 0 {
		updates["duration"] = *input.Duration
	}

	newStartTime := exam.StartTime
	if input.StartTime != nil {
		newStartTime = *input.StartTime
		updates["start_time"] = newStartTime
	}

	newEndTime := exam.EndTime
	if input.EndTime != nil {
		newEndTime = *input.EndTime
		updates["end_time"] = newEndTime
	}

	if !newEndTime.After(newStartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_time must be after start_time"})
		return
	}

	// Status / IsPublished handling
	targetStatus := exam.Status
	if input.IsPublished != nil {
		if *input.IsPublished {
			targetStatus = "published"
		} else {
			targetStatus = "draft"
		}
	} else if input.Status != nil {
		st := strings.ToLower(strings.TrimSpace(*input.Status))
		if st == "published" || st == "draft" {
			targetStatus = st
		}
	}

	if targetStatus != exam.Status {
		if targetStatus == "published" {
			var qCount int64
			config.DB.Model(&models.Question{}).Where("exam_id = ?", exam.ID).Count(&qCount)
			if qCount == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot publish an exam with no questions"})
				return
			}
			now := time.Now()
			updates["published_at"] = &now
		}
		updates["status"] = targetStatus
	}

	if len(updates) > 0 {
		if err := config.DB.Model(&exam).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exam"})
			return
		}
	}

	config.DB.First(&exam, exam.ID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Exam updated successfully",
		"exam":    exam,
	})
}

// TeacherDeleteExam deletes an exam and its associated questions, attempts, answers, and appeals.
// DELETE /api/teacher/exams/:id
func TeacherDeleteExam(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var exam models.Exam
	if err := config.DB.Where("id = ? AND teacher_id = ?", examID, teacherID).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found or you do not have permission"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Delete attempt answers
		if err := tx.Exec(`
			DELETE FROM attempt_answers
			WHERE attempt_id IN (SELECT id FROM exam_attempts WHERE exam_id = ?)
		`, exam.ID).Error; err != nil {
			return err
		}

		// 2. Delete appeals
		if err := tx.Where("exam_id = ?", exam.ID).Delete(&models.Appeal{}).Error; err != nil {
			return err
		}

		// 3. Delete exam attempts
		if err := tx.Where("exam_id = ?", exam.ID).Delete(&models.ExamAttempt{}).Error; err != nil {
			return err
		}

		// 4. Delete questions
		if err := tx.Where("exam_id = ?", exam.ID).Delete(&models.Question{}).Error; err != nil {
			return err
		}

		// 5. Delete exam (hard delete or soft delete)
		return tx.Delete(&exam).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete exam"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exam and all related records deleted successfully"})
}

type UpdateQuestionInput struct {
	Prompt        string  `json:"prompt" binding:"required"`
	OptionA       string  `json:"option_a" binding:"required"`
	OptionB       string  `json:"option_b" binding:"required"`
	OptionC       string  `json:"option_c" binding:"required"`
	OptionD       string  `json:"option_d" binding:"required"`
	CorrectOption string  `json:"correct_option" binding:"required,oneof=A B C D"`
	Points        float64 `json:"points"`
}

// TeacherUpdateQuestion updates a question's content, choices, or correct answer.
// PUT /api/teacher/questions/:id
func TeacherUpdateQuestion(c *gin.Context) {
	teacherID, _ := c.Get("user_id")

	// Support both :id and :question_id params
	paramKey := "id"
	if c.Param("id") == "" && c.Param("question_id") != "" {
		paramKey = "question_id"
	}
	questionID, ok := parseUintParam(c, paramKey)
	if !ok {
		return
	}

	var question models.Question
	if err := config.DB.First(&question, questionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	// Verify teacher owns the parent exam
	var exam models.Exam
	if err := config.DB.Where("id = ? AND teacher_id = ?", question.ExamID, teacherID).First(&exam).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to edit this question"})
		return
	}

	var input UpdateQuestionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	points := input.Points
	if points <= 0 {
		points = 1.0
	}

	updates := map[string]interface{}{
		"prompt":         strings.TrimSpace(input.Prompt),
		"option_a":       strings.TrimSpace(input.OptionA),
		"option_b":       strings.TrimSpace(input.OptionB),
		"option_c":       strings.TrimSpace(input.OptionC),
		"option_d":       strings.TrimSpace(input.OptionD),
		"correct_option": input.CorrectOption,
		"points":         points,
		"updated_at":     time.Now(),
	}

	if err := config.DB.Model(&question).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update question"})
		return
	}

	config.DB.First(&question, question.ID)
	c.JSON(http.StatusOK, gin.H{
		"message":  "Question updated successfully",
		"question": question,
	})
}

type BulkQuestionItem struct {
	Prompt        string  `json:"prompt" binding:"required"`
	OptionA       string  `json:"option_a" binding:"required"`
	OptionB       string  `json:"option_b" binding:"required"`
	OptionC       string  `json:"option_c" binding:"required"`
	OptionD       string  `json:"option_d" binding:"required"`
	CorrectOption string  `json:"correct_option" binding:"required,oneof=A B C D"`
	Points        float64 `json:"points"`
}

type BulkQuestionsPayload struct {
	Questions []BulkQuestionItem `json:"questions"`
}

// TeacherBulkAddQuestions adds multiple questions at once from a JSON array or object.
// POST /api/teacher/exams/:id/questions/bulk
func TeacherBulkAddQuestions(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var exam models.Exam
	if err := config.DB.Where("id = ? AND teacher_id = ?", examID, teacherID).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	if exam.Status == "published" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot add questions to a published exam"})
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var items []BulkQuestionItem
	// Try parsing as { "questions": [...] } first
	var wrapper BulkQuestionsPayload
	if err := json.Unmarshal(bodyBytes, &wrapper); err == nil && len(wrapper.Questions) > 0 {
		items = wrapper.Questions
	} else {
		// Try parsing as direct array [...]
		if err := json.Unmarshal(bodyBytes, &items); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format: expected array of questions or { questions: [...] }"})
			return
		}
	}

	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No questions provided"})
		return
	}

	// Validate items
	for i, q := range items {
		prompt := strings.TrimSpace(q.Prompt)
		optA := strings.TrimSpace(q.OptionA)
		optB := strings.TrimSpace(q.OptionB)
		optC := strings.TrimSpace(q.OptionC)
		optD := strings.TrimSpace(q.OptionD)
		corr := strings.ToUpper(strings.TrimSpace(q.CorrectOption))

		if prompt == "" || optA == "" || optB == "" || optC == "" || optD == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Question %d: prompt and all 4 options are required", i+1)})
			return
		}
		if corr != "A" && corr != "B" && corr != "C" && corr != "D" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Question %d: correct_option must be A, B, C, or D", i+1)})
			return
		}
	}

	// Determine start order
	var maxOrder int
	config.DB.Model(&models.Question{}).
		Where("exam_id = ?", exam.ID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxOrder)

	createdQuestions := make([]models.Question, len(items))
	now := time.Now()

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		for i, q := range items {
			pts := q.Points
			if pts <= 0 {
				pts = 1.0
			}
			newQ := models.Question{
				ExamID:        exam.ID,
				SortOrder:     maxOrder + 1 + i,
				Prompt:        strings.TrimSpace(q.Prompt),
				OptionA:       strings.TrimSpace(q.OptionA),
				OptionB:       strings.TrimSpace(q.OptionB),
				OptionC:       strings.TrimSpace(q.OptionC),
				OptionD:       strings.TrimSpace(q.OptionD),
				CorrectOption: strings.ToUpper(strings.TrimSpace(q.CorrectOption)),
				Points:        pts,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if err := tx.Create(&newQ).Error; err != nil {
				return err
			}
			createdQuestions[i] = newQ
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk create questions"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   fmt.Sprintf("Successfully imported %d questions", len(createdQuestions)),
		"questions": createdQuestions,
	})
}

type UpdateAttemptScoreInput struct {
	Score float64 `json:"score"`
	Note  string  `json:"note"`
}

// TeacherUpdateAttemptScore manually updates a student attempt's score with a note.
// PATCH /api/teacher/attempts/:id/score
func TeacherUpdateAttemptScore(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	attemptID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var attempt models.ExamAttempt
	if err := config.DB.First(&attempt, attemptID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	var exam models.Exam
	if err := config.DB.Where("id = ? AND teacher_id = ?", attempt.ExamID, teacherID).First(&exam).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to modify this attempt"})
		return
	}

	var input UpdateAttemptScoreInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Score < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Score cannot be negative"})
		return
	}

	updates := map[string]interface{}{
		"score":        input.Score,
		"teacher_note": strings.TrimSpace(input.Note),
	}

	if err := config.DB.Model(&attempt).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update attempt score"})
		return
	}

	// Record audit log
	audit := models.AuditLog{
		ActorID:   teacherID.(uint),
		Action:    "UPDATE_SCORE",
		TargetID:  attempt.ID,
		Details:   fmt.Sprintf("Hiệu chỉnh điểm từ %.2f sang %.2f. Ghi chú: %s", attempt.Score, input.Score, strings.TrimSpace(input.Note)),
		CreatedAt: time.Now(),
	}
	_ = config.DB.Create(&audit)

	CreateNotification(attempt.UserID, "Cập nhật điểm thi", fmt.Sprintf("Giảng viên đã hiệu chỉnh điểm thi của bạn trên đề thi \"%s\" thành %.2f điểm.", exam.Title, input.Score))

	config.DB.First(&attempt, attempt.ID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Attempt score updated successfully",
		"attempt": attempt,
	})
}

// TeacherResetAttempt deletes a student's attempt to allow retaking the exam from scratch.
// DELETE /api/teacher/attempts/:id
func TeacherResetAttempt(c *gin.Context) {
	teacherID, _ := c.Get("user_id")
	attemptID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var attempt models.ExamAttempt
	if err := config.DB.First(&attempt, attemptID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	var exam models.Exam
	if err := config.DB.Where("id = ? AND teacher_id = ?", attempt.ExamID, teacherID).First(&exam).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to reset this attempt"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Delete answers
		if err := tx.Where("attempt_id = ?", attempt.ID).Delete(&models.AttemptAnswer{}).Error; err != nil {
			return err
		}
		// Delete appeals on this attempt
		if err := tx.Where("submission_id = ?", attempt.ID).Delete(&models.Appeal{}).Error; err != nil {
			return err
		}
		// Record audit log inside transaction
		audit := models.AuditLog{
			ActorID:   teacherID.(uint),
			Action:    "RESET_ATTEMPT",
			TargetID:  attempt.ID,
			Details:   fmt.Sprintf("Xóa lượt thi của thí sinh (User ID: %d, Điểm cũ: %.2f) trên đề thi %d", attempt.UserID, attempt.Score, attempt.ExamID),
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&audit).Error; err != nil {
			return err
		}
		// Delete the attempt itself
		return tx.Delete(&attempt).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset attempt"})
		return
	}

	CreateNotification(attempt.UserID, "Được phép thi lại", fmt.Sprintf("Lượt thi của bạn trên đề thi \"%s\" đã được giảng viên thiết lập lại. Bạn có thể bắt đầu thi lại từ đầu.", exam.Title))

	c.JSON(http.StatusOK, gin.H{"message": "Attempt reset successfully. Student can now retake this exam."})
}

type StudentListItem struct {
	ID                uint      `json:"id"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	Role              string    `json:"role"`
	CreatedAt         time.Time `json:"created_at"`
	TotalAttempts     int64     `json:"total_attempts"`
	CompletedAttempts int64     `json:"completed_attempts"`
}

// TeacherListStudents returns all student accounts with participation statistics.
// GET /api/teacher/students
func TeacherListStudents(c *gin.Context) {
	search := strings.ToLower(strings.TrimSpace(c.Query("q")))
	if search == "" {
		search = strings.ToLower(strings.TrimSpace(c.Query("search")))
	}

	query := config.DB.Table("users u").
		Select(`
			u.id, u.full_name, u.email, u.role, u.created_at,
			COUNT(ea.id) AS total_attempts,
			COUNT(CASE WHEN ea.status = 'completed' THEN 1 END) AS completed_attempts
		`).
		Joins("LEFT JOIN exam_attempts ea ON ea.user_id = u.id").
		Where("u.role = 'student' AND u.deleted_at IS NULL")

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("(LOWER(u.full_name) LIKE ? OR LOWER(u.email) LIKE ?)", like, like)
	}

	var students []StudentListItem
	if err := query.Group("u.id").Order("u.created_at DESC, u.id DESC").Scan(&students).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load student list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"students": students})
}

type CreateStudentInput struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

// TeacherCreateStudent creates a new student account.
// POST /api/teacher/students
func TeacherCreateStudent(c *gin.Context) {
	var input CreateStudentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fullName := strings.TrimSpace(input.FullName)
	if fullName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "full_name cannot be empty"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address format"})
		return
	}

	var existing models.User
	if err := config.DB.Where("email = ?", email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	pwLen := utf8.RuneCountInString(input.Password)
	if pwLen < 6 || pwLen > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be between 6 and 72 characters"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	student := models.User{
		FullName:     fullName,
		Email:        email,
		PasswordHash: string(hash),
		Role:         models.RoleStudent,
	}

	if err := config.DB.Create(&student).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create student account"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Student created successfully",
		"student": gin.H{
			"id":         student.ID,
			"full_name":  student.FullName,
			"email":      student.Email,
			"role":       student.Role,
			"created_at": student.CreatedAt,
		},
	})
}

type UpdateStudentInput struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TeacherUpdateStudent updates student's name, email, or password.
// PUT /api/teacher/students/:id
func TeacherUpdateStudent(c *gin.Context) {
	studentID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var student models.User
	if err := config.DB.Where("id = ? AND role = 'student'", studentID).First(&student).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student account not found"})
		return
	}

	var input UpdateStudentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}

	if fullName := strings.TrimSpace(input.FullName); fullName != "" {
		updates["full_name"] = fullName
	}

	if rawEmail := strings.TrimSpace(input.Email); rawEmail != "" {
		email := strings.ToLower(rawEmail)
		if _, err := mail.ParseAddress(email); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address format"})
			return
		}
		if email != student.Email {
			var dup models.User
			if err := config.DB.Where("email = ? AND id != ?", email, student.ID).First(&dup).Error; err == nil {
				c.JSON(http.StatusConflict, gin.H{"error": "Email is already used by another account"})
				return
			}
			updates["email"] = email
		}
	}

	if input.Password != "" {
		pwLen := utf8.RuneCountInString(input.Password)
		if pwLen < 6 || pwLen > 72 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be between 6 and 72 characters"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
			return
		}
		updates["password_hash"] = string(hash)
	}

	if len(updates) > 0 {
		if err := config.DB.Model(&student).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update student account"})
			return
		}
	}

	config.DB.First(&student, student.ID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Student updated successfully",
		"student": gin.H{
			"id":         student.ID,
			"full_name":  student.FullName,
			"email":      student.Email,
			"role":       student.Role,
			"created_at": student.CreatedAt,
			"updated_at": student.UpdatedAt,
		},
	})
}

// TeacherDeleteStudent deletes a student account along with all their exam attempts and appeals.
// DELETE /api/teacher/students/:id
func TeacherDeleteStudent(c *gin.Context) {
	studentID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var student models.User
	if err := config.DB.Where("id = ? AND role = 'student'", studentID).First(&student).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student account not found"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Delete answers
		if err := tx.Exec(`
			DELETE FROM attempt_answers
			WHERE attempt_id IN (SELECT id FROM exam_attempts WHERE user_id = ?)
		`, student.ID).Error; err != nil {
			return err
		}

		// 2. Delete appeals
		if err := tx.Where("student_id = ?", student.ID).Delete(&models.Appeal{}).Error; err != nil {
			return err
		}

		// 3. Delete exam attempts
		if err := tx.Where("user_id = ?", student.ID).Delete(&models.ExamAttempt{}).Error; err != nil {
			return err
		}

		// 4. Delete user
		return tx.Delete(&student).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete student account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student account and related history deleted successfully"})
}

type AvailableExamItem struct {
	ID            uint                `json:"id"`
	ExamCode      string              `json:"exam_code"`
	Title         string              `json:"title"`
	Description   string              `json:"description"`
	Duration      int                 `json:"duration"`
	StartTime     time.Time           `json:"start_time"`
	EndTime       time.Time           `json:"end_time"`
	QuestionCount int64               `json:"question_count"`
	TotalPoints   float64             `json:"total_points"`
	ExamStatus    string              `json:"exam_status"`
	AttemptStatus string              `json:"attempt_status"`
	Attempt       *models.ExamAttempt `json:"attempt,omitempty"`
}

// StudentGetAvailableExams returns all published exams with current student's status.
// GET /api/student/exams/available
func StudentGetAvailableExams(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// Proactively grade expired attempts for this student
	var activeExamIDs []uint
	config.DB.Model(&models.ExamAttempt{}).
		Where("user_id = ? AND status = ?", userID, "in_progress").
		Distinct("exam_id").
		Pluck("exam_id", &activeExamIDs)
	for _, eid := range activeExamIDs {
		gradeExpiredAttempts(eid)
	}

	var exams []models.Exam
	if err := config.DB.Where("status = ?", "published").
		Order("start_time DESC, id DESC").
		Find(&exams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load exams"})
		return
	}

	now := time.Now()
	items := make([]AvailableExamItem, 0, len(exams))

	for _, exam := range exams {
		// Calculate time status
		var examTimeStatus string
		switch {
		case now.Before(exam.StartTime):
			examTimeStatus = "not_started_yet"
		case now.After(exam.EndTime):
			examTimeStatus = "expired"
		default:
			examTimeStatus = "open"
		}

		// Check student attempt
		var attempt models.ExamAttempt
		hasAttempt := config.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&attempt).Error == nil

		var attemptStatus string
		if hasAttempt {
			if attempt.Status == "completed" {
				attemptStatus = "submitted"
			} else {
				deadline := attemptDeadline(attempt.StartedAt, exam)
				if !now.Before(deadline) {
					attemptStatus = "time_up"
				} else {
					attemptStatus = "in_progress"
				}
			}
		} else {
			attemptStatus = "not_attempted"
		}

		// Question count and points
		var qStats struct {
			Count  int64
			Points float64
		}
		config.DB.Model(&models.Question{}).
			Where("exam_id = ?", exam.ID).
			Select("COUNT(id) AS count, COALESCE(SUM(points), 0) AS points").
			Scan(&qStats)

		item := AvailableExamItem{
			ID:            exam.ID,
			ExamCode:      exam.ExamCode,
			Title:         exam.Title,
			Description:   exam.Description,
			Duration:      exam.Duration,
			StartTime:     exam.StartTime,
			EndTime:       exam.EndTime,
			QuestionCount: qStats.Count,
			TotalPoints:   qStats.Points,
			ExamStatus:    examTimeStatus,
			AttemptStatus: attemptStatus,
		}
		if hasAttempt {
			item.Attempt = &attempt
		}

		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{"exams": items})
}
