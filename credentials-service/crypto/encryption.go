package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

var encryptionKey []byte

// InitEncryption initializes the encryption key from environment variable
func InitEncryption() error {
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		// Generate a random key if not provided (for development)
		key = generateRandomKey()
		os.Setenv("ENCRYPTION_KEY", key)
	}

	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		// If not base64, use the string directly and pad/truncate to 32 bytes
		keyBytes := []byte(key)
		if len(keyBytes) < 32 {
			// Pad with zeros
			padded := make([]byte, 32)
			copy(padded, keyBytes)
			encryptionKey = padded
		} else {
			// Truncate to 32 bytes
			encryptionKey = keyBytes[:32]
		}
	} else {
		encryptionKey = decoded
	}

	if len(encryptionKey) != 32 {
		return errors.New("encryption key must be 32 bytes for AES-256")
	}

	return nil
}

// generateRandomKey generates a random 32-byte key for AES-256
func generateRandomKey() string {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(key)
}

// Encrypt encrypts plain text using AES-256-GCM
func Encrypt(plaintext string) (string, error) {
	if encryptionKey == nil {
		if err := InitEncryption(); err != nil {
			return "", err
		}
	}

	block, err := aes.NewCipher(encryptionKey)
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

// Decrypt decrypts cipher text using AES-256-GCM
func Decrypt(ciphertext string) (string, error) {
	if encryptionKey == nil {
		if err := InitEncryption(); err != nil {
			return "", err
		}
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
