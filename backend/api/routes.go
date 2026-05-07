package api

import (
	"github.com/gorilla/mux"
)

func NewRouter(h *Handler) *mux.Router {
	r := mux.NewRouter()

	// Apply logging middleware to all routes
	r.Use(LoggingMiddleware)

	r.HandleFunc("/chat", h.Chat).Methods("POST")
	r.HandleFunc("/upload", h.Upload).Methods("POST")

	return r
}
