package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"prompt-service-server/utils"
	"time"

	"github.com/gorilla/mux"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyHash := vars["id"]

	// Check if cookie contains matching public key
	cookie, err := r.Cookie("publicKey")
	if err != nil {
		http.Error(w, "Invalid publicKey cookie", http.StatusBadRequest)
		return
	}

	// Verify the cookie's public key matches the hash
	cookieKey := cookie.Value
	hashedKey := sha256.Sum256([]byte(cookieKey))
	if hex.EncodeToString(hashedKey[:]) != keyHash {
		http.Error(w, "Invalid publicKey cookie", http.StatusBadRequest)
		return
	}

	// If we get here, we have a valid key
	// Generate CSRF token
	csrfToken, err := utils.GenerateCSRFToken(keyHash)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	expiration := time.Now().Add(5 * time.Minute)
	csrfCookie := http.Cookie{Name: "CSRFToken", Value: csrfToken, Expires: expiration, Path: "/api"}
	http.SetCookie(w, &csrfCookie)

	// Return the challenge
	w.Header().Set("Content-Type", "plain/text")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(csrfToken))
}
