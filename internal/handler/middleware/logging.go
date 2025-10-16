package middleware

import (
	"time"

    "service-short-link/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger middleware for structured logging
func Logger() gin.HandlerFunc {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		if param.StatusCode >= 400 {
			logger.WithFields(logrus.Fields{
				"status_code":  param.StatusCode,
				"latency":      param.Latency,
				"client_ip":    param.ClientIP,
				"method":       param.Method,
				"path":         param.Path,
				"user_agent":   param.Request.UserAgent(),
				"error":        param.ErrorMessage,
				"timestamp":    param.TimeStamp.Format(time.RFC3339),
			}).Error("HTTP request error")
		}
		return ""
	})
}

// Recovery middleware with custom error handling
func Recovery() gin.HandlerFunc {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.WithFields(logrus.Fields{
			"panic":     recovered,
			"path":      c.Request.URL.Path,
			"method":    c.Request.Method,
			"client_ip": c.ClientIP(),
			"timestamp": time.Now().Format(time.RFC3339),
		}).Error("Panic recovered")

		c.JSON(500, gin.H{
			"error":   "Internal Server Error",
			"message": "An unexpected error occurred",
		})
	})
}

// AppErrorLogger logs non-2xx responses to error.log using our file logger
func AppErrorLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        status := c.Writer.Status()
        if status >= 400 {
            if last := c.Errors.Last(); last != nil && last.Err != nil {
                logger.ErrorWithCockroach(last.Err, "HTTP error response", map[string]interface{}{
                    "status":    status,
                    "latency":   time.Since(start).String(),
                    "client_ip": c.ClientIP(),
                    "method":    c.Request.Method,
                    "path":      c.Request.URL.Path,
                })
            } else {
                logger.ErrorWithCockroach(nil, "HTTP error response", map[string]interface{}{
                    "status":    status,
                    "latency":   time.Since(start).String(),
                    "client_ip": c.ClientIP(),
                    "method":    c.Request.Method,
                    "path":      c.Request.URL.Path,
                })
            }
        }
    }
}