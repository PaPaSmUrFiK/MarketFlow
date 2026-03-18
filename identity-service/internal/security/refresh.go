package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

const refreshTokenSize = 32

var ErrInvalidTokenLength = errors.New("invalid refresh token length")

func GenerateRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, refreshTokenSize)

	if _, err = rand.Read(b); err != nil {
		return "", "", nil
	}

	raw = base64.RawURLEncoding.EncodeToString(b)

	hashBytes := sha256.Sum256([]byte(raw))
	hash = base64.RawURLEncoding.EncodeToString(hashBytes[:])

	return raw, hash, nil
}

func HashRefreshToken(raw string) (string, error) {

	if raw == "" {
		return "", ErrInvalidTokenLength
	}

	hashBytes := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(hashBytes[:]), nil
}
