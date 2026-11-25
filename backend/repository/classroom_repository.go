package repository

import (
	"exam-system-mvp/models"

	"gorm.io/gorm"
)

type ClassroomRepository interface {
	Create(class *models.Classroom) error
	FindByID(id uint) (*models.Classroom, error)
	FindByIDWithTeacher(id uint) (*models.Classroom, error)
	FindByCode(code string) (*models.Classroom, error)
	Update(class *models.Classroom, updates map[string]interface{}) error
	Delete(id uint) error
	FindByTeacherID(teacherID uint) ([]models.Classroom, error)
	FindEnrolledByStudentID(studentID uint) ([]models.Classroom, error)
	FindMember(classID, studentID uint) (*models.ClassroomMember, error)
	AddMember(member *models.ClassroomMember) error
	RemoveMember(classID, studentID uint) (int64, error)
	GetMembersWithStudent(classID uint) ([]models.ClassroomMember, error)
	GetClassExams(classID uint) ([]models.Exam, error)
	GetClassAttempts(examIDs []uint) ([]models.ExamAttempt, error)
	CountMembers(classID uint) (int64, error)
	CountExams(classID uint) (int64, error)
	IsCodeExists(code string) (bool, error)
}

type gormClassroomRepository struct {
	db *gorm.DB
}

func NewClassroomRepository(db *gorm.DB) ClassroomRepository {
	return &gormClassroomRepository{db: db}
}

func (r *gormClassroomRepository) Create(class *models.Classroom) error {
	return r.db.Create(class).Error
}

func (r *gormClassroomRepository) FindByID(id uint) (*models.Classroom, error) {
	var class models.Classroom
	if err := r.db.First(&class, id).Error; err != nil {
		return nil, err
	}
	return &class, nil
}

func (r *gormClassroomRepository) FindByIDWithTeacher(id uint) (*models.Classroom, error) {
	var class models.Classroom
	if err := r.db.Preload("Teacher").First(&class, id).Error; err != nil {
		return nil, err
	}
	return &class, nil
}

func (r *gormClassroomRepository) FindByCode(code string) (*models.Classroom, error) {
	var class models.Classroom
	if err := r.db.Where("code = ?", code).First(&class).Error; err != nil {
		return nil, err
	}
	return &class, nil
}

func (r *gormClassroomRepository) Update(class *models.Classroom, updates map[string]interface{}) error {
	return r.db.Model(class).Updates(updates).Error
}

func (r *gormClassroomRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("classroom_id = ?", id).Delete(&models.ClassroomMember{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Exam{}).Where("classroom_id = ?", id).Update("classroom_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Classroom{}, id).Error
	})
}

func (r *gormClassroomRepository) FindByTeacherID(teacherID uint) ([]models.Classroom, error) {
	var classes []models.Classroom
	err := r.db.Where("teacher_id = ?", teacherID).Order("created_at DESC").Find(&classes).Error
	return classes, err
}

func (r *gormClassroomRepository) FindEnrolledByStudentID(studentID uint) ([]models.Classroom, error) {
	var classIDs []uint
	r.db.Model(&models.ClassroomMember{}).Where("student_id = ?", studentID).Pluck("classroom_id", &classIDs)
	if len(classIDs) == 0 {
		return []models.Classroom{}, nil
	}
	var classes []models.Classroom
	err := r.db.Where("id IN ?", classIDs).Preload("Teacher").Order("created_at DESC").Find(&classes).Error
	return classes, err
}

func (r *gormClassroomRepository) FindMember(classID, studentID uint) (*models.ClassroomMember, error) {
	var member models.ClassroomMember
	if err := r.db.Where("classroom_id = ? AND student_id = ?", classID, studentID).First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *gormClassroomRepository) AddMember(member *models.ClassroomMember) error {
	return r.db.Create(member).Error
}

func (r *gormClassroomRepository) RemoveMember(classID, studentID uint) (int64, error) {
	res := r.db.Where("classroom_id = ? AND student_id = ?", classID, studentID).Delete(&models.ClassroomMember{})
	return res.RowsAffected, res.Error
}

func (r *gormClassroomRepository) GetMembersWithStudent(classID uint) ([]models.ClassroomMember, error) {
	var members []models.ClassroomMember
	err := r.db.Where("classroom_id = ?", classID).
		Preload("Student").
		Order("joined_at DESC").
		Find(&members).Error
	return members, err
}

func (r *gormClassroomRepository) GetClassExams(classID uint) ([]models.Exam, error) {
	var exams []models.Exam
	err := r.db.Where("classroom_id = ?", classID).Order("start_time DESC, id DESC").Find(&exams).Error
	return exams, err
}

func (r *gormClassroomRepository) GetClassAttempts(examIDs []uint) ([]models.ExamAttempt, error) {
	if len(examIDs) == 0 {
		return []models.ExamAttempt{}, nil
	}
	var attempts []models.ExamAttempt
	err := r.db.Where("exam_id IN ?", examIDs).Find(&attempts).Error
	return attempts, err
}

func (r *gormClassroomRepository) CountMembers(classID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ClassroomMember{}).Where("classroom_id = ?", classID).Count(&count).Error
	return count, err
}

func (r *gormClassroomRepository) CountExams(classID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Exam{}).Where("classroom_id = ?", classID).Count(&count).Error
	return count, err
}

func (r *gormClassroomRepository) IsCodeExists(code string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Classroom{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}
