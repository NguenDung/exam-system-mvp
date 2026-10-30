package handlers

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"exam-system-mvp/config"
	"exam-system-mvp/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RecordViolationInput struct {
	Reason string `json:"reason"`
}

func StudentRecordViolation(c *gin.Context) {
	userID, _ := c.Get("user_id")
	attemptID, ok := parseUintParam(c, "attempt_id")
	if !ok {
		return
	}

	var autoSubmitted bool
	var updatedAttempt models.ExamAttempt

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var attempt models.ExamAttempt
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", attemptID, userID).
			First(&attempt).Error; err != nil {
			return err
		}
		if attempt.Status != "in_progress" {
			return errAlreadySubmitted
		}

		attempt.ViolationCount++
		if err := tx.Model(&attempt).Update("violation_count", attempt.ViolationCount).Error; err != nil {
			return err
		}

		if attempt.ViolationCount >= 3 {
			if err := gradeAttemptInTx(tx, attempt.ID); err != nil {
				return err
			}
			autoSubmitted = true
		}

		return tx.First(&updatedAttempt, attempt.ID).Error
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}
	if errors.Is(err, errAlreadySubmitted) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exam already submitted"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record violation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Violation recorded",
		"violation_count": updatedAttempt.ViolationCount,
		"auto_submitted":  autoSubmitted,
		"attempt":         updatedAttempt,
	})
}

func TeacherGetAuditLogs(c *gin.Context) {
	var logs []models.AuditLog
	if err := config.DB.Preload("Actor").Order("created_at DESC").Limit(100).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load audit logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func TeacherExportResultsCSV(c *gin.Context) {
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

	gradeExpiredAttempts(exam.ID)

	var results []StudentResult
	config.DB.Table("exam_attempts ea").
		Select("ea.id as attempt_id, ea.user_id, u.full_name, u.email, ea.score, ea.teacher_note, ea.violation_count, ea.correct_count, ea.wrong_count, ea.status, ea.submitted_at, ea.started_at").
		Joins("JOIN users u ON u.id = ea.user_id").
		Where("ea.exam_id = ?", exam.ID).
		Order("ea.score DESC").
		Scan(&results)

	buf := &bytes.Buffer{}
	// UTF-8 BOM for Excel
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(buf)
	_ = writer.Write([]string{"Họ tên", "Email", "Điểm số", "Trạng thái", "Số lần vi phạm", "Thời gian nộp"})

	for _, r := range results {
		statusStr := "Đang làm"
		if r.Status == "completed" {
			statusStr = "Đã nộp"
		}
		submittedAtStr := ""
		if r.SubmittedAt != nil {
			submittedAtStr = r.SubmittedAt.Format("2006-01-02 15:04:05")
		}
		_ = writer.Write([]string{
			r.FullName,
			r.Email,
			fmt.Sprintf("%.2f", r.Score),
			statusStr,
			strconv.Itoa(r.ViolationCount),
			submittedAtStr,
		})
	}
	writer.Flush()

	filename := fmt.Sprintf("bang-diem-exam-%d.csv", exam.ID)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
}
