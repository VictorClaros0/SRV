package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims del JWT.
type Claims struct {
	UserID  uint   `json:"userId"`
	EsAdmin bool   `json:"esAdmin"`
	jwt.RegisteredClaims
}

// GenerateToken emite un JWT con id de usuario y rol.
func GenerateToken(userID uint, esAdmin bool, secret string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID:  userID,
		EsAdmin: esAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// ParseToken valida y parsea el JWT.
func ParseToken(tokenStr, secret string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de firma inesperado")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if c, ok := t.Claims.(*Claims); ok && t.Valid {
		return c, nil
	}
	return nil, errors.New("token inválido")
}
