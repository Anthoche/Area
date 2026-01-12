package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

const encryptedPrefix = "enc:"

// Enabled reports whether encryption is configured.
func Enabled() bool {
	_, err := loadKey()
	return err == nil
}

// EncryptString encrypts a string with AES-GCM and returns a tagged ciphertext.
func EncryptString(plain string) (string, error) {
	key, err := loadKey()
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
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(plain), nil)
	raw := append(nonce, ciphertext...)
	return encryptedPrefix + base64.StdEncoding.EncodeToString(raw), nil
}

// DecryptString decrypts a tagged ciphertext and returns the plaintext.
func DecryptString(cipherText string) (string, error) {
	if !strings.HasPrefix(cipherText, encryptedPrefix) {
		return cipherText, nil
	}
	key, err := loadKey()
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(cipherText, encryptedPrefix))
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
	if len(raw) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce := raw[:gcm.NonceSize()]
	payload := raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// loadKey loads and decodes the encryption key from the environment variable.
func loadKey() ([]byte, error) {
	val := strings.TrimSpace(os.Getenv("APP_SECRET_KEY"))
	if val == "" {
		return nil, fmt.Errorf("missing APP_SECRET_KEY")
	}
	if strings.HasPrefix(val, "base64:") {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(val, "base64:"))
		if err != nil {
			return nil, err
		}
		if len(decoded) != 32 {
			return nil, fmt.Errorf("APP_SECRET_KEY must be 32 bytes")
		}
		return decoded, nil
	}
	if len(val) == 32 {
		return []byte(val), nil
	}
	decoded, err := base64.StdEncoding.DecodeString(val)
	if err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	return nil, fmt.Errorf("APP_SECRET_KEY must be 32 bytes (raw or base64)")
}
