package utils

import (
	"fmt"
	"net"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type ValidationUtils struct {
	emailRegex  *regexp.Regexp
	nodeIDRegex *regexp.Regexp
	ipRegex     *regexp.Regexp
	urlRegex    *regexp.Regexp
}

func NewValidationUtils() *ValidationUtils {
	return &ValidationUtils{
		emailRegex:  regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		nodeIDRegex: regexp.MustCompile(`^[a-zA-Z0-9_-]{3,64}$`),
		ipRegex:     regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`),
		urlRegex:    regexp.MustCompile(`^(https?://)?([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}(:\d+)?(/.*)?$`),
	}
}

func (vu *ValidationUtils) ValidateEmail(email string) error {
	if email == "" {
		return types.NewError(types.ErrInvalidInput, "email is required")
	}

	if !vu.emailRegex.MatchString(email) {
		return types.NewError(types.ErrInvalidInput, "invalid email format")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return types.NewError(types.ErrInvalidInput, "invalid email address")
	}

	return nil
}

func (vu *ValidationUtils) ValidateNodeID(nodeID string) error {
	if nodeID == "" {
		return types.NewError(types.ErrInvalidInput, "node ID is required")
	}

	if len(nodeID) < 3 || len(nodeID) > 64 {
		return types.NewError(types.ErrInvalidInput, "node ID must be between 3 and 64 characters")
	}

	if !vu.nodeIDRegex.MatchString(nodeID) {
		return types.NewError(types.ErrInvalidInput, "node ID can only contain letters, numbers, hyphens and underscores")
	}

	return nil
}

func (vu *ValidationUtils) ValidateIPAddress(ip string) error {
	if ip == "" {
		return types.NewError(types.ErrInvalidInput, "IP address is required")
	}

	if !vu.ipRegex.MatchString(ip) {
		return types.NewError(types.ErrInvalidInput, "invalid IP address format")
	}

	parts := strings.Split(ip, ".")
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 || num > 255 {
			return types.NewError(types.ErrInvalidInput, "invalid IP address range")
		}
	}

	if net.ParseIP(ip) == nil {
		return types.NewError(types.ErrInvalidInput, "invalid IP address")
	}

	return nil
}

func (vu *ValidationUtils) ValidateURL(url string) error {
	if url == "" {
		return types.NewError(types.ErrInvalidInput, "URL is required")
	}

	if !vu.urlRegex.MatchString(url) {
		return types.NewError(types.ErrInvalidInput, "invalid URL format")
	}

	return nil
}

func (vu *ValidationUtils) ValidatePosition(pos types.Position) error {
	if pos.Latitude < -90 || pos.Latitude > 90 {
		return types.NewError(types.ErrInvalidInput, "latitude must be between -90 and 90")
	}

	if pos.Longitude < -180 || pos.Longitude > 180 {
		return types.NewError(types.ErrInvalidInput, "longitude must be between -180 and 180")
	}

	if pos.Altitude < 0 {
		return types.NewError(types.ErrInvalidInput, "altitude cannot be negative")
	}

	return nil
}

func (vu *ValidationUtils) ValidatePassword(password string) error {
	if len(password) < 8 {
		return types.NewError(types.ErrInvalidInput, "password must be at least 8 characters long")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return types.NewError(types.ErrInvalidInput, "password must contain at least one uppercase letter")
	}
	if !hasLower {
		return types.NewError(types.ErrInvalidInput, "password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return types.NewError(types.ErrInvalidInput, "password must contain at least one number")
	}
	if !hasSpecial {
		return types.NewError(types.ErrInvalidInput, "password must contain at least one special character")
	}

	return nil
}

func (vu *ValidationUtils) ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return types.NewError(types.ErrInvalidInput, "port must be between 1 and 65535")
	}
	return nil
}

func (vu *ValidationUtils) ValidateNotEmpty(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return types.NewError(types.ErrInvalidInput, fieldName+" is required")
	}
	return nil
}

func (vu *ValidationUtils) ValidateLength(value string, min, max int, fieldName string) error {
	length := len(strings.TrimSpace(value))
	if length < min {
		return types.NewError(types.ErrInvalidInput, fmt.Sprintf("%s must be at least %d characters", fieldName, min))
	}
	if length > max {
		return types.NewError(types.ErrInvalidInput, fmt.Sprintf("%s must be at most %d characters", fieldName, max))
	}
	return nil
}

func (vu *ValidationUtils) ValidateRange(value, min, max float64, fieldName string) error {
	if value < min || value > max {
		return types.NewError(types.ErrInvalidInput, fmt.Sprintf("%s must be between %f and %f", fieldName, min, max))
	}
	return nil
}
