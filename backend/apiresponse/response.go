package apiresponse

import (
	"errors"
	"net/http"
	"strings"

	"exam-system-mvp/observability"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Code          string   `json:"code"`
	Error         string   `json:"error"`
	TraceID       string   `json:"trace_id"`
	Details       []string `json:"details,omitempty"`
	AutoSubmitted bool     `json:"auto_submitted,omitempty"`
	Attempt       any      `json:"attempt,omitempty"`
}

type ErrorData struct {
	AutoSubmitted bool
	Attempt       any
}

// JSON keeps response serialization behind one helper so handlers can focus
// on assembling typed DTOs and business logic.
func JSON(c *gin.Context, status int, payload any) { c.JSON(status, payload) }

func SuccessMessage(c *gin.Context, code string) string {
	messages, ok := successMessages[code]
	if !ok {
		return ""
	}
	if localeFor(c.GetHeader("Accept-Language")) == "vi" {
		return messages[1]
	}
	return messages[0]
}

// Error returns a localized public message and keeps server diagnostics in
// structured logs. Internal error details are never exposed in the response.
func Error(c *gin.Context, status int, message string, cause ...error) {
	writeError(c, status, message, ErrorData{}, cause...)
}

func ErrorWithData(c *gin.Context, status int, message string, data ErrorData) {
	writeError(c, status, message, data)
}

func writeError(c *gin.Context, status int, message string, data ErrorData, cause ...error) {
	code := codeForMessage(status, message)
	traceID, _ := c.Get("trace_id")
	trace, _ := traceID.(string)
	locale := localeFor(c.GetHeader("Accept-Language"))

	if status >= http.StatusInternalServerError {
		attrs := []any{"trace_id", trace, "status", status, "code", code, "message", message}
		if len(cause) > 0 && cause[0] != nil {
			attrs = append(attrs, "error", cause[0].Error())
		}
		observability.Error("request failed", attrs...)
	} else if status >= http.StatusBadRequest {
		observability.Warn("request rejected", "trace_id", trace, "status", status, "code", code, "reason", message)
	}

	public := translatedMessage(message, code, locale)
	JSON(c, status, ErrorResponse{
		Code: code, Error: public, TraceID: trace,
		AutoSubmitted: data.AutoSubmitted, Attempt: data.Attempt,
	})
}

// Validation returns field-level guidance without exposing framework or parser internals.
func Validation(c *gin.Context, err error) {
	details := make([]string, 0)
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, fieldError := range validationErrors {
			details = append(details, validationDetail(fieldError.Field(), fieldError.Tag(), localeFor(c.GetHeader("Accept-Language"))))
		}
	}
	traceID, _ := c.Get("trace_id")
	trace, _ := traceID.(string)
	locale := localeFor(c.GetHeader("Accept-Language"))
	errorMessage := messageFor("VALIDATION_ERROR", locale)
	if len(details) > 0 {
		errorMessage += " " + strings.Join(details, " ")
	}
	observability.Warn("request validation failed", "trace_id", trace, "status", http.StatusBadRequest, "code", "VALIDATION_ERROR", "fields", details)
	JSON(c, http.StatusBadRequest, ErrorResponse{
		Code: "VALIDATION_ERROR", Error: errorMessage, TraceID: trace, Details: details,
	})
}

func codeForStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusPreconditionRequired:
		return "PRECONDITION_REQUIRED"
	case http.StatusMethodNotAllowed:
		return "METHOD_NOT_ALLOWED"
	case http.StatusBadGateway:
		return "BAD_GATEWAY"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	case http.StatusNotImplemented:
		return "NOT_IMPLEMENTED"
	default:
		if status >= http.StatusInternalServerError {
			return "INTERNAL_SERVER_ERROR"
		}
		return "REQUEST_ERROR"
	}
}

func codeForMessage(status int, message string) string {
	if status >= http.StatusInternalServerError {
		return codeForStatus(status)
	}
	if strings.Contains(message, "does not belong to this exam") {
		return "ANSWER_QUESTION_MISMATCH"
	}
	if code, ok := publicMessageCodes[message]; ok {
		return code
	}
	if strings.HasPrefix(message, "Invalid ") {
		return "INVALID_PARAMETER"
	}
	return codeForStatus(status)
}

func localeFor(header string) string {
	for _, part := range strings.Split(header, ",") {
		lang := strings.ToLower(strings.TrimSpace(strings.SplitN(part, ";", 2)[0]))
		if strings.HasPrefix(lang, "vi") {
			return "vi"
		}
		if strings.HasPrefix(lang, "en") {
			return "en"
		}
	}
	return "en"
}

