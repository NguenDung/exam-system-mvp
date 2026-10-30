package handlers

import (
	"strconv"
	"net/http"

	"github.com/gin-gonic/gin"
)

func parseUintParam(c *gin.Context, param string) (uint, bool) {
	v, err := strconv.ParseUint(c.Param(param), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid " + param})
		return 0, false
	}
	return uint(v), true
}

func TeacherListExams(c *gin.Context) {
	getDefaultExamHandler().TeacherListExams(c)
}

func TeacherCreateExam(c *gin.Context) {
	getDefaultExamHandler().TeacherCreateExam(c)
}
