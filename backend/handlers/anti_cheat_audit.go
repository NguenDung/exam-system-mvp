package handlers

import (
	"github.com/gin-gonic/gin"
)

func TeacherGetAuditLogs(c *gin.Context) {
	getDefaultExamHandler().TeacherGetAuditLogs(c)
}
