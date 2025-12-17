package handlers

import (
    "crypto/sha256"
    "encoding/hex"
    "net/http"
	"github.com/gorilla/mux"
    "time"
)

type SSEHandler struct{}

func NewSSEHandler() *SSEHandler {
    return &SSEHandler{}
}

func (h *SSEHandler) Get(w http.ResponseWriter, r *http.Request) {
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
	
    // Verify signature (similar to prompt handlers)
    signature := r.Header.Get("Authorization")
    if signature == "" {
        http.Error(w, "Missing signature", http.StatusUnauthorized)
        return
    }
    
    // Validate signature against public key
    
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    // In a real implementation, we would:
    // 1. Verify the signature
    // 2. Establish the SSE connection
    // 3. Send events to the client
    
    // For demonstration, we'll send a simple event
    go func() {
        // Simulate sending events
        for {
            select {
            case <-time.After(5 * time.Second):
                // Send a test event
                event := `data: {"type": "heartbeat", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}\n\n`
                w.Write([]byte(event))
                w.(http.Flusher).Flush()
            }
        }
    }()
    
    // Keep connection alive
    <-r.Context().Done()
}