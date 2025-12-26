package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCSRFToken(t *testing.T) {
	keyHash := "test-key-hash"

	token, err := GenerateCSRFToken(keyHash)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify the token can be parsed and contains expected claims
	err = VerifyJWT(token)
	assert.NoError(t, err)
}

func TestVerifyJWT_ValidToken(t *testing.T) {
	keyHash := "test-key-hash"

	token, err := GenerateCSRFToken(keyHash)
	require.NoError(t, err)

	err = VerifyJWT(token)
	assert.NoError(t, err)
}

func TestVerifyJWT_InvalidToken(t *testing.T) {
	invalidToken := "invalid.jwt.token"

	err := VerifyJWT(invalidToken)
	assert.Error(t, err)
}

func TestVerifyJWT_ExpiredToken(t *testing.T) {
	// Create a token that expires immediately
	now := time.Now().Add(-10 * time.Minute) // 10 minutes ago
	claims := &Claims{
		KeyHash: "test-key-hash",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(-5 * time.Minute)), // Already expired
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	require.NoError(t, err)

	err = VerifyJWT(tokenString)
	assert.Error(t, err) // Should fail because token is expired
}
