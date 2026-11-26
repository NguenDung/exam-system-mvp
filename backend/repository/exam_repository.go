package repository

import (
	"time"

	"exam-system-mvp/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExamRepository interface {
	CreateExam(exam *models.Exam) error
	FindExamByID(id uint) (*models.Exam, error)
	FindExamByIDWithQuestions(id uint) (*models.Exam, error)
	FindExamByCode(code string) (*models.Exam, error)
	ListExamsByTeacher(teacherID uint) ([]models.Exam, error)
	UpdateExam(exam *models.Exam, updates map[string]interface{}) error
	DeleteExam(id uint) error
	GetPublishedExams(enrolledClassIDs []uint) ([]models.Exam, error)
	GetAllClassrooms() ([]models.Classroom, error)
	FindClassroomByID(id uint) (*models.Classroom, error)

	AddQuestions(examID uint, questions []models.Question) ([]models.Question, error)
	FindQuestionByID(id uint) (*models.Question, error)
	UpdateQuestion(q *models.Question) error
	DeleteQuestion(id uint) error
	GetQuestionStats(examID uint) (count int64, points float64, err error)
	GetMaxQuestionSortOrder(examID uint) (int, error)
	GetQuestionsByExamID(examID uint) ([]models.Question, error)

	CreateAttempt(attempt *models.ExamAttempt) error
	FindAttemptByID(id uint) (*models.ExamAttempt, error)
	FindAttemptByExamAndUser(examID, userID uint) (*models.ExamAttempt, error)
	FindInProgressAttemptsByExam(examID uint) ([]models.ExamAttempt, error)
	FindAllInProgressAttempts() ([]models.ExamAttempt, error)
	ListAttemptsByExamID(examID uint) ([]models.ExamAttempt, error)
	ListAttemptsByUserID(userID uint) ([]models.ExamAttempt, error)
	CountAttemptsByExamID(examID uint) (int64, error)
	UpdateAttempt(attempt *models.ExamAttempt) error
	DeleteAttempt(id uint) error

	SaveAnswersAtomic(attemptID uint, answers []models.AttemptAnswer) error
	GradeAttemptAtomic(attemptID uint, updatedAnswers []models.AttemptAnswer, attempt *models.ExamAttempt) error
	GetAttemptAnswers(attemptID uint) ([]models.AttemptAnswer, error)

	CreateAuditLog(log *models.AuditLog) error
	ListAuditLogs(limit int) ([]models.AuditLog, error)
}

type gormExamRepository struct {
	db *gorm.DB
}

func NewExamRepository(db *gorm.DB) ExamRepository {
	return &gormExamRepository{db: db}
}

func (r *gormExamRepository) CreateExam(exam *models.Exam) error {
	return r.db.Create(exam).Error
}

func (r *gormExamRepository) FindExamByID(id uint) (*models.Exam, error) {
	var exam models.Exam
	if err := r.db.First(&exam, id).Error; err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *gormExamRepository) FindExamByIDWithQuestions(id uint) (*models.Exam, error) {
	var exam models.Exam
	if err := r.db.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC, id ASC")
	}).First(&exam, id).Error; err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *gormExamRepository) FindExamByCode(code string) (*models.Exam, error) {
	var exam models.Exam
	if err := r.db.Where("exam_code = ?", code).First(&exam).Error; err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *gormExamRepository) ListExamsByTeacher(teacherID uint) ([]models.Exam, error) {
	var exams []models.Exam
	err := r.db.Where("teacher_id = ?", teacherID).Order("created_at DESC").Find(&exams).Error
	return exams, err
}

func (r *gormExamRepository) UpdateExam(exam *models.Exam, updates map[string]interface{}) error {
	return r.db.Model(exam).Updates(updates).Error
}

