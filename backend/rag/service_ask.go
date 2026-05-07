package rag

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/qdrant/go-client/qdrant"
)

func (s *Service) Ask(question string) string {
	overall := time.Now()
	log.Printf("[ASK] start question=%q", truncateStr(question, 80))

	ctx := context.Background()

	// 1. Embed the question
	t0 := time.Now()
	qVec := s.E.Embed(ctx, question)
	log.Printf("[ASK] embedding done dims=%d elapsed=%s", len(qVec), time.Since(t0))

	// 2. Search Qdrant for nearest chunks
	t0 = time.Now()
	search, err := s.Q.Client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: "docs",
		Query:          qdrant.NewQuery(qVec...),
		Limit:          qdrant.PtrOf(uint64(3)),
		WithPayload: &qdrant.WithPayloadSelector{
			SelectorOptions: &qdrant.WithPayloadSelector_Enable{
				Enable: true,
			},
		},
	})
	if err != nil {
		log.Printf("[ASK] ERROR qdrant search: %v", err)
		return "search error: " + err.Error()
	}
	log.Printf("[ASK] qdrant search done hits=%d elapsed=%s", len(search), time.Since(t0))

	// 3. Build context from retrieved chunks
	var contextText string
	for i, hit := range search {
		text := hit.Payload["text"].GetStringValue()
		source := hit.Payload["source"].GetStringValue()
		log.Printf("[ASK] hit[%d] source=%q score=%.4f chars=%d", i, source, hit.Score, len(text))
		contextText += text + "\n"
	}

	if contextText == "" {
		log.Printf("[ASK] WARNING no context found, answering without RAG")
	}

	// 4. Build prompt
	prompt := fmt.Sprintf(`
You are a helpful assistant.

Use the provided context when it is relevant to the question.
If the context does not contain the answer, use your general knowledge.

When possible:
- Prefer information from the context
- Mention when information comes from general knowledge instead of the context

Context:
%s

Question:
%s
`, contextText, question)

	// 5. Call LLM
	t0 = time.Now()
	log.Printf("[ASK] calling LLM prompt_chars=%d", len(prompt))
	answer := s.L.Ask(prompt)
	log.Printf("[ASK] LLM response answer_chars=%d elapsed=%s", len(answer), time.Since(t0))

	log.Printf("[ASK] complete total_elapsed=%s", time.Since(overall))
	return answer
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
