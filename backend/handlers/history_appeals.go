package handlers

import (
	"github.com/gin-gonic/gin"
)

func StudentGetHistory(c *gin.Context) {
	getDefaultAppealHandler().StudentGetHistory(c)
}

func StudentGetHistoryDetail(c *gin.Context) {
	getDefaultAppealHandler().StudentGetHistoryDetail(c)
}

func StudentCreateAppeal(c *gin.Context) {
	getDefaultAppealHandler().StudentCreateAppeal(c)
}

func TeacherListExamAppeals(c *gin.Context) {
	getDefaultAppealHandler().TeacherListExamAppeals(c)
}

func TeacherGetAppeal(c *gin.Context) {
	getDefaultAppealHandler().TeacherGetAppeal(c)
}

func TeacherReplyAppeal(c *gin.Context) {
	getDefaultAppealHandler().TeacherReplyAppeal(c)
}

func TeacherRejectAppeal(c *gin.Context) {
	getDefaultAppealHandler().TeacherRejectAppeal(c)
}

func TeacherResolveAppeal(c *gin.Context) {
	getDefaultAppealHandler().TeacherResolveAppeal(c)
}