func (r *gormExamRepository) DeleteExam(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var attemptIDs []uint
		if err := tx.Model(&models.ExamAttempt{}).Where("exam_id = ?", id).Pluck("id", &attemptIDs).Error; err != nil {
			return err
		}
		if err := tx.Where("exam_id = ?", id).Delete(&models.Appeal{}).Error; err != nil {
			return err
		}
		if len(attemptIDs) > 0 {
			if err := tx.Where("attempt_id IN ?", attemptIDs).Delete(&models.AttemptAnswer{}).Error; err != nil {
				return err
			}
			if err := tx.Where("submission_id IN ?", attemptIDs).Delete(&models.Appeal{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("exam_id = ?", id).Delete(&models.ExamAttempt{}).Error; err != nil {
			return err
		}
		if err := tx.Where("exam_id = ?", id).Delete(&models.Question{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Exam{}, id).Error
	})
}

func (r *gormExamRepository) GetPublishedExams(enrolledClassIDs []uint) ([]models.Exam, error) {
	var exams []models.Exam
	query := r.db.Where("status = ?", "published")
	if len(enrolledClassIDs) > 0 {
		query = query.Where("classroom_id IS NULL OR classroom_id IN ?", enrolledClassIDs)
	} else {
		query = query.Where("classroom_id IS NULL")
	}
	err := query.Order("start_time DESC, id DESC").Find(&exams).Error
	return exams, err
}

func (r *gormExamRepository) GetAllClassrooms() ([]models.Classroom, error) {
	var classrooms []models.Classroom
	err := r.db.Find(&classrooms).Error
	return classrooms, err
}

func (r *gormExamRepository) FindClassroomByID(id uint) (*models.Classroom, error) {
	var classroom models.Classroom
	if err := r.db.First(&classroom, id).Error; err != nil {
		return nil, err
	}
	return &classroom, nil
}

func (r *gormExamRepository) AddQuestions(examID uint, questions []models.Question) ([]models.Question, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for i := range questions {
			questions[i].ExamID = examID
			if err := tx.Create(&questions[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return questions, err
}

func (r *gormExamRepository) FindQuestionByID(id uint) (*models.Question, error) {
	var q models.Question
	if err := r.db.First(&q, id).Error; err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *gormExamRepository) UpdateQuestion(q *models.Question) error {
	return r.db.Save(q).Error
}

func (r *gormExamRepository) DeleteQuestion(id uint) error {
	return r.db.Delete(&models.Question{}, id).Error
}

func (r *gormExamRepository) GetQuestionStats(examID uint) (count int64, points float64, err error) {
	var stats struct {
		Count  int64
		Points float64
	}
	err = r.db.Model(&models.Question{}).
		Where("exam_id = ?", examID).
		Select("COUNT(id) AS count, COALESCE(SUM(points), 0) AS points").
		Scan(&stats).Error
	return stats.Count, stats.Points, err
}

func (r *gormExamRepository) GetMaxQuestionSortOrder(examID uint) (int, error) {
	var maxOrder int
	err := r.db.Model(&models.Question{}).
		Where("exam_id = ?", examID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxOrder).Error
	return maxOrder, err
}

func (r *gormExamRepository) GetQuestionsByExamID(examID uint) ([]models.Question, error) {
	var questions []models.Question
	err := r.db.Where("exam_id = ?", examID).Order("sort_order ASC, id ASC").Find(&questions).Error
	return questions, err
}

func (r *gormExamRepository) CreateAttempt(attempt *models.ExamAttempt) error {
	return r.db.Create(attempt).Error
}

func (r *gormExamRepository) FindAttemptByID(id uint) (*models.ExamAttempt, error) {
	var attempt models.ExamAttempt
	if err := r.db.First(&attempt, id).Error; err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *gormExamRepository) FindAttemptByExamAndUser(examID, userID uint) (*models.ExamAttempt, error) {
	var attempt models.ExamAttempt
	if err := r.db.Where("exam_id = ? AND user_id = ?", examID, userID).First(&attempt).Error; err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *gormExamRepository) FindInProgressAttemptsByExam(examID uint) ([]models.ExamAttempt, error) {
	var attempts []models.ExamAttempt
	err := r.db.Where("exam_id = ? AND status = ?", examID, "in_progress").Find(&attempts).Error
	return attempts, err
}

func (r *gormExamRepository) FindAllInProgressAttempts() ([]models.ExamAttempt, error) {
	var attempts []models.ExamAttempt
	err := r.db.Where("status = ?", "in_progress").Find(&attempts).Error
	return attempts, err
}

func (r *gormExamRepository) ListAttemptsByExamID(examID uint) ([]models.ExamAttempt, error) {
	var attempts []models.ExamAttempt
	err := r.db.Where("exam_id = ?", examID).Preload("User").Order("started_at DESC").Find(&attempts).Error
	return attempts, err
}

func (r *gormExamRepository) ListAttemptsByUserID(userID uint) ([]models.ExamAttempt, error) {
	var attempts []models.ExamAttempt
	err := r.db.Where("user_id = ?", userID).Preload("Exam").Order("started_at DESC").Find(&attempts).Error
	return attempts, err
}

func (r *gormExamRepository) CountAttemptsByExamID(examID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ExamAttempt{}).Where("exam_id = ?", examID).Count(&count).Error
	return count, err
}

func (r *gormExamRepository) UpdateAttempt(attempt *models.ExamAttempt) error {
	return r.db.Save(attempt).Error
}

func (r *gormExamRepository) DeleteAttempt(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		tx.Where("attempt_id = ?", id).Delete(&models.AttemptAnswer{})
		return tx.Delete(&models.ExamAttempt{}, id).Error
	})
}

func (r *gormExamRepository) SaveAnswersAtomic(attemptID uint, answers []models.AttemptAnswer) error {
	if len(answers) == 0 {
		return nil
	}
	now := time.Now()
	for i := range answers {
		answers[i].AttemptID = attemptID
		answers[i].UpdatedAt = now
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "attempt_id"}, {Name: "question_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"chosen_option", "updated_at"}),
	}).Create(&answers).Error
}

func (r *gormExamRepository) GradeAttemptAtomic(attemptID uint, updatedAnswers []models.AttemptAnswer, attempt *models.ExamAttempt) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		res := tx.Model(&models.ExamAttempt{}).
			Where("id = ? AND status = ?", attemptID, "in_progress").
			Updates(map[string]interface{}{
				"score":         attempt.Score,
				"correct_count": attempt.CorrectCount,
				"wrong_count":   attempt.WrongCount,
				"status":        "completed",
				"submitted_at":  attempt.SubmittedAt,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		if len(updatedAnswers) > 0 {
			for i := range updatedAnswers {
				updatedAnswers[i].AttemptID = attemptID
				if updatedAnswers[i].UpdatedAt.IsZero() {
					updatedAnswers[i].UpdatedAt = now
				}
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "attempt_id"}, {Name: "question_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"chosen_option", "is_correct", "points_awarded", "updated_at"}),
			}).Create(&updatedAnswers).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *gormExamRepository) GetAttemptAnswers(attemptID uint) ([]models.AttemptAnswer, error) {
	var answers []models.AttemptAnswer
	err := r.db.Where("attempt_id = ?", attemptID).Find(&answers).Error
	return answers, err
}

func (r *gormExamRepository) CreateAuditLog(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *gormExamRepository) ListAuditLogs(limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Preload("Actor").Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
