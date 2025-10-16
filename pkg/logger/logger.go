package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
    "runtime"
    "runtime/debug"
	"time"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
	cockroachErrors "github.com/cockroachdb/errors"
)

var (
	ErrorFile *os.File
	DebugFile *os.File
)

// InitLogger initializes file-based logging for error and debug logs
func InitLogger() error {
	if err := os.MkdirAll("/app/logs", 0755); err != nil {
		return err
	}

	log.SetOutput(io.Discard)

	return nil
}

// Debug logs debug messages to debug.log file
func Debug(msg string, fields map[string]interface{}) {
	debugLog := &lumberjack.Logger{
		Filename:   filepath.Join("/app/logs", "debug.log"),
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     7, // days
		Compress:   true,
	}

	logLine := fmt.Sprintf("[DEBUG] %s %s", time.Now().Format(time.RFC3339), msg)
	if fields != nil {
		logLine += fmt.Sprintf(" %+v", fields)
	}
	logLine += "\n"
	
	debugLog.Write([]byte(logLine))
	debugLog.Close()
}

// Info logs info messages to debug.log file
func Info(msg string, fields map[string]interface{}) {
	Debug("[INFO] "+msg, fields)
}

// Warn logs warning messages to debug.log file
func Warn(msg string, fields map[string]interface{}) {
	Debug("[WARN] "+msg, fields)
}

// Error logs error messages to error.log file
func Error(msg string, fields map[string]interface{}) {
	if err := os.MkdirAll("/app/logs", 0755); err != nil {
		Debug("[ERROR] "+msg, fields)
		return
	}

	errorLogPath := filepath.Join("/app/logs", "error.log")
	
	file, err := os.OpenFile(errorLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		Debug("[ERROR] Failed to open error.log, writing to debug.log instead", map[string]interface{}{
			"original_msg": msg,
			"original_fields": fields,
			"error": err.Error(),
		})
		return
	}
	defer file.Close()

	timestamp := time.Now().Format(time.RFC3339)
	logLine := fmt.Sprintf("[ERROR] %s | %s", timestamp, msg)
	
	if fields != nil {
		if location, ok := fields["location"].(string); ok {
			logLine += fmt.Sprintf(" | Location: %s", location)
		}
		
		if function, ok := fields["function"].(string); ok {
			logLine += fmt.Sprintf(" | Function: %s", function)
		}
		
		if errorType, ok := fields["error_type"].(string); ok {
			logLine += fmt.Sprintf(" | ErrorType: %s", errorType)
		}
		
		otherFields := make(map[string]interface{})
		for k, v := range fields {
			if k != "location" && k != "function" && k != "error_type" && k != "stack_trace" && k != "file" && k != "line" && k != "full_path" && k != "full_function" {
				otherFields[k] = v
			}
		}
		
		if len(otherFields) > 0 {
			logLine += fmt.Sprintf(" | Context: %+v", otherFields)
		}
		
		if stackTrace, ok := fields["stack_trace"].(string); ok {
			logLine += fmt.Sprintf("\nStack Trace:\n%s", stackTrace)
		}
	}
	
	logLine += "\n" + strings.Repeat("-", 80) + "\n"
	
	if _, err := file.Write([]byte(logLine)); err != nil {
		Debug("[ERROR] Failed to write to error.log, writing to debug.log instead", map[string]interface{}{
			"original_msg": msg,
			"original_fields": fields,
			"error": err.Error(),
		})
	}
}

// ErrorWithTrace logs error with file, line, function and stack trace
func ErrorWithTrace(err error, msg string, fields map[string]interface{}) {
    if fields == nil {
        fields = map[string]interface{}{}
    }
	
    if pc, file, line, ok := runtime.Caller(1); ok {
        fn := runtime.FuncForPC(pc)
        fileName := filepath.Base(file)
        
        fields["location"] = fmt.Sprintf("%s:%d", fileName, line)
        fields["file"] = fileName
        fields["line"] = line
        fields["full_path"] = file
        
        if fn != nil {
            funcName := fn.Name()
            if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
                funcName = funcName[lastSlash+1:]
            }
            fields["function"] = funcName
            fields["full_function"] = fn.Name()
        }
    }
    
    if err != nil {
        fields["error"] = err.Error()
        fields["error_type"] = fmt.Sprintf("%T", err)
    }
    
    stack := string(debug.Stack())
    fields["stack_trace"] = stack
    
    readableMsg := msg
    if err != nil {
        readableMsg = fmt.Sprintf("%s: %s", msg, err.Error())
    }
    
    Error(readableMsg, fields)
}

// ErrorWithCockroach logs error using CockroachDB errors with automatic stack trace
func ErrorWithCockroach(err error, msg string, fields map[string]interface{}) {
	if err == nil {
		return
	}

	wrappedErr := cockroachErrors.Wrap(err, msg)
	
	if fields != nil {
		if hint, ok := fields["hint"].(string); ok {
			wrappedErr = cockroachErrors.WithHint(wrappedErr, hint)
		}
		
		if detail, ok := fields["detail"].(string); ok {
			wrappedErr = cockroachErrors.WithDetail(wrappedErr, detail)
		}
		
		otherFields := make(map[string]interface{})
		for k, v := range fields {
			if k != "hint" && k != "detail" {
				otherFields[k] = v
			}
		}
		
		if len(otherFields) > 0 {
			contextMsg := buildContextMessage(otherFields)
			wrappedErr = cockroachErrors.WithDetail(wrappedErr, contextMsg)
		}
	}
	
	writeCockroachErrorToFile(fmt.Sprintf("%+v", wrappedErr))
}

// ErrorWithCockroachSimple logs error using CockroachDB errors (simplified version)
func ErrorWithCockroachSimple(err error, msg string, context ...string) {
	if err == nil {
		return
	}

	wrappedErr := cockroachErrors.Wrap(err, msg)
	
	if len(context) > 0 {
		contextMsg := strings.Join(context, ", ")
		wrappedErr = cockroachErrors.WithDetail(wrappedErr, contextMsg)
	}
	
	writeCockroachErrorToFile(fmt.Sprintf("%+v", wrappedErr))
}

// writeCockroachErrorToFile writes CockroachDB formatted error directly to error.log
func writeCockroachErrorToFile(formattedError string) {
	if err := os.MkdirAll("/app/logs", 0755); err != nil {
		Debug("[ERROR] Failed to create logs directory, writing to debug.log instead", map[string]interface{}{
			"original_error": formattedError,
			"error": err.Error(),
		})
		return
	}

	errorLogPath := filepath.Join("/app/logs", "error.log")
	
	file, err := os.OpenFile(errorLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		Debug("[ERROR] Failed to open error.log, writing to debug.log instead", map[string]interface{}{
			"original_error": formattedError,
			"error": err.Error(),
		})
		return
	}
	defer file.Close()

	timestamp := time.Now().Format(time.RFC3339)
	logLine := fmt.Sprintf("[COCKROACH] %s\n%s\n%s\n", timestamp, formattedError, strings.Repeat("-", 80))
	
	if _, err := file.Write([]byte(logLine)); err != nil {
		Debug("[ERROR] Failed to write to error.log, writing to debug.log instead", map[string]interface{}{
			"original_error": formattedError,
			"error": err.Error(),
		})
	}
}

// Helper function to build context message
func buildContextMessage(fields map[string]interface{}) string {
	var parts []string
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(parts, ", ")
}