func translatedMessage(original, code, locale string) string {
	switch code {
	case "INTERNAL_SERVER_ERROR", "BAD_GATEWAY", "NOT_IMPLEMENTED", "SERVICE_UNAVAILABLE":
		return messageFor(code, locale)
	}
	if locale == "en" && original != "" {
		if translated, ok := enPublicMessages[code]; ok {
			return translated
		}
		return original
	}
	if locale == "vi" {
		if translated, ok := viPublicMessages[code]; ok {
			return translated
		}
		if strings.Contains(original, "does not belong to this exam") {
			return "Có câu hỏi không thuộc kỳ thi này."
		}
	}
	return messageFor(code, locale)
}

func messageFor(code, locale string) string {
	messages := map[string][2]string{
		"BAD_REQUEST":              {"Invalid request.", "Yêu cầu không hợp lệ."},
		"VALIDATION_ERROR":         {"Please check the submitted fields.", "Vui lòng kiểm tra lại các trường dữ liệu."},
		"UNAUTHORIZED":             {"Authentication is required or has expired.", "Thông tin xác thực không hợp lệ hoặc đã hết hạn."},
		"FORBIDDEN":                {"You do not have permission to perform this action.", "Bạn không có quyền thực hiện thao tác này."},
		"NOT_FOUND":                {"The requested resource was not found.", "Không tìm thấy dữ liệu được yêu cầu."},
		"CONFLICT":                 {"This request conflicts with the current resource state.", "Yêu cầu bị xung đột với trạng thái hiện tại."},
		"PRECONDITION_REQUIRED":    {"Confirm the required action before continuing.", "Vui lòng xác nhận trước khi tiếp tục."},
		"METHOD_NOT_ALLOWED":       {"This method is not supported for this resource.", "Phương thức này không được hỗ trợ."},
		"BAD_GATEWAY":              {"The service is temporarily unavailable.", "Dịch vụ tạm thời không khả dụng."},
		"SERVICE_UNAVAILABLE":      {"The service is temporarily unavailable.", "Dịch vụ tạm thời không khả dụng."},
		"NOT_IMPLEMENTED":          {"This operation is not available.", "Thao tác này chưa được hỗ trợ."},
		"INTERNAL_SERVER_ERROR":    {"An unexpected error occurred. Please try again later.", "Đã xảy ra lỗi hệ thống. Vui lòng thử lại sau."},
		"REQUEST_ERROR":            {"The request could not be completed.", "Không thể hoàn tất yêu cầu."},
		"INVALID_PARAMETER":        {"A request parameter is invalid.", "Tham số yêu cầu không hợp lệ."},
		"ANSWER_QUESTION_MISMATCH": {"One or more questions do not belong to this exam.", "Có câu hỏi không thuộc kỳ thi này."},
		"EXAM_TIME_UP":             {"Time is up. Your exam has been automatically submitted.", "Đã hết giờ. Bài thi đã được tự động nộp."},
	}
	messagesForCode, ok := messages[code]
	if !ok {
		messagesForCode = messages["REQUEST_ERROR"]
	}
	if locale == "vi" {
		return messagesForCode[1]
	}
	return messagesForCode[0]
}

func validationDetail(field, tag, locale string) string {
	field = strings.ToLower(field)
	if locale == "vi" {
		fieldNames := map[string]string{"fullname": "Họ và tên", "email": "Email", "password": "Mật khẩu", "title": "Tiêu đề", "duration": "Thời lượng", "starttime": "Thời gian bắt đầu", "endtime": "Thời gian kết thúc", "prompt": "Nội dung câu hỏi", "optiona": "Lựa chọn A", "optionb": "Lựa chọn B", "optionc": "Lựa chọn C", "optiond": "Lựa chọn D", "correctoption": "Đáp án đúng", "questions": "Danh sách câu hỏi", "questionid": "Mã câu hỏi", "chosenoption": "Lựa chọn", "submissionid": "Mã bài làm", "content": "Nội dung", "response": "Phản hồi", "sortorder": "Thứ tự câu hỏi"}
		name := fieldNames[field]
		if name == "" {
			name = field
		}
		switch tag {
		case "required":
			return name + " là bắt buộc."
		case "email":
			return name + " không đúng định dạng email."
		case "min":
			return name + " chưa đạt độ dài hoặc giá trị tối thiểu."
		case "max":
			return name + " vượt quá độ dài hoặc giá trị tối đa."
		case "gte":
			return name + " nhỏ hơn giá trị tối thiểu."
		case "lte":
			return name + " vượt quá giá trị tối đa."
		case "oneof":
			return name + " có giá trị không hợp lệ."
		case "dive":
			return name + " chứa phần tử không hợp lệ."
		default:
			return name + " không hợp lệ."
		}
	}
	switch tag {
	case "required":
		return field + " is required."
	case "email":
		return field + " must be a valid email address."
	case "min":
		return field + " is below the minimum value or length."
	case "max":
		return field + " exceeds the maximum value or length."
	case "gte":
		return field + " is below the minimum value."
	case "lte":
		return field + " exceeds the maximum value."
	case "oneof":
		return field + " has an unsupported value."
	default:
		return field + " is invalid."
	}
}

