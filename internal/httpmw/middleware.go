package httpmw

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"strconv"

	"github.com/anton1ks96/college-core-api/internal/services"
	"github.com/anton1ks96/college-core-api/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

func AuthMiddleware(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		user, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("role", user.Role)
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		logger.Debug(requestID + " " + c.Request.Method + " " + c.Request.URL.Path)
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		logger.Debug(requestID + " completed in " + latency.String() + " with status " + strconv.Itoa(c.Writer.Status()))
	}
}

func RateLimitMiddleware(rps int) gin.HandlerFunc {
	limiters := make(map[string]*rate.Limiter)
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}
		key := userID.(string)
		limiter, ok := limiters[key]
		if !ok {
			limiter = rate.NewLimiter(rate.Limit(rps), rps*2)
			limiters[key] = limiter
		}
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(err.(error))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				c.Abort()
			}
		}()
		c.Next()
	}
}

func extractToken(c *gin.Context) (string, error) {
	token := c.GetHeader("Authorization")
	if token != "" {
		if !strings.HasPrefix(token, "Bearer ") {
			return "", fmt.Errorf("invalid authorization header format")
		}
		extractedToken := strings.TrimPrefix(token, "Bearer ")
		if extractedToken == "" {
			return "", fmt.Errorf("empty token in authorization header")
		}
		return extractedToken, nil
	}

	cookieToken, err := c.Cookie("access_token")
	if err == nil && cookieToken != "" {
		return cookieToken, nil
	}

	return "", fmt.Errorf("authorization header is missing")
}
