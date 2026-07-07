package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

type SecretBox struct {
	aead cipher.AEAD
}

func NewSecretBox(key string) (*SecretBox, error) {
	if key == "" {
		return nil, fmt.Errorf("secret encryption key is required")
	}
	digest := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(digest[:])
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	return &SecretBox{aead: aead}, nil
}

func (s *SecretBox) Encrypt(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	sealed := s.aead.Seal(nonce, nonce, []byte(value), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (s *SecretBox) Decrypt(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	payload, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf("decode secret: %w", err)
	}
	nonceSize := s.aead.NonceSize()
	if len(payload) < nonceSize {
		return "", fmt.Errorf("invalid secret payload")
	}
	nonce, ciphertext := payload[:nonceSize], payload[nonceSize:]
	plain, err := s.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt secret: %w", err)
	}
	return string(plain), nil
}
