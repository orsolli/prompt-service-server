package handlers

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
)

// AuthenticateAndVerifyCSRF checks the publicKey cookie, verifies it matches the keyHash,
// and validates the CSRF token and signature. Returns the decoded public key if valid, or writes an error/redirect and returns error.
func AuthenticateAndVerifyCSRF(w http.ResponseWriter, r *http.Request, keyHash string) ([]byte, error) {
	cookie, err := r.Cookie("publicKey")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return nil, err
	}
	cookieKey := cookie.Value
	hashedKey := sha256.Sum256([]byte(cookieKey))
	if hex.EncodeToString(hashedKey[:]) != keyHash {
		http.Redirect(w, r, "/", http.StatusFound)
		return nil, http.ErrNoCookie
	}
	signature, err := r.Cookie("CSRFChallenge")
	if err != nil {
		http.Error(w, "Missing signature", http.StatusUnauthorized)
		return nil, err
	}
	token, err := r.Cookie("CSRFToken")
	if err != nil {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return nil, err
	}
	publicKey, err := base64.StdEncoding.DecodeString(cookieKey)
	if err != nil {
		http.Error(w, "Failed to decode", http.StatusUnauthorized)
		return nil, err
	}
	signatureBytes, err := base64.StdEncoding.DecodeString(signature.Value)
	if err != nil {
		http.Error(w, "Failed to decode", http.StatusUnauthorized)
		return nil, err
	}
	valid := ed25519.Verify(publicKey, []byte(token.Value), signatureBytes)
	if !valid {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return nil, http.ErrNoCookie
	}
	return publicKey, nil
}
