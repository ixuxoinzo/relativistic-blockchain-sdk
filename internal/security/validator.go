package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

type SecurityValidator struct {
	logger        *zap.Logger
	passwordRegex *regexp.Regexp
	rateLimiter   *RateLimiter
	blacklist     map[string]time.Time
}

func NewSecurityValidator(logger *zap.Logger) *SecurityValidator {
	return &SecurityValidator{
		logger: logger,
		passwordRegex: regexp.MustCompile(`^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$`),
		rateLimiter:   NewRateLimiter(100, time.Minute),
		blacklist:     make(map[string]time.Time),
	}
}

func (sv *SecurityValidator) ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if !sv.passwordRegex.MatchString(password) {
		return fmt.Errorf("password must contain uppercase, lowercase, number and special character")
	}

	return nil
}

func (sv *SecurityValidator) ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

func (sv *SecurityValidator) ValidateNodeID(nodeID string) error {
	if len(nodeID) < 3 || len(nodeID) > 64 {
		return fmt.Errorf("node ID must be between 3 and 64 characters")
	}

	validRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validRegex.MatchString(nodeID) {
		return fmt.Errorf("node ID can only contain letters, numbers, hyphens and underscores")
	}

	return nil
}

func (sv *SecurityValidator) ValidateIPAddress(ip string) error {
	ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	if !ipRegex.MatchString(ip) {
		return fmt.Errorf("invalid IP address format")
	}

	parts := strings.Split(ip, ".")
	for _, part := range parts {
		if num := parseInt(part); num < 0 || num > 255 {
			return fmt.Errorf("invalid IP address range")
		}
	}

	return nil
}

func (sv *SecurityValidator) HashData(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (sv *SecurityValidator) CheckRateLimit(identifier string) bool {
	return sv.rateLimiter.Allow(identifier)
}

func (sv *SecurityValidator) AddToBlacklist(identifier string, duration time.Duration) {
	sv.blacklist[identifier] = time.Now().Add(duration)
}

func (sv *SecurityValidator) IsBlacklisted(identifier string) bool {
	if expiry, exists := sv.blacklist[identifier]; exists {
		if time.Now().Before(expiry) {
			return true
		}
		delete(sv.blacklist, identifier)
	}
	return false
}

func (sv *SecurityValidator) CleanupBlacklist() {
	now := time.Now()
	for identifier, expiry := range sv.blacklist {
		if now.After(expiry) {
			delete(sv.blacklist, identifier)
		}
	}
}

func (sv *SecurityValidator) ValidateTimestamp(timestamp time.Time, maxAge time.Duration) error {
	now := time.Now()
	if timestamp.After(now) {
		return fmt.Errorf("timestamp cannot be in the future")
	}

	if now.Sub(timestamp) > maxAge {
		return fmt.Errorf("timestamp is too old")
	}

	return nil
}

func (sv *SecurityValidator) ValidatePosition(latitude, longitude, altitude float64) error {
	if latitude < -90 || latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}

	if longitude < -180 || longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	if altitude < 0 {
		return fmt.Errorf("altitude cannot be negative")
	}

	return nil
}

func parseInt(s string) int {
	var result int
	for _, ch := range s {
		result = result*10 + int(ch-'0')
	}
	return result
}