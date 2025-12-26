package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key-here") // In production, use environment variable

type Claims struct {
	KeyHash string `json:"key_hash"`
	jwt.RegisteredClaims
}

func GenerateCSRFToken(keyHash string) (string, error) {
	// Generate a JWT with a secret key
	now := time.Now()
	claims := &Claims{
		KeyHash: keyHash, // This would be the user's public key hash
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func VerifyJWT(tokenString string) error {
	// Verify the JWT signature
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	return nil
}
