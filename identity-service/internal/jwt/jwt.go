package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token expired")
	ErrInvalidSecretKey = errors.New("jwt: secret key must be at least 32 bytes")
)

const AccessTokenSize = 32

type manager struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewManager(secret string, issuer string, accessTTL, refreshTTL time.Duration) (Manager, error) {
	if len(secret) < AccessTokenSize {
		return nil, ErrInvalidSecretKey
	}
	return &manager{
		secret:     []byte(secret),
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}, nil
}

func (m *manager) AccessTTL() time.Duration {
	return m.accessTTL
}

func (m *manager) RefreshTTL() time.Duration {
	return m.refreshTTL
}

func (m *manager) GenerateAccessToken(userID uuid.UUID, appID uuid.UUID, roles []string) (string, error) {
	now := time.Now()

	claims := AccessClaims{
		UserID: userID,
		AppID:  appID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(m.secret)
}

func (m *manager) ParseAccessToken(accessToken string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Enforce HMAC signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
