// Package encrypt provides AES-256-GCM encryption for sensitive values at rest (e.g. API keys).
// Stored format: "v1:" + base64(nonce || ciphertext || tag). Values without the "v1:" prefix
// are treated as legacy plaintext for backward compatibility.
package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

const (
	versionPrefix = "v1:"
	nonceSize     = 12
	gcmTagSize    = 16
	aesKeySize    = 32
)

// Encrypt encrypts plaintext with AES-256-GCM. Returns "v1:" + base64(nonce || ciphertext || tag).
// key must be 32 bytes. If plaintext is empty, returns empty string without error.
func Encrypt(plaintext string, key []byte) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	if len(key) != aesKeySize {
		return "", errors.New("encrypt: key must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aead.Seal(nil, nonce, []byte(plaintext), nil)
	// combined = nonce || ciphertext (which includes tag in GCM)
	combined := make([]byte, 0, len(nonce)+len(ciphertext))
	combined = append(combined, nonce...)
	combined = append(combined, ciphertext...)

	return versionPrefix + base64.StdEncoding.EncodeToString(combined), nil
}

// Decrypt decrypts a value produced by Encrypt. If the value does not have the "v1:" prefix,
// it is treated as legacy plaintext and returned as-is.
func Decrypt(ciphertext string, key []byte) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	if !strings.HasPrefix(ciphertext, versionPrefix) {
		// Legacy plaintext
		return ciphertext, nil
	}
	if len(key) != aesKeySize {
		return "", errors.New("encrypt: key must be 32 bytes")
	}

	encoded := ciphertext[len(versionPrefix):]
	combined, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	if len(combined) < nonceSize+gcmTagSize {
		return "", errors.New("encrypt: ciphertext too short")
	}

	nonce := combined[:nonceSize]
	sealed := combined[nonceSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plain, err := aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// DecodeKey decodes a 32-byte key from config. Accepts:
// - Raw 32-byte string (32 ASCII chars used directly as the key)
// - Hex (64 hex chars = 32 bytes)
// - Base64 (44 chars for 32 bytes)
func DecodeKey(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return nil, errors.New("encrypt: encryption key is empty")
	}
	// Raw 32-byte string (common for env vars like SECRETENCRYPTIONKEY=...)
	if len(s) == 32 {
		return []byte(s), nil
	}
	// Hex (64 hex chars = 32 bytes)
	if len(s) == 64 && isHex(s) {
		return hex.DecodeString(s)
	}
	// Base64
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(decoded) != 32 {
		return nil, errors.New("encrypt: key must decode to 32 bytes")
	}
	return decoded, nil
}

func isHex(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}