var publicMessageCodes = map[string]string{
	"Invalid email format":                                    "INVALID_EMAIL_FORMAT",
	"Email already registered":                                "EMAIL_ALREADY_REGISTERED",
	"Invalid email or password":                               "INVALID_CREDENTIALS",
	"Exam not found":                                          "EXAM_NOT_FOUND",
	"Exam not found or not published":                         "EXAM_UNAVAILABLE",
	"Attempt not found":                                       "ATTEMPT_NOT_FOUND",
	"Submission not found":                                    "SUBMISSION_NOT_FOUND",
	"Completed submission not found":                          "COMPLETED_SUBMISSION_NOT_FOUND",
	"Appeal not found":                                        "APPEAL_NOT_FOUND",
	"Exam has not started yet":                                "EXAM_NOT_STARTED",
	"Exam has expired":                                        "EXAM_EXPIRED",
	"Confirm participation before starting the exam":          "EXAM_CONFIRMATION_REQUIRED",
	"You have already submitted this exam":                    "EXAM_ALREADY_SUBMITTED",
	"Exam already submitted":                                  "EXAM_ALREADY_SUBMITTED",
	"Exam not yet submitted":                                  "EXAM_NOT_SUBMITTED",
	"No answers provided":                                     "ANSWERS_REQUIRED",
	"Invalid appeal_status":                                   "INVALID_APPEAL_STATUS",
	"Invalid status":                                          "INVALID_STATUS",
	"Access denied":                                           "ACCESS_DENIED",
	"Access denied: insufficient role":                        "ACCESS_DENIED",
	"Authorization header required":                           "AUTHORIZATION_REQUIRED",
	"Invalid Authorization header format":                     "AUTHORIZATION_FORMAT_INVALID",
	"Invalid or expired token":                                "TOKEN_INVALID_OR_EXPIRED",
	"end_time must be after start_time":                       "EXAM_TIME_RANGE_INVALID",
	"Cannot add questions to a published exam":                "EXAM_ALREADY_PUBLISHED",
	"Cannot delete questions from a published exam":           "EXAM_ALREADY_PUBLISHED",
	"Exam is already published":                               "EXAM_ALREADY_PUBLISHED",
	"Exam must have at least 1 question before publishing":    "EXAM_QUESTIONS_REQUIRED",
	"Query parameter 'code' is required":                      "EXAM_CODE_REQUIRED",
	"Password must be at most 72 bytes":                       "PASSWORD_TOO_LONG",
	"Full name must contain between 2 and 120 characters":     "FULL_NAME_LENGTH_INVALID",
	"Exam title cannot be blank":                              "EXAM_TITLE_REQUIRED",
	"Question prompt and options cannot be blank":             "QUESTION_CONTENT_REQUIRED",
	"content must contain between 10 and 1000 characters":     "APPEAL_CONTENT_LENGTH_INVALID",
	"response must contain between 1 and 2000 characters":     "APPEAL_RESPONSE_LENGTH_INVALID",
	"Bạn đã có khiếu nại đang chờ xử lý cho bài thi này.":     "PENDING_APPEAL_EXISTS",
	"appeal is already closed":                                "APPEAL_ALREADY_CLOSED",
	"Time is up. Your exam has been automatically submitted.": "EXAM_TIME_UP",
}

