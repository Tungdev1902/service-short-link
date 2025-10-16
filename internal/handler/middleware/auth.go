package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"service-short-link/internal/domain"
	"service-short-link/internal/handler"
	"service-short-link/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

const (
	AuthPlatformKey = "auth_platform"
	AuthRoleKey     = "auth_role"
	AuthIsInternalKey = "auth_is_internal"
)

type AuthMiddleware struct {
	jwtSecret string
	apiKeys   []string
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtSecret string, apiKeys []string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
		apiKeys:   apiKeys,
	}
}

// FlexibleAuth middleware supports both JWT and API Key authentication
func (a *AuthMiddleware) FlexibleAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if err := a.validateJWT(c, token); err == nil {
					c.Next()
					return
				}
			}
		}

		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.GetHeader("apikey")
		}

		if apiKey != "" {
			if a.validateAPIKey(apiKey) {
				c.Set(AuthPlatformKey, "external")
				c.Set(AuthRoleKey, "client")
				c.Set(AuthIsInternalKey, false)
				c.Next()
				return
			}
		}

		handler.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Valid JWT token or API key required", nil)
		c.Abort()
	}
}

// validateJWT validates JWT token and extracts claims
func (a *AuthMiddleware) validateJWT(c *gin.Context, tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.jwtSecret), nil
	})

	if err != nil {
		logger.ErrorWithCockroachSimple(err, "AuthMiddleware.validateToken: failed to parse token", "error_type=jwt_parse_failed")
		return fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid token claims")
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return fmt.Errorf("token has expired")
		}
	} else {
		return fmt.Errorf("token missing expiration time")
	}

	var platform, role, channelCode string
	if p, ok := claims["platform"].(string); ok {
		platform = p
	}
	if r, ok := claims["role"].(string); ok {
		role = r
	}
	if cc, ok := claims["channel_code"].(string); ok {
		channelCode = cc
	}

	c.Set(AuthPlatformKey, platform)
	c.Set(AuthRoleKey, role)
	c.Set("auth_channel_code", channelCode)
	c.Set(AuthIsInternalKey, true)

	return nil
}

// validateAPIKey validates API key
func (a *AuthMiddleware) validateAPIKey(apiKey string) bool {
	for _, validKey := range a.apiKeys {
		if apiKey == validKey {
			return true
		}
	}
	return false
}

// GetAuthContext extracts authentication context from Gin context
func GetAuthContext(c *gin.Context) (platform, role string, isInternal bool) {
	platformVal, _ := c.Get(AuthPlatformKey)
	roleVal, _ := c.Get(AuthRoleKey)
	isInternalVal, _ := c.Get(AuthIsInternalKey)
	
	platformStr, _ := platformVal.(string)
	roleStr, _ := roleVal.(string)
	isInternalBool, _ := isInternalVal.(bool)
	
	return platformStr, roleStr, isInternalBool
}
