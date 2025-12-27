package handlers

import (
	"embed"
	"net/http"

	"github.com/gorilla/mux"
)

type KeyHandler struct {
	staticFiles embed.FS
}

func NewKeyHandler(staticFiles embed.FS) *KeyHandler {
	return &KeyHandler{
		staticFiles: staticFiles,
	}
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
	p, err := h.staticFiles.ReadFile("static/key.html")
	if err != nil {
		http.Error(w, "Failed to load key.html", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(p); err != nil {
		// Optionally log the error or handle it as needed
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
