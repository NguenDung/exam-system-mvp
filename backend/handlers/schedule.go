package handlers

import (
	"net/http"
	"time"

	"exam-system-mvp/config"
	"exam-system-mvp/models"

	"github.com/gin-gonic/gin"
)

type ScheduleExamItem struct {
	ID               uint      `json:"id"`
	ExamCode         string    `json:"exam_code"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Duration         int       `json:"duration"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	Status           string    `json:"status"`
	ClassroomID      *uint     `json:"classroom_id,omitempty"`
	ClassroomName    string    `json:"classroom_name"`
	ClassroomCode    string    `json:"classroom_code"`
	TeacherID        uint      `json:"teacher_id"`
	TeacherName      string    `json:"teacher_name"`
	QuestionCount    int64     `json:"question_count"`
	TotalPoints      float64   `json:"total_points"`
	CountdownSeconds int64     `json:"countdown_seconds"`
	TimeStatus       string    `json:"time_status"`

	AttemptStatus string   `json:"attempt_status,omitempty"`
	AttemptID     *uint    `json:"attempt_id,omitempty"`
	Score         *float64 `json:"score,omitempty"`

	TotalAttempts int64 `json:"total_attempts,omitempty"`
}

func GetExamSchedule(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	roleVal, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	role := roleVal.(string)

	now := time.Now()

	var exams []models.Exam
	if role == models.RoleTeacher {
		if err := config.DB.Where("teacher_id = ?", userID).
			Order("start_time ASC, id ASC").
			Find(&exams).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load schedule"})
			return
		}
	} else {
		var activeExamIDs []uint
		config.DB.Model(&models.ExamAttempt{}).
			Where("user_id = ? AND status = ?", userID, "in_progress").
			Distinct("exam_id").
			Pluck("exam_id", &activeExamIDs)
		for _, eid := range activeExamIDs {
			gradeExpiredAttempts(eid)
		}

		var enrolledClassIDs []uint
		config.DB.Model(&models.ClassroomMember{}).Where("student_id = ?", userID).Pluck("classroom_id", &enrolledClassIDs)

		query := config.DB.Where("status = ?", "published")
		if len(enrolledClassIDs) > 0 {
			query = query.Where("classroom_id IS NULL OR classroom_id IN ?", enrolledClassIDs)
		} else {
			query = query.Where("classroom_id IS NULL")
		}
		if err := query.Order("start_time ASC, id ASC").Find(&exams).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load schedule"})
			return
		}
	}

	ongoing := make([]ScheduleExamItem, 0)
	upcoming := make([]ScheduleExamItem, 0)
	past := make([]ScheduleExamItem, 0)

	classroomMap := make(map[uint]models.Classroom)
	var classrooms []models.Classroom
	if err := config.DB.Find(&classrooms).Error; err == nil {
		for _, cls := range classrooms {
			classroomMap[cls.ID] = cls
		}
	}

	teacherMap := make(map[uint]string)
	var teachers []models.User
	if err := config.DB.Where("role = ?", models.RoleTeacher).Find(&teachers).Error; err == nil {
		for _, t := range teachers {
			teacherMap[t.ID] = t.FullName
		}
	}

	for _, exam := range exams {
		var timeStatus string
		var countdownSeconds int64

		if now.Before(exam.StartTime) {
			timeStatus = "upcoming"
			countdownSeconds = int64(exam.StartTime.Sub(now).Seconds())
			if countdownSeconds < 0 {
				countdownSeconds = 0
			}
		} else if now.After(exam.EndTime) {
			timeStatus = "past"
			countdownSeconds = 0
		} else {
			timeStatus = "ongoing"
			countdownSeconds = int64(exam.EndTime.Sub(now).Seconds())
			if countdownSeconds < 0 {
				countdownSeconds = 0
			}
		}

		var qStats struct {
			Count  int64
			Points float64
		}
		config.DB.Model(&models.Question{}).
			Where("exam_id = ?", exam.ID).
			Select("COUNT(id) AS count, COALESCE(SUM(points), 0) AS points").
			Scan(&qStats)

		className := "Toàn trường"
		classCode := ""
		if exam.ClassroomID != nil {
			if cls, ok := classroomMap[*exam.ClassroomID]; ok {
				className = cls.Name
				classCode = cls.Code
			}
		}

		teacherName := teacherMap[exam.TeacherID]
		if teacherName == "" {
			teacherName = "Giảng viên"
		}

		item := ScheduleExamItem{
			ID:               exam.ID,
			ExamCode:         exam.ExamCode,
			Title:            exam.Title,
			Description:      exam.Description,
			Duration:         exam.Duration,
			StartTime:        exam.StartTime,
			EndTime:          exam.EndTime,
			Status:           exam.Status,
			ClassroomID:      exam.ClassroomID,
			ClassroomName:    className,
			ClassroomCode:    classCode,
			TeacherID:        exam.TeacherID,
			TeacherName:      teacherName,
			QuestionCount:    qStats.Count,
			TotalPoints:      qStats.Points,
			CountdownSeconds: countdownSeconds,
			TimeStatus:       timeStatus,
		}

		if role == models.RoleTeacher {
			var attemptsCount int64
			config.DB.Model(&models.ExamAttempt{}).Where("exam_id = ?", exam.ID).Count(&attemptsCount)
			item.TotalAttempts = attemptsCount
		} else {
			var attempt models.ExamAttempt
			hasAttempt := config.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&attempt).Error == nil
			if hasAttempt {
				item.AttemptID = &attempt.ID
				if attempt.Status == "completed" {
					item.AttemptStatus = "submitted"
					sc := attempt.Score
					item.Score = &sc
				} else {
					deadline := attemptDeadline(attempt.StartedAt, exam)
					if !now.Before(deadline) {
						item.AttemptStatus = "time_up"
					} else {
						item.AttemptStatus = "in_progress"
					}
				}
			} else {
				item.AttemptStatus = "not_attempted"
			}
		}

		switch timeStatus {
		case "ongoing":
			ongoing = append(ongoing, item)
		case "upcoming":
			upcoming = append(upcoming, item)
		case "past":
			past = append(past, item)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"role":     role,
		"ongoing":  ongoing,
		"upcoming": upcoming,
		"past":     past,
		"summary": gin.H{
			"ongoing_count":  len(ongoing),
			"upcoming_count": len(upcoming),
			"past_count":     len(past),
			"total":          len(ongoing) + len(upcoming) + len(past),
		},
	})
}
