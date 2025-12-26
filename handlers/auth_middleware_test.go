package handlers

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"prompt-service-server/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyKeyHash_ValidCookie(t *testing.T) {
	// Generate a test public key
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	pubKeyB64 := base64.StdEncoding.EncodeToString(pub)
	hashedKey := sha256.Sum256([]byte(pubKeyB64))
	keyHash := hex.EncodeToString(hashedKey[:])

	// Create request with valid cookie
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "publicKey", Value: pubKeyB64})

	w := httptest.NewRecorder()

	key, err := VerifyKeyHash(w, req, keyHash)
	assert.NoError(t, err)
	assert.Equal(t, pubKeyB64, key)
}

func TestVerifyKeyHash_InvalidHash(t *testing.T) {
	// Generate a test public key
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	pubKeyB64 := base64.StdEncoding.EncodeToString(pub)

	// Create request with cookie but wrong hash
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "publicKey", Value: pubKeyB64})

	w := httptest.NewRecorder()

	key, err := VerifyKeyHash(w, req, "wrong-hash")
	assert.Error(t, err)
	assert.Equal(t, pubKeyB64, key) // Function returns the key even on hash mismatch

	// Verify redirect was issued
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))
}

func TestVerifyKeyHash_NoCookie(t *testing.T) {
	// Create request without cookie
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	key, err := VerifyKeyHash(w, req, "some-hash")
	assert.Error(t, err)
	assert.Empty(t, key)

	// Verify redirect was issued
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))
}

func TestAuthenticateAndVerifyCSRF_Valid(t *testing.T) {
	// This test checks behavior when cookies are missing
	// It should redirect due to missing publicKey cookie

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	key, err := AuthenticateAndVerifyCSRF(w, req, "some-hash")
	assert.Error(t, err) // Should fail due to missing cookies
	assert.Empty(t, key)
	assert.Equal(t, http.StatusFound, w.Code) // Redirect due to missing publicKey
}

// Helper function to create a signed JWT for testing
func createTestJWTAndSignature(keyHash string, pubKey ed25519.PublicKey, privKey ed25519.PrivateKey) (string, string, error) {
	// Generate JWT token
	token, err := utils.GenerateCSRFToken(keyHash)
	if err != nil {
		return "", "", err
	}

	// Sign the token
	signature := ed25519.Sign(privKey, []byte(token))
	sigB64 := base64.StdEncoding.EncodeToString(signature)

	return token, sigB64, nil
}

func TestAuthenticateAndVerifyCSRF_FullFlow(t *testing.T) {
	// Generate keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	pubKeyB64 := base64.StdEncoding.EncodeToString(pub)
	hashedKey := sha256.Sum256([]byte(pubKeyB64))
	keyHash := hex.EncodeToString(hashedKey[:])

	// Create JWT and signature
	token, signature, err := createTestJWTAndSignature(keyHash, pub, priv)
	require.NoError(t, err)

	// Create request with all required cookies
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "publicKey", Value: pubKeyB64})
	req.AddCookie(&http.Cookie{Name: "CSRFToken", Value: token})
	req.AddCookie(&http.Cookie{Name: "CSRFChallenge", Value: signature})

	w := httptest.NewRecorder()

	key, err := AuthenticateAndVerifyCSRF(w, req, keyHash)
	assert.NoError(t, err)
	assert.Equal(t, pubKeyB64, key)
}

func TestAuthenticateAndVerifyCSRF_InvalidSignature(t *testing.T) {
	// Generate keypair
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	pubKeyB64 := base64.StdEncoding.EncodeToString(pub)
	hashedKey := sha256.Sum256([]byte(pubKeyB64))
	keyHash := hex.EncodeToString(hashedKey[:])

	// Create request with invalid signature
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "publicKey", Value: pubKeyB64})
	req.AddCookie(&http.Cookie{Name: "CSRFToken", Value: "valid.jwt.token"})
	req.AddCookie(&http.Cookie{Name: "CSRFChallenge", Value: "invalid-signature"})

	w := httptest.NewRecorder()

	key, err := AuthenticateAndVerifyCSRF(w, req, keyHash)
	assert.Error(t, err)
	assert.Equal(t, pubKeyB64, key) // Function returns the key even on signature failure
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
