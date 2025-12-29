package main

import (
	"embed"
	"log"
	"net/http"
	"prompt-service-server/config"
	"prompt-service-server/core"
	"prompt-service-server/handlers"

	"github.com/gorilla/mux"
)

//go:embed static/*
//go:embed favicon.ico
var staticFiles embed.FS

func InitializeRouter(cfg *config.Config) *mux.Router {
	promptStore := core.NewPromptStore()

	// Initialize handlers
	indexHandler := handlers.NewIndexHandler(staticFiles)
	keyHandler := handlers.NewKeyHandler(staticFiles)
	authHandler := handlers.NewAuthHandler()
	promptHandler := handlers.NewPromptHandler(promptStore)
	sseHandler := handlers.NewSSEHandler(promptStore)
	corsMiddleware := handlers.NewCORSMiddleware(cfg)

	// Create router
	r := mux.NewRouter()

	// Apply CORS middleware to all routes
	r.Use(corsMiddleware.Handler)

	// This will serve files under http://localhost:8000/static/<filename>
	r.PathPrefix("/static/").Handler(http.FileServer(http.FS(staticFiles)))
	r.Handle("/favicon.ico", http.FileServer(http.FS(staticFiles)))

	// API endpoints
	r.HandleFunc("/", indexHandler.Get).Methods("GET")
	r.HandleFunc("/key/{id}", keyHandler.Get).Methods("GET")
	r.HandleFunc("/api/auth/{id}", authHandler.Get).Methods("GET")
	r.HandleFunc("/api/prompts", promptHandler.Post).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/prompts/{id}", promptHandler.Respond).Methods("POST")
	r.HandleFunc("/api/prompts/{id}", promptHandler.Get).Methods("GET")
	r.HandleFunc("/api/sse/{id}", sseHandler.Get).Methods("GET")

	return r
}

func main() {
	// Load config
	cfg := config.LoadConfig()
	r := InitializeRouter(cfg)

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
