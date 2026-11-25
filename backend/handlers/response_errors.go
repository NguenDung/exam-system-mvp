package handlers

import (
	"errors"
	"net/http"

	"exam-system-mvp/apiresponse"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func lookupFailure(c *gin.Context, err error, missingStatus int, missingMessage, operation string) {
	if err == nil {
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		apiresponse.Error(c, missingStatus, missingMessage)
		return
	}
	apiresponse.Error(c, http.StatusInternalServerError, operation, err)
}
