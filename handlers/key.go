package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

type KeyHandler struct{}

func NewKeyHandler() *KeyHandler {
	return &KeyHandler{}
}

func (h *KeyHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyHash := vars["id"]
	VerifyKeyHash(w, r, keyHash) // Ensure the publicKey cookie matches the keyHash

	// Set security headers
	w.Header().Set("Content-Security-Policy", buildCSP(getCSPConfig()))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")

	// Serve key.html with embedded JavaScript
	http.ServeFile(w, r, "static/key.html")
}
