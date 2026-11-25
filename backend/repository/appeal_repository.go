package repository

import (
	"strings"
	"time"

	"exam-system-mvp/dto"
	"exam-system-mvp/models"

	"gorm.io/gorm"
)

type AppealRepository interface {
	CreateAppeal(appeal *models.Appeal) error
	FindAppealByID(id uint) (*models.Appeal, error)
	FindAppealBySubmissionAndStudent(submissionID, studentID uint) (*models.Appeal, error)
	FindPendingAppealBySubmissionAndStudent(submissionID, studentID uint) (*models.Appeal, error)
	UpdateAppeal(appeal *models.Appeal) error
	ListAppealsByExam(examID, teacherID uint, status, query string, page, pageSize int) ([]models.Appeal, int64, error)
	ListTeacherAppealsByExam(examID, teacherID uint, status, query string, page, pageSize int) ([]dto.TeacherAppealItem, int64, error)
	GetTeacherAppealDetail(appealID, teacherID uint) (map[string]interface{}, error)
	GetStudentHistory(studentID uint, search, status string, page, pageSize int) ([]map[string]interface{}, int64, error)
	GetSubmissionDetail(submissionID, studentID uint) (*models.ExamAttempt, *models.Exam, []map[string]interface{}, []models.Appeal, float64, error)
}

type gormAppealRepository struct {
	db *gorm.DB
}

func NewAppealRepository(db *gorm.DB) AppealRepository {
	return &gormAppealRepository{db: db}
}

func (r *gormAppealRepository) CreateAppeal(appeal *models.Appeal) error {
	return r.db.Create(appeal).Error
}

func (r *gormAppealRepository) FindAppealByID(id uint) (*models.Appeal, error) {
	var appeal models.Appeal
	if err := r.db.Preload("Student").Preload("Exam").First(&appeal, id).Error; err != nil {
		return nil, err
	}
	return &appeal, nil
}

func (r *gormAppealRepository) FindAppealBySubmissionAndStudent(submissionID, studentID uint) (*models.Appeal, error) {
	var appeal models.Appeal
	if err := r.db.Where("submission_id = ? AND student_id = ?", submissionID, studentID).First(&appeal).Error; err != nil {
		return nil, err
	}
	return &appeal, nil
}

func (r *gormAppealRepository) FindPendingAppealBySubmissionAndStudent(submissionID, studentID uint) (*models.Appeal, error) {
	var appeal models.Appeal
	if err := r.db.Where("submission_id = ? AND student_id = ? AND status = ?", submissionID, studentID, models.AppealStatusPending).First(&appeal).Error; err != nil {
		return nil, err
	}
	return &appeal, nil
}

func (r *gormAppealRepository) UpdateAppeal(appeal *models.Appeal) error {
	return r.db.Save(appeal).Error
}

