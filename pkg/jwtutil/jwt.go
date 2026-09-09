package jwtutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userId int64, role, secretKey string, duration time.Duration) (string, error) {
	claims := CustomClaims{
		UserID: userId,
		Role:   role,
		// PERBAIKAN 1: Huruf R kapital pada RegisteredClaims
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			// PERBAIKAN 2: Typo IssuedAt (bukan IsuuedAt)
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func ValidateToken(tokenString, secretKey string) (*CustomClaims, error) {
	// PERBAIKAN 3: Ubah parameter anonymous function menjadi 'token *jwt.Token' (hapus koma)
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			// PERBAIKAN 4: Ganti nill menjadi nil
			return nil, errors.New("signature is invalid!")
		}
		return []byte(secretKey), nil
	})

	// PERBAIKAN 4: Ganti nill menjadi nil
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// PERBAIKAN 4: Ganti nill menjadi nil
		return claims, nil
	}

	// PERBAIKAN 5: Huruf N kapital pada errors.New dan ganti nill menjadi nil
	return nil, errors.New("token is invalid!")
}