var viPublicMessages = map[string]string{
	"INVALID_EMAIL_FORMAT":           "Email không đúng định dạng.",
	"EMAIL_ALREADY_REGISTERED":       "Email này đã được đăng ký.",
	"INVALID_CREDENTIALS":            "Email hoặc mật khẩu không chính xác.",
	"EXAM_NOT_FOUND":                 "Không tìm thấy kỳ thi.",
	"EXAM_UNAVAILABLE":               "Không tìm thấy kỳ thi hoặc kỳ thi chưa được mở.",
	"ATTEMPT_NOT_FOUND":              "Không tìm thấy lượt làm bài.",
	"SUBMISSION_NOT_FOUND":           "Không tìm thấy bài làm.",
	"COMPLETED_SUBMISSION_NOT_FOUND": "Không tìm thấy bài làm đã hoàn thành.",
	"APPEAL_NOT_FOUND":               "Không tìm thấy khiếu nại.",
	"EXAM_NOT_STARTED":               "Kỳ thi chưa bắt đầu.",
	"EXAM_EXPIRED":                   "Kỳ thi đã kết thúc.",
	"EXAM_CONFIRMATION_REQUIRED":     "Vui lòng xác nhận tham gia trước khi bắt đầu.",
	"EXAM_ALREADY_SUBMITTED":         "Bài thi đã được nộp.",
	"EXAM_NOT_SUBMITTED":             "Bài thi chưa được nộp.",
	"ANSWERS_REQUIRED":               "Vui lòng gửi ít nhất một câu trả lời.",
	"INVALID_APPEAL_STATUS":          "Trạng thái khiếu nại không hợp lệ.",
	"INVALID_STATUS":                 "Trạng thái không hợp lệ.",
	"ACCESS_DENIED":                  "Bạn không có quyền truy cập dữ liệu này.",
	"AUTHORIZATION_REQUIRED":         "Vui lòng đăng nhập để tiếp tục.",
	"AUTHORIZATION_FORMAT_INVALID":   "Thông tin xác thực không đúng định dạng.",
	"TOKEN_INVALID_OR_EXPIRED":       "Phiên đăng nhập không hợp lệ hoặc đã hết hạn.",
	"EXAM_TIME_RANGE_INVALID":        "Thời gian kết thúc phải sau thời gian bắt đầu.",
	"EXAM_ALREADY_PUBLISHED":         "Đề thi đã được mở và không thể chỉnh sửa.",
	"EXAM_QUESTIONS_REQUIRED":        "Cần có ít nhất một câu hỏi trước khi mở đề thi.",
	"EXAM_CODE_REQUIRED":             "Vui lòng nhập mã kỳ thi.",
	"PASSWORD_TOO_LONG":              "Mật khẩu không được dài quá 72 byte.",
	"FULL_NAME_LENGTH_INVALID":       "Họ và tên phải có từ 2 đến 120 ký tự.",
	"EXAM_TITLE_REQUIRED":            "Vui lòng nhập tiêu đề đề thi.",
	"QUESTION_CONTENT_REQUIRED":      "Nội dung câu hỏi và cả bốn lựa chọn đều bắt buộc.",
	"APPEAL_CONTENT_LENGTH_INVALID":  "Nội dung khiếu nại phải có từ 10 đến 1000 ký tự.",
	"APPEAL_RESPONSE_LENGTH_INVALID": "Phản hồi phải có từ 1 đến 2000 ký tự.",
	"PENDING_APPEAL_EXISTS":          "Bạn đã có khiếu nại đang chờ xử lý cho bài thi này.",
	"APPEAL_ALREADY_CLOSED":          "Khiếu nại này đã được xử lý.",
	"EXAM_TIME_UP":                   "Đã hết giờ. Bài thi đã được tự động nộp.",
}

var enPublicMessages = map[string]string{
	"PENDING_APPEAL_EXISTS": "You already have a pending appeal for this exam.",
}

var successMessages = map[string][2]string{
	"USER_REGISTERED":    {"Registration successful.", "Đăng ký thành công."},
	"QUESTIONS_CREATED":  {"Questions added successfully.", "Đã thêm câu hỏi thành công."},
	"QUESTION_DELETED":   {"Question deleted successfully.", "Đã xóa câu hỏi thành công."},
	"EXAM_PUBLISHED":     {"Exam published successfully.", "Đã mở đề thi thành công."},
	"EXAM_AUTOSUBMITTED": {"Time is up. Your exam has been automatically submitted.", "Đã hết giờ. Bài thi đã được tự động nộp."},
	"ANSWERS_SAVED":      {"Answers saved successfully.", "Đã lưu câu trả lời thành công."},
	"EXAM_SUBMITTED":     {"Exam submitted and graded successfully.", "Đã nộp và chấm điểm bài thi thành công."},
	"APPEAL_CREATED":     {"Appeal sent successfully.", "Gửi khiếu nại thành công."},
}
