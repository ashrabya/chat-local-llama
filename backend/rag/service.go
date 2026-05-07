package rag

import "log"

type Service struct {
	Q *QdrantStore
	L *LLM
	E *Embedder
}

func NewService() *Service {
	log.Println("[SERVICE] initializing components...")
	s := &Service{
		Q: NewQdrant(),
		L: NewLLM(),
		E: NewEmbedder(),
	}
	log.Println("[SERVICE] all components ready")
	return s
}
