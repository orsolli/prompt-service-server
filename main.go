package main

import (
    "log"
    "net/http"
	"github.com/gorilla/mux"
    "prompt-service-server/handlers"
    "prompt-service-server/config"
)

func main() {
    // Load config
    cfg := config.LoadConfig()
    
    // Initialize handlers
    indexHandler := handlers.NewIndexHandler()
    keyHandler := handlers.NewKeyHandler()
    promptHandler := handlers.NewPromptHandler()
    sseHandler := handlers.NewSSEHandler()
    
    // Create router
    r := mux.NewRouter()
    
     // This will serve files under http://localhost:8000/static/<filename>
    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

    // API endpoints
    r.HandleFunc("/", indexHandler.Get).Methods("GET")
    r.HandleFunc("/key/{id}", keyHandler.Get).Methods("GET")
    r.HandleFunc("/api/prompts", promptHandler.Post).Methods("POST")
    r.HandleFunc("/api/prompts", promptHandler.Get).Methods("GET")
    r.HandleFunc("/api/sse/{id}", sseHandler.Get).Methods("GET")
    
    // Start server
    port := cfg.Port
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}