package handler

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "service-short-link/pkg/logger"
)

// SuccessEnvelope is the standard shape for successful responses
type SuccessEnvelope struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Meta    interface{} `json:"meta,omitempty"`
}

// ErrorEnvelope is the standard shape for error responses
type ErrorEnvelope struct {
    Success bool          `json:"success"`
    Error   ErrorPayload  `json:"error"`
    Meta    interface{}   `json:"meta,omitempty"`
}

// ErrorPayload carries structured error information
type ErrorPayload struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

// RespondOK sends a 200 response with the standard envelope
func RespondOK(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, SuccessEnvelope{
        Success: true,
        Data:    data,
    })
}

// RespondCreated sends a 201 response with the standard envelope
func RespondCreated(c *gin.Context, data interface{}) {
    c.JSON(http.StatusCreated, SuccessEnvelope{
        Success: true,
        Data:    data,
    })
}

// RespondNoContent sends a 204 response
func RespondNoContent(c *gin.Context) {
    c.Status(http.StatusNoContent)
}

// RespondError sends an error response with the standard envelope
func RespondError(c *gin.Context, httpStatus int, code, message string, details interface{}) {
    var detailsStr string
    if details != nil {
        detailsStr = fmt.Sprintf("%v", details)
    }
    
    errorMsg := fmt.Sprintf("HTTP %d: %s - %s", httpStatus, code, message)
    if detailsStr != "" {
        errorMsg += fmt.Sprintf(" (Details: %s)", detailsStr)
    }
    
    logger.ErrorWithCockroachSimple(
        fmt.Errorf("%s", message),
        errorMsg,
        "http_status", fmt.Sprintf("%d", httpStatus),
        "error_code", code,
        "endpoint", c.Request.URL.Path,
        "method", c.Request.Method,
        "client_ip", c.ClientIP(),
    )
    
    c.JSON(httpStatus, ErrorEnvelope{
        Success: false,
        Error: ErrorPayload{
            Code:    code,
            Message: message,
            Details: details,
        },
    })
}
