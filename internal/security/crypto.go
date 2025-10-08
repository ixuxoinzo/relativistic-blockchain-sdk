package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type CryptoManager struct {
	encryptionKey []byte
}

func NewCryptoManager(key string) (*CryptoManager, error) {
	if len(key) != 32 {
		return nil, errors.New("encryption key must be 32 bytes")
	}

	hash := sha256.Sum256([]byte(key))
	return &CryptoManager{
		encryptionKey: hash[:],
	}, nil
}

func (cm *CryptoManager) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(cm.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (cm *CryptoManager) Decrypt(encoded string) ([]byte, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(cm.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

func (cm *CryptoManager) HashData(data []byte) string {
	hash := sha256.Sum256(data)
	return base64.URLEncoding.EncodeToString(hash[:])
}

func (cm *CryptoManager) GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

func (cm *CryptoManager) GenerateSalt() (string, error) {
	salt, err := cm.GenerateRandomBytes(16)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(salt), nil
}

func (cm *CryptoManager) HashPassword(password, salt string) string {
	saltedPassword := password + salt
	hash := sha256.Sum256([]byte(saltedPassword))
	return base64.URLEncoding.EncodeToString(hash[:])
}

func (cm *CryptoManager) VerifyPassword(password, salt, expectedHash string) bool {
	actualHash := cm.HashPassword(password, salt)
	return actualHash == expectedHash
}

func (cm *CryptoManager) GenerateKeyPair() (privateKey, publicKey string, err error) {
	privateBytes, err := cm.GenerateRandomBytes(32)
	if err != nil {
		return "", "", err
	}

	publicBytes, err := cm.GenerateRandomBytes(32)
	if err != nil {
		return "", "", err
	}

	privateKey = base64.URLEncoding.EncodeToString(privateBytes)
	publicKey = base64.URLEncoding.EncodeToString(publicBytes)

	return privateKey, publicKey, nil
}
