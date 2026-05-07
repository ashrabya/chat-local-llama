package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/ashrabya/chat-local-llama.git/backend/api"
	"github.com/ashrabya/chat-local-llama.git/backend/rag"
)

//go:embed frontend
var staticFiles embed.FS

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Println("[BOOT] Starting LocalAI RAG server...")

	log.Println("[BOOT] Initializing RAG service (Qdrant + Ollama + Embedder)...")
	service := rag.NewService()
	log.Println("[BOOT] RAG service ready")

	handler := api.NewHandler(service)
	router := api.NewRouter(handler)

	frontendFS, err := fs.Sub(staticFiles, "frontend")
	if err != nil {
		log.Fatal("[BOOT] Failed to sub frontend FS:", err)
	}
	router.PathPrefix("/").Handler(http.FileServer(http.FS(frontendFS)))

	log.Println("[BOOT] Backend + UI running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("[BOOT] Server error:", err)
	}
}
