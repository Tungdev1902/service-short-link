package handler

// API response error codes (HTTP/transport-specific)
const (
	CodeBadRequest         = "BAD_REQUEST"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT" // e.g., short code exists
	CodeGone               = "GONE"     // resource no longer available
	CodeRequestTimeout     = "REQUEST_TIMEOUT"
	CodeTooManyRequests    = "TOO_MANY_REQUESTS"
	CodeInternalError      = "INTERNAL_ERROR"
	CodeBadGateway         = "BAD_GATEWAY"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeGatewayTimeout     = "GATEWAY_TIMEOUT"
	CodeMethodNotAllowed   = "METHOD_NOT_ALLOWED"
)
