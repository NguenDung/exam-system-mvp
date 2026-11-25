package handlers

import (
	"github.com/gin-gonic/gin"
)

func TeacherUpdateExam(c *gin.Context) {
	getDefaultExamHandler().TeacherUpdateExam(c)
}

func TeacherGetExam(c *gin.Context) {
	getDefaultExamHandler().TeacherGetExam(c)
}

func TeacherDeleteExam(c *gin.Context) {
	getDefaultExamHandler().TeacherDeleteExam(c)
}

func TeacherAddQuestions(c *gin.Context) {
	getDefaultExamHandler().TeacherAddQuestions(c)
}

func TeacherBulkAddQuestions(c *gin.Context) {
	getDefaultExamHandler().TeacherBulkAddQuestions(c)
}

func TeacherPublishExam(c *gin.Context) {
	getDefaultExamHandler().TeacherPublishExam(c)
}

func TeacherGetResults(c *gin.Context) {
	getDefaultExamHandler().TeacherGetResults(c)
}

func TeacherExportResultsCSV(c *gin.Context) {
	getDefaultExamHandler().TeacherExportResultsCSV(c)
}

func TeacherUpdateQuestion(c *gin.Context) {
	getDefaultExamHandler().TeacherUpdateQuestion(c)
}

func TeacherDeleteQuestion(c *gin.Context) {
	getDefaultExamHandler().TeacherDeleteQuestion(c)
}

func TeacherUpdateAttemptScore(c *gin.Context) {
	getDefaultExamHandler().TeacherUpdateAttemptScore(c)
}

func TeacherResetAttempt(c *gin.Context) {
	getDefaultExamHandler().TeacherResetAttempt(c)
}

func TeacherListStudents(c *gin.Context) {
	getDefaultExamHandler().TeacherListStudents(c)
}

func TeacherCreateStudent(c *gin.Context) {
	getDefaultExamHandler().TeacherCreateStudent(c)
}

func TeacherUpdateStudent(c *gin.Context) {
	getDefaultExamHandler().TeacherUpdateStudent(c)
}

func TeacherDeleteStudent(c *gin.Context) {
	getDefaultExamHandler().TeacherDeleteStudent(c)
}
