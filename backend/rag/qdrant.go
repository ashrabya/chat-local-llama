package rag

import (
	"log"

	"github.com/qdrant/go-client/qdrant"
)

type QdrantStore struct {
	Client *qdrant.Client
}

func NewQdrant() *QdrantStore {
	log.Println("[QDRANT] connecting to localhost:6334...")
	c, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		log.Fatalf("[QDRANT] failed to create client: %v", err)
	}
	log.Println("[QDRANT] client ready")
	return &QdrantStore{Client: c}
}
