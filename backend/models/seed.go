package models

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedData seeds initial data if the users table is empty.
func SeedData(db *gorm.DB) {
	var count int64
	db.Model(&User{}).Count(&count)
	if count > 0 {
		log.Println("Seed data already exists, skipping.")
		return
	}

	log.Println("Seeding initial data...")

	// ── Hash passwords ────────────────────────────────────────────────────────
	teacherHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	studentHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	// ── Create teacher ────────────────────────────────────────────────────────
	teacher := User{
		Email:        "teacher@test.com",
		PasswordHash: string(teacherHash),
		FullName:     "Nguyễn Văn Giáo",
		Role:         "teacher",
	}
	if err := db.Create(&teacher).Error; err != nil {
		log.Printf("Seed: failed to create teacher: %v", err)
		return
	}

	// ── Create student ────────────────────────────────────────────────────────
	student := User{
		Email:        "student@test.com",
		PasswordHash: string(studentHash),
		FullName:     "Trần Thị Học",
		Role:         "student",
	}
	if err := db.Create(&student).Error; err != nil {
		log.Printf("Seed: failed to create student: %v", err)
		return
	}

	// ── Create sample published exam ──────────────────────────────────────────
	now := time.Now()
	exam := Exam{
		TeacherID: teacher.ID,
		ExamCode:  "IT2026",
		Title:     "Kiểm tra Lập trình Web & Mạng máy tính",
		Duration:  15,
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now.Add(7 * 24 * time.Hour),
		Status:    "published",
	}
	if err := db.Create(&exam).Error; err != nil {
		log.Printf("Seed: failed to create exam: %v", err)
		return
	}

	// ── Create 5 sample questions ─────────────────────────────────────────────
	questions := []Question{
		{
			ExamID:        exam.ID,
			SortOrder:     1,
			Prompt:        "Giao thức nào được sử dụng để truyền tải trang web an toàn (mã hóa) qua Internet?",
			OptionA:       "HTTP",
			OptionB:       "FTP",
			OptionC:       "HTTPS",
			OptionD:       "SMTP",
			CorrectOption: "C",
			Points:        1.0,
		},
		{
			ExamID:        exam.ID,
			SortOrder:     2,
			Prompt:        "Trong HTML5, thẻ nào dùng để nhúng video vào trang web?",
			OptionA:       "<media>",
			OptionB:       "<video>",
			OptionC:       "<embed>",
			OptionD:       "<stream>",
			CorrectOption: "B",
			Points:        1.0,
		},
		{
			ExamID:        exam.ID,
			SortOrder:     3,
			Prompt:        "Mô hình OSI có bao nhiêu tầng (layer)?",
			OptionA:       "4",
			OptionB:       "5",
			OptionC:       "6",
			OptionD:       "7",
			CorrectOption: "D",
			Points:        1.0,
		},
		{
			ExamID:        exam.ID,
			SortOrder:     4,
			Prompt:        "Trong CSS, thuộc tính nào dùng để thay đổi màu nền của phần tử?",
			OptionA:       "color",
			OptionB:       "background-color",
			OptionC:       "bg-color",
			OptionD:       "fill",
			CorrectOption: "B",
			Points:        1.0,
		},
		{
			ExamID:        exam.ID,
			SortOrder:     5,
			Prompt:        "DNS (Domain Name System) có chức năng chính là gì?",
			OptionA:       "Mã hóa dữ liệu truyền tải trên mạng",
			OptionB:       "Phân chia băng thông giữa các máy tính",
			OptionC:       "Chuyển đổi tên miền thành địa chỉ IP",
			OptionD:       "Quản lý kết nối VPN",
			CorrectOption: "C",
			Points:        1.0,
		},
	}

	for _, q := range questions {
		if err := db.Create(&q).Error; err != nil {
			log.Printf("Seed: failed to create question: %v", err)
		}
	}

	log.Printf("Seed completed: teacher=%s, student=%s, exam=%s (id=%d)",
		teacher.Email, student.Email, exam.ExamCode, exam.ID)
}