func (r *gormAppealRepository) ListAppealsByExam(examID, teacherID uint, status, query string, page, pageSize int) ([]models.Appeal, int64, error) {
	db := r.db.Model(&models.Appeal{}).
		Joins("JOIN users u ON u.id = appeals.student_id").
		Where("appeals.exam_id = ? AND appeals.teacher_id = ?", examID, teacherID)

	if status != "" {
		db = db.Where("appeals.status = ?", status)
	}
	if query != "" {
		term := "%" + query + "%"
		db = db.Where("LOWER(appeals.title) LIKE ? OR LOWER(u.full_name) LIKE ? OR LOWER(u.email) LIKE ?", term, term, term)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var appeals []models.Appeal
	err := db.Preload("Student").Preload("Submission").
		Order("appeals.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&appeals).Error
	return appeals, total, err
}

func (r *gormAppealRepository) ListTeacherAppealsByExam(examID, teacherID uint, status, query string, page, pageSize int) ([]dto.TeacherAppealItem, int64, error) {
	var exam models.Exam
	if err := r.db.Where("id = ? AND teacher_id = ?", examID, teacherID).First(&exam).Error; err != nil {
		return nil, 0, err
	}

	db := r.db.Table("appeals ap").
		Joins("JOIN users u ON u.id = ap.student_id").
		Where("ap.exam_id = ? AND ap.teacher_id = ?", examID, teacherID)

	if status != "" {
		db = db.Where("ap.status = ?", status)
	}
	if query != "" {
		like := "%" + strings.ToLower(query) + "%"
		db = db.Where("(LOWER(u.full_name) LIKE ? OR LOWER(ap.title) LIKE ? OR LOWER(ap.content) LIKE ?)", like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := make([]dto.TeacherAppealItem, 0)
	err := db.Select(`
		ap.id, ap.submission_id, ap.student_id, u.full_name AS student_name,
		u.email AS student_email, e.title AS exam_title, ap.title, ap.content,
		ap.status, ap.created_at, ea.score
	`).
		Joins("JOIN exams e ON e.id = ap.exam_id").
		Joins("JOIN exam_attempts ea ON ea.id = ap.submission_id").
		Order("ap.created_at DESC, ap.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *gormAppealRepository) GetTeacherAppealDetail(appealID, teacherID uint) (map[string]interface{}, error) {
	var appeal models.Appeal
	if err := r.db.Where(`
		id = ? AND teacher_id = ?
		AND EXISTS (SELECT 1 FROM exams e WHERE e.id = appeals.exam_id AND e.teacher_id = ?)
	`, appealID, teacherID, teacherID).First(&appeal).Error; err != nil {
		return nil, err
	}

	var student models.User
	if err := r.db.Select("id, full_name, email").First(&student, appeal.StudentID).Error; err != nil {
		return nil, err
	}

	var exam models.Exam
	if err := r.db.First(&exam, appeal.ExamID).Error; err != nil {
		return nil, err
	}

	var attempt models.ExamAttempt
	if err := r.db.Where("id = ? AND exam_id = ? AND user_id = ?", appeal.SubmissionID, appeal.ExamID, appeal.StudentID).
		First(&attempt).Error; err != nil {
		return nil, err
	}

	var maxScore float64
	_ = r.db.Model(&models.Question{}).Where("exam_id = ?", exam.ID).
		Select("COALESCE(SUM(points), 0)").Scan(&maxScore)

	type TeacherAppealAnswer struct {
		QuestionID    uint    `json:"question_id"`
		SortOrder     int     `json:"sort_order"`
		Prompt        string  `json:"prompt"`
		OptionA       string  `json:"option_a"`
		OptionB       string  `json:"option_b"`
		OptionC       string  `json:"option_c"`
		OptionD       string  `json:"option_d"`
		ChosenOption  *string `json:"chosen_option"`
		CorrectOption string  `json:"correct_option"`
		IsCorrect     bool    `json:"is_correct"`
		Points        float64 `json:"points"`
	}

	answers := make([]TeacherAppealAnswer, 0)
	_ = r.db.Table("questions q").
		Select(`
			q.id AS question_id, q.sort_order, q.prompt,
			q.option_a, q.option_b, q.option_c, q.option_d,
			aa.chosen_option, q.correct_option, COALESCE(aa.is_correct, false) AS is_correct, q.points
		`).
		Joins("LEFT JOIN attempt_answers aa ON aa.question_id = q.id AND aa.attempt_id = ?", attempt.ID).
		Where("q.exam_id = ?", exam.ID).
		Order("q.sort_order ASC, q.id ASC").
		Scan(&answers)

	return map[string]interface{}{
		"appeal":  appeal,
		"student": student,
		"exam": map[string]interface{}{
			"id":    exam.ID,
			"title": exam.Title,
		},
		"submission": map[string]interface{}{
			"id":           attempt.ID,
			"score":        attempt.Score,
			"max_score":    maxScore,
			"started_at":   attempt.StartedAt,
			"submitted_at": attempt.SubmittedAt,
		},
		"answers": answers,
	}, nil
}

func (r *gormAppealRepository) GetStudentHistory(studentID uint, search, status string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	query := r.db.Table("exam_attempts ea").
		Joins("JOIN exams e ON e.id = ea.exam_id").
		Where("ea.user_id = ? AND ea.status = ?", studentID, "completed")

	if search != "" {
		query = query.Where("LOWER(e.title) LIKE ?", "%"+strings.ToLower(search)+"%")
	}
	if status != "" {
		query = query.Where(`
			EXISTS (
				SELECT 1 FROM appeals ap
				WHERE ap.submission_id = ea.id AND ap.student_id = ea.user_id AND ap.status = ?
			)
		`, status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type HistoryRow struct {
		SubmissionID uint
		ExamID       uint
		ExamTitle    string
		PublishedAt  *time.Time
		StartedAt    time.Time
		SubmittedAt  *time.Time
		Score        float64
		MaxScore     float64
		AppealStatus string
		AppealCount  int
	}

	var rows []HistoryRow
	err := query.Select(`
		ea.id AS submission_id,
		e.id AS exam_id,
		e.title AS exam_title,
		COALESCE(e.published_at, e.updated_at, e.created_at) AS published_at,
		ea.started_at,
		ea.submitted_at,
		ea.score,
		COALESCE((SELECT SUM(q.points) FROM questions q WHERE q.exam_id = e.id), 0) AS max_score,
		COALESCE((
			SELECT ap.status FROM appeals ap
			WHERE ap.submission_id = ea.id AND ap.student_id = ea.user_id
			ORDER BY ap.created_at DESC, ap.id DESC LIMIT 1
		), '') AS appeal_status,
		(SELECT COUNT(*) FROM appeals ap
		 WHERE ap.submission_id = ea.id AND ap.student_id = ea.user_id) AS appeal_count
	`).
		Order("ea.submitted_at DESC, ea.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	res := make([]map[string]interface{}, 0, len(rows))
	for _, rRow := range rows {
		appealStatus := rRow.AppealStatus
		if appealStatus == "" {
			appealStatus = "NOT_SUBMITTED"
		}
		res = append(res, map[string]interface{}{
			"submission_id": rRow.SubmissionID,
			"exam_id":       rRow.ExamID,
			"exam_title":    rRow.ExamTitle,
			"published_at":  rRow.PublishedAt,
			"started_at":    rRow.StartedAt,
			"submitted_at":  rRow.SubmittedAt,
			"score":         rRow.Score,
			"max_score":     rRow.MaxScore,
			"appeal_status": appealStatus,
			"appeal_count":  rRow.AppealCount,
		})
	}

	return res, total, nil
}

func (r *gormAppealRepository) GetSubmissionDetail(submissionID, studentID uint) (*models.ExamAttempt, *models.Exam, []map[string]interface{}, []models.Appeal, float64, error) {
	var attempt models.ExamAttempt
	if err := r.db.Where("id = ? AND user_id = ? AND status = 'completed'", submissionID, studentID).First(&attempt).Error; err != nil {
		return nil, nil, nil, nil, 0, err
	}

	var exam models.Exam
	if err := r.db.First(&exam, attempt.ExamID).Error; err != nil {
		return nil, nil, nil, nil, 0, err
	}

	type HistoryAnswer struct {
		QuestionID    uint    `json:"question_id"`
		SortOrder     int     `json:"sort_order"`
		Prompt        string  `json:"prompt"`
		OptionA       string  `json:"option_a"`
		OptionB       string  `json:"option_b"`
		OptionC       string  `json:"option_c"`
		OptionD       string  `json:"option_d"`
		ChosenOption  *string `json:"chosen_option"`
		CorrectOption string  `json:"correct_option"`
		IsCorrect     bool    `json:"is_correct"`
		Points        float64 `json:"points"`
		PointsAwarded float64 `json:"points_awarded"`
	}

	var answers []HistoryAnswer
	err := r.db.Table("questions q").
		Select(`
			q.id AS question_id, q.sort_order, q.prompt,
			q.option_a, q.option_b, q.option_c, q.option_d,
			aa.chosen_option, q.correct_option, COALESCE(aa.is_correct, false) AS is_correct,
			q.points, COALESCE(aa.points_awarded, 0) AS points_awarded
		`).
		Joins("LEFT JOIN attempt_answers aa ON aa.question_id = q.id AND aa.attempt_id = ?", attempt.ID).
		Where("q.exam_id = ?", exam.ID).
		Order("q.sort_order ASC, q.id ASC").
		Scan(&answers).Error
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}

	var maxScore float64
	_ = r.db.Model(&models.Question{}).Where("exam_id = ?", exam.ID).
		Select("COALESCE(SUM(points), 0)").Scan(&maxScore)

	var appeals []models.Appeal
	_ = r.db.Where("submission_id = ? AND student_id = ?", attempt.ID, studentID).
		Order("created_at DESC, id DESC").Find(&appeals)

	ansList := make([]map[string]interface{}, 0, len(answers))
	for _, a := range answers {
		ansList = append(ansList, map[string]interface{}{
			"question_id":    a.QuestionID,
			"sort_order":     a.SortOrder,
			"prompt":         a.Prompt,
			"option_a":       a.OptionA,
			"option_b":       a.OptionB,
			"option_c":       a.OptionC,
			"option_d":       a.OptionD,
			"chosen_option":  a.ChosenOption,
			"correct_option": a.CorrectOption,
			"is_correct":     a.IsCorrect,
			"points":         a.Points,
			"points_awarded": a.PointsAwarded,
		})
	}

	return &attempt, &exam, ansList, appeals, maxScore, nil
}
