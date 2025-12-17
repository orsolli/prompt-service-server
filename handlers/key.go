package handlers

import (
    "crypto/sha256"
    "encoding/hex"
    "net/http"
    "prompt-service-server/utils"
	"github.com/gorilla/mux"
)

type KeyHandler struct{}

func NewKeyHandler() *KeyHandler {
    return &KeyHandler{}
}

func (h *KeyHandler) Get(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    keyHash := vars["id"]
    
    // Check if cookie contains matching public key
    cookie, err := r.Cookie("publicKey")
    if err != nil {
        // Redirect to root
        http.Redirect(w, r, "/", http.StatusFound)
        return
    }
    
    // Verify the cookie's public key matches the hash
    cookieKey := cookie.Value
    hashedKey := sha256.Sum256([]byte(cookieKey))
    if hex.EncodeToString(hashedKey[:]) != keyHash {
        // Redirect to root
        http.Redirect(w, r, "/", http.StatusFound)
        return
    }
    
    // If we get here, we have a valid key
    // Generate CSRF token
    csrfToken, err := utils.GenerateCSRFToken()
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    
    // Set CSRF header
    w.Header().Set("X-CSRF", csrfToken)
    
    // Serve key.html with embedded JavaScript
    http.ServeFile(w, r, "static/key.html")
}