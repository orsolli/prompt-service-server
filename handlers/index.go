package handlers

import (
    "net/http"
)

type IndexHandler struct{}

func NewIndexHandler() *IndexHandler {
    return &IndexHandler{}
}

func (h *IndexHandler) Get(w http.ResponseWriter, r *http.Request) {
    // Check if user has keys in localStorage
    // This is a client-side check, but we can set up the redirect logic here
    // For now, we'll just serve the index.html
    http.ServeFile(w, r, "static/index.html")
}