package compliance

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// EncryptPII encrypts sensitive data using AES-GCM
func EncryptPII(plaintext string) (string, error) {
	key := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(key) != 32 {
		// Fallback for dev/test if key is not set or invalid length
		// In production, this should panic or return error
		if os.Getenv("APP_ENV") != "production" {
			key = []byte("12345678901234567890123456789012") // 32 bytes
		} else {
			return "", errors.New("invalid encryption key length")
		}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPII decrypts sensitive data
func DecryptPII(ciphertext string) (string, error) {
	key := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(key) != 32 {
		if os.Getenv("APP_ENV") != "production" {
			key = []byte("12345678901234567890123456789012")
		} else {
			return "", errors.New("invalid encryption key length")
		}
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", errors.New("malformed ciphertext")
	}

	nonce, ciphertextBytes := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
