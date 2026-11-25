package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"

	"exam-system-mvp/apiresponse"
	"exam-system-mvp/observability"

	"github.com/gin-gonic/gin"
)

var traceIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{8,64}$`)

func RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if !traceIDPattern.MatchString(traceID) {
			traceID = newTraceID()
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)

		started := time.Now()
		c.Next()

		status := c.Writer.Status()
		attrs := []any{
			"trace_id", traceID, "method", c.Request.Method, "path", c.Request.URL.Path,
			"status", status, "latency_ms", time.Since(started).Milliseconds(), "client_ip", c.ClientIP(),
		}
		if status >= http.StatusInternalServerError {
			observability.Error("http request completed", attrs...)
		} else if status >= http.StatusBadRequest {
			observability.Warn("http request completed", attrs...)
		} else {
			observability.Info("http request completed", attrs...)
		}
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				traceID, _ := c.Get("trace_id")
				observability.Error("request panic recovered", "trace_id", traceID, "panic", fmt.Sprint(recovered), "stack", string(debug.Stack()))
				if !c.Writer.Written() {
					apiresponse.Error(c, http.StatusInternalServerError, "Unexpected server error")
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}

func NotFound(c *gin.Context) {
	apiresponse.Error(c, http.StatusNotFound, "The requested route was not found.")
}
func MethodNotAllowed(c *gin.Context) {
	apiresponse.Error(c, http.StatusMethodNotAllowed, "The HTTP method is not allowed.")
}

func newTraceID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprintf("trace-%d", time.Now().UnixNano())
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(value[:])
	return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:]
}
