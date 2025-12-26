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

	// Serve key.html with embedded JavaScript
	http.ServeFile(w, r, "static/key.html")
}
