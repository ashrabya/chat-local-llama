package rag

import (
	"context"
	"log"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
)

type Embedder struct {
	Model embeddings.Embedder
}

func NewEmbedder() *Embedder {
	log.Println("[EMBEDDER] initializing nomic-embed-text via Ollama...")
	llm, err := ollama.New(
		ollama.WithModel("nomic-embed-text"),
	)
	if err != nil {
		log.Fatalf("[EMBEDDER] failed to create Ollama LLM: %v", err)
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatalf("[EMBEDDER] failed to create embedder: %v", err)
	}

	log.Println("[EMBEDDER] ready")
	return &Embedder{Model: e}
}

func (e *Embedder) Embed(ctx context.Context, text string) []float32 {
	vecs, err := e.Model.EmbedDocuments(ctx, []string{text})
	if err != nil {
		log.Printf("[EMBEDDER] ERROR embedding text: %v", err)
		return nil
	}
	if len(vecs) == 0 {
		log.Printf("[EMBEDDER] WARNING empty embedding returned")
		return nil
	}
	return vecs[0]
}
