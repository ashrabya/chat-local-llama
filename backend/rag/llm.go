package rag

import (
	"context"
	"log"

	"github.com/tmc/langchaingo/llms/ollama"
)

type LLM struct {
	Model *ollama.LLM
}

func NewLLM() *LLM {
	log.Println("[LLM] initializing llama3 via Ollama...")
	m, err := ollama.New(
		ollama.WithModel("llama3"),
	)
	if err != nil {
		log.Fatalf("[LLM] failed to create Ollama LLM: %v", err)
	}
	log.Println("[LLM] ready")
	return &LLM{Model: m}
}

func (l *LLM) Ask(prompt string) string {
	resp, err := l.Model.Call(context.Background(), prompt)
	if err != nil {
		log.Printf("[LLM] ERROR calling model: %v", err)
		return "LLM error: " + err.Error()
	}
	return resp
}
