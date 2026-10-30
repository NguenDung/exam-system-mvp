package handlers

import (
	"github.com/gin-gonic/gin"
)

func GetExamSchedule(c *gin.Context) {
	getDefaultExamTakingHandler().GetExamSchedule(c)
}
