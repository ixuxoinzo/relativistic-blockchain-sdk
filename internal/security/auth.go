package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

type AuthManager struct {
	jwtSecret   []byte
	tokenExpiry time.Duration
	logger      *zap.Logger
}

type Claims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func NewAuthManager(jwtSecret string, tokenExpiry time.Duration, logger *zap.Logger) (*AuthManager, error) {
	if len(jwtSecret) < 32 {
		return nil, errors.New("JWT secret must be at least 32 characters")
	}

	return &AuthManager{
		jwtSecret:   []byte(jwtSecret),
		tokenExpiry: tokenExpiry,
		logger:      logger,
	}, nil
}

func (am *AuthManager) GenerateToken(userID string, roles []string) (string, error) {
	expirationTime := time.Now().Add(am.tokenExpiry)

	claims := &Claims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "relativistic-sdk",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(am.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

func (am *AuthManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return am.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (am *AuthManager) RefreshToken(tokenString string) (string, error) {
	claims, err := am.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid token: %w", err)
	}

	if time.Until(claims.ExpiresAt.Time) > 30*time.Minute {
		return "", errors.New("token is not expired yet")
	}

	newToken, err := am.GenerateToken(claims.UserID, claims.Roles)
	if err != nil {
		return "", fmt.Errorf("failed to generate new token: %w", err)
	}

	return newToken, nil
}

func (am *AuthManager) GenerateAPIKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}

	apiKey := base64.URLEncoding.EncodeToString(key)
	return "rk_" + apiKey, nil
}

func (am *AuthManager) ValidateAPIKey(apiKey string) bool {
	if len(apiKey) != 45 || !strings.HasPrefix(apiKey, "rk_") {
		return false
	}

	decoded, err := base64.URLEncoding.DecodeString(apiKey[3:])
	if err != nil {
		return false
	}

	return len(decoded) == 32
}

func (am *AuthManager) HasRole(claims *Claims, requiredRole string) bool {
	for _, role := range claims.Roles {
		if role == requiredRole {
			return true
		}
	}
	return false
}

func (am *AuthManager) HasAnyRole(claims *Claims, requiredRoles []string) bool {
	for _, userRole := range claims.Roles {
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

func (am *AuthManager) ExtractUserIDFromToken(tokenString string) (string, error) {
	claims, err := am.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

func (am *AuthManager) GetTokenExpiry(tokenString string) (time.Time, error) {
	claims, err := am.ValidateToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}
	return claims.ExpiresAt.Time, nil
}
