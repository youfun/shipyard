package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

// key must be 32 bytes for AES-256
var key []byte

// Init initializes the encryption key from a hex-encoded string.
func Init(hexKey string) error {
	k, err := hex.DecodeString(hexKey)
	if err != nil {
		return fmt.Errorf("failed to decode encryption key: %w", err)
	}
	if len(k) != 32 {
		return fmt.Errorf("encryption key must be 32 bytes (64 hex characters), got %d bytes", len(k))
	}
	key = k
	return nil
}

func Encrypt(stringToEncrypt string) (string, error) {
	if len(key) == 0 {
		return "", fmt.Errorf("encryption key not initialized")
	}

	plaintext := []byte(stringToEncrypt)

	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	return hex.EncodeToString(gcm.Seal(nonce, nonce, plaintext, nil)), nil
}

func Decrypt(encryptedString string) (string, error) {
	if len(key) == 0 {
		return "", fmt.Errorf("encryption key not initialized")
	}

	enc, err := hex.DecodeString(encryptedString)
	if err != nil {
		return "", err
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(enc) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GeneratePhoenixSecret generates a 64-character long secret suitable for Phoenix's SECRET_KEY_BASE.
func GeneratePhoenixSecret() (string, error) {
	randomBytes := make([]byte, 48) // 48 bytes = 384 bits, which results in a 64-char Base64 string
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate SECRET_KEY_BASE: %w", err)
	}
	return base64.StdEncoding.EncodeToString(randomBytes), nil
}
