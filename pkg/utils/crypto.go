package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
   "crypto/hmac"
	"fmt"
)

type CryptoUtils struct{}

func NewCryptoUtils() *CryptoUtils {
	return &CryptoUtils{}
}

func (cu *CryptoUtils) GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

func (cu *CryptoUtils) GenerateRandomString(length int) (string, error) {
	bytes, err := cu.GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (cu *CryptoUtils) HashString(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (cu *CryptoUtils) HashBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (cu *CryptoUtils) GenerateUUID() (string, error) {
	bytes, err := cu.GenerateRandomBytes(16)
	if err != nil {
		return "", err
	}
	
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		bytes[0:4],
		bytes[4:6],
		bytes[6:8],
		bytes[8:10],
		bytes[10:16]), nil
}

func (cu *CryptoUtils) GenerateShortID() (string, error) {
	bytes, err := cu.GenerateRandomBytes(8)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (cu *CryptoUtils) Base64Encode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

func (cu *CryptoUtils) Base64Decode(encoded string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(encoded)
}

func (cu *CryptoUtils) HexEncode(data []byte) string {
	return hex.EncodeToString(data)
}

func (cu *CryptoUtils) HexDecode(encoded string) ([]byte, error) {
	return hex.DecodeString(encoded)
}

func (cu *CryptoUtils) GenerateSalt() (string, error) {
	salt, err := cu.GenerateRandomBytes(16)
	if err != nil {
		return "", err
	}
	return cu.Base64Encode(salt), nil
}

func (cu *CryptoUtils) HashPassword(password, salt string) string {
	saltedPassword := password + salt
	return cu.HashString(saltedPassword)
}

func (cu *CryptoUtils) VerifyPassword(password, salt, expectedHash string) bool {
	actualHash := cu.HashPassword(password, salt)
	return actualHash == expectedHash
}

func (cu *CryptoUtils) GenerateAPIKey() (string, error) {
	key, err := cu.GenerateRandomBytes(32)
	if err != nil {
		return "", err
	}
	return "rk_" + cu.Base64Encode(key), nil
}

func (cu *CryptoUtils) ValidateAPIKey(apiKey string) bool {
	if len(apiKey) != 45 || apiKey[:3] != "rk_" {
		return false
	}
	
	decoded, err := cu.Base64Decode(apiKey[3:])
	if err != nil {
		return false
	}
	
	return len(decoded) == 32
}

func (cu *CryptoUtils) GenerateNonce() (string, error) {
	return cu.GenerateRandomString(16)
}

func (cu *CryptoUtils) CalculateHMAC(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (cu *CryptoUtils) VerifyHMAC(data, key, expectedMAC string) bool {
	actualMAC := cu.CalculateHMAC(data, key)
	return hmac.Equal([]byte(actualMAC), []byte(expectedMAC))
}