package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/security"
)

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := s.securityValidator.ValidateToken(token)
		if err != nil {
			s.logger.Warn("Invalid token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}

func (s *Server) rateLimitMiddleware() gin.HandlerFunc {
	limiter := security.NewRateLimiter(100, time.Minute)
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		if !limiter.Allow(clientIP) {
			s.logger.Warn("Rate limit exceeded", zap.String("ip", clientIP))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": limiter.RetryAfter(clientIP).Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (s *Server) adminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		roleList, ok := roles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		hasAdmin := false
		for _, role := range roleList {
			if role == "admin" {
				hasAdmin = true
				break
			}
		}

		if !hasAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (s *Server) validateNodeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		nodeID := c.Param("id")
		if nodeID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Node ID required"})
			c.Abort()
			return
		}

		_, err := s.topologyManager.GetNode(nodeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (s *Server) timeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		if ctx.Err() == context.DeadlineExceeded {
			s.logger.Warn("Request timeout", 
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
			c.Abort()
		}
	}
}

func (s *Server) requestSizeMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		
		c.Next()
	}
}

func (s *Server) securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		
		c.Next()
	}
}

func (s *Server) cacheControlMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			c.Header("Cache-Control", "public, max-age=300")
		} else {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		}
		
		c.Next()
	}
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (s *Server) validationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			var validationErrors []ValidationError
			
			for _, err := range c.Errors {
				if fieldErr, ok := err.Err.(validator.ValidationError); ok {
					validationErrors = append(validationErrors, ValidationError{
						Field:   fieldErr.Field(),
						Message: fieldErr.Tag(),
					})
				} else {
					validationErrors = append(validationErrors, ValidationError{
						Field:   "general",
						Message: err.Error(),
					})
				}
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Validation failed",
				"details": validationErrors,
			})
			c.Abort()
		}
	}
}

func (s *Server) recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}