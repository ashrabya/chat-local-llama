package rag

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/qdrant/go-client/qdrant"
)

// Ingest extracts text from any supported file type, chunks it, embeds it, and stores it in Qdrant.
func (s *Service) Ingest(path, ext string) error {
	log.Printf("[INGEST] start file=%s ext=%s", filepath.Base(path), ext)
	overall := time.Now()

	// 1. Extract text
	t0 := time.Now()
	text, err := ExtractText(path, ext)
	if err != nil {
		return fmt.Errorf("text extraction failed: %w", err)
	}
	log.Printf("[INGEST] extraction done chars=%d elapsed=%s", len(text), time.Since(t0))

	if len(text) == 0 {
		return fmt.Errorf("no text could be extracted from %s", filepath.Base(path))
	}

	// 2. Chunk
	t0 = time.Now()
	chunks := Chunk(text, 800)
	log.Printf("[INGEST] chunking done chunks=%d elapsed=%s", len(chunks), time.Since(t0))

	ctx := context.Background()

	// 3. Embed + store each chunk
	for i, chunk := range chunks {
		t0 = time.Now()
		vec := s.E.Embed(ctx, chunk)
		log.Printf("[INGEST] chunk=%d/%d embed_dims=%d elapsed=%s", i+1, len(chunks), len(vec), time.Since(t0))

		pointID := rand.Int()

		t0 = time.Now()
		_, err := s.Q.Client.Upsert(ctx, &qdrant.UpsertPoints{
			CollectionName: "docs",
			Points: []*qdrant.PointStruct{
				{
					Id:      qdrant.NewIDNum(uint64(pointID)),
					Vectors: qdrant.NewVectors(vec...),
					Payload: map[string]*qdrant.Value{
						"text":   qdrant.NewValueString(chunk),
						"source": qdrant.NewValueString(filepath.Base(path)),
						"type":   qdrant.NewValueString(ext),
					},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("qdrant upsert chunk %d: %w", i, err)
		}
		log.Printf("[INGEST] chunk=%d/%d upserted elapsed=%s", i+1, len(chunks), time.Since(t0))
	}

	log.Printf("[INGEST] complete chunks=%d total_elapsed=%s", len(chunks), time.Since(overall))
	return nil
}
