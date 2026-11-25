package handlers

import (
	"github.com/gin-gonic/gin"
)

func StudentGetAvailableExams(c *gin.Context) {
	getDefaultExamTakingHandler().StudentGetAvailableExams(c)
}

func StudentLookupExam(c *gin.Context) {
	getDefaultExamTakingHandler().StudentLookupExam(c)
}

func StudentStartExam(c *gin.Context) {
	getDefaultExamTakingHandler().StudentStartExam(c)
}

func StudentSaveAnswers(c *gin.Context) {
	getDefaultExamTakingHandler().StudentSaveAnswers(c)
}

func StudentRecordViolation(c *gin.Context) {
	getDefaultExamTakingHandler().StudentRecordViolation(c)
}

func StudentSubmitExam(c *gin.Context) {
	getDefaultExamTakingHandler().StudentSubmitExam(c)
}

func StudentGetResult(c *gin.Context) {
	getDefaultExamTakingHandler().StudentGetResult(c)
}

func StudentListAttempts(c *gin.Context) {
	getDefaultExamTakingHandler().StudentListAttempts(c)
}
