package models

import (
	"errors"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedData creates demo users and sample exam data when they are missing.
func SeedData(db *gorm.DB) {
	demoUsers := []struct {
		email    string
		fullName string
		role     string
	}{
		{email: "teacher@test.com", fullName: "Nguyễn Văn Giáo", role: RoleTeacher},
		{email: "student@test.com", fullName: "Trần Thị Học", role: RoleStudent},
		{email: "student2@test.com", fullName: "Lê Văn Minh", role: RoleStudent},
		{email: "student3@test.com", fullName: "Phạm Thị Lan", role: RoleStudent},
	}

	var teacher User
	for _, demoUser := range demoUsers {
		user, err := ensureDemoUser(db, demoUser.email, demoUser.fullName, demoUser.role)
		if err != nil {
			log.Printf("Seed: failed to ensure demo user %s: %v", demoUser.email, err)
			return
		}
		if demoUser.role == RoleTeacher {
			teacher = user
		}
	}

	var exam Exam
	err := db.Where("exam_code = ?", "IT2026").First(&exam).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Seed: failed to find sample exam: %v", err)
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		now := time.Now()
		exam = Exam{
			TeacherID:   teacher.ID,
			ExamCode:    "IT2026",
			Title:       "Kiểm tra Lập trình Web & Mạng máy tính",
			Duration:    15,
			StartTime:   now.Add(-1 * time.Hour),
			EndTime:     now.Add(7 * 24 * time.Hour),
			PublishedAt: &now,
			Status:      "published",
		}
		if err := db.Create(&exam).Error; err != nil {
			log.Printf("Seed: failed to create sample exam: %v", err)
			return
		}
	}

	var questionCount int64
	if err := db.Model(&Question{}).Where("exam_id = ?", exam.ID).Count(&questionCount).Error; err != nil {
		log.Printf("Seed: failed to count sample questions: %v", err)
		return
	}
	if questionCount > 0 {
		log.Printf("Seed data ready: demo users ensured, exam=%s (id=%d)", exam.ExamCode, exam.ID)
		return
	}

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
	if err := db.Create(&questions).Error; err != nil {
		log.Printf("Seed: failed to create sample questions: %v", err)
		return
	}

	log.Printf("Seed data ready: demo users ensured, exam=%s (id=%d)", exam.ExamCode, exam.ID)
}

func ensureDemoUser(db *gorm.DB, email, fullName, role string) (User, error) {
	var user User
	err := db.Where("email = ?", email).First(&user).Error
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	user = User{
		Email:        email,
		PasswordHash: string(passwordHash),
		FullName:     fullName,
		Role:         role,
	}
	if err := db.Create(&user).Error; err != nil {
		return User{}, err
	}
	return user, nil
}
