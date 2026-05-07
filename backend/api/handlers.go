package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ashrabya/chat-local-llama.git/backend/rag"
)

// Supported file extensions mapped to a human-readable label.
var supportedExts = map[string]string{
	".pdf":  "pdf",
	".txt":  "text",
	".md":   "markdown",
	".docx": "docx",
	".doc":  "doc",
	".csv":  "csv",
	".html": "html",
	".htm":  "html",
	".xml":  "xml",
	".json": "json",
}

type Handler struct {
	Service *rag.Service
}

func NewHandler(s *rag.Service) *Handler {
	return &Handler{Service: s}
}

// Chat handles POST /chat
func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	reqID := w.Header().Get("X-Request-ID")

	log.Printf("[CHAT] id=%s parsing request body", reqID)

	var req struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[CHAT] id=%s ERROR decoding body: %v", reqID, err)
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		log.Printf("[CHAT] id=%s ERROR empty message", reqID)
		http.Error(w, "message cannot be empty", http.StatusBadRequest)
		return
	}

	log.Printf("[CHAT] id=%s question=%q", reqID, truncate(req.Message, 80))
	log.Printf("[CHAT] id=%s calling RAG service...", reqID)

	answer := h.Service.Ask(req.Message)

	log.Printf("[CHAT] id=%s answer_len=%d elapsed=%s", reqID, len(answer), time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"answer": answer,
	})
}

// Upload handles POST /upload — accepts PDF, TXT, MD, DOCX, CSV, HTML, XML, JSON
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	reqID := w.Header().Get("X-Request-ID")

	log.Printf("[UPLOAD] id=%s parsing multipart form", reqID)

	// 32 MB max memory
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("[UPLOAD] id=%s ERROR parsing form: %v", reqID, err)
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("[UPLOAD] id=%s ERROR getting file: %v", reqID, err)
		http.Error(w, "missing file field: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	origName := header.Filename
	ext := strings.ToLower(filepath.Ext(origName))
	fileType, supported := supportedExts[ext]

	log.Printf("[UPLOAD] id=%s file=%q ext=%s size=%d supported=%v",
		reqID, origName, ext, header.Size, supported)

	if !supported {
		supported := make([]string, 0, len(supportedExts))
		for e := range supportedExts {
			supported = append(supported, e)
		}
		msg := fmt.Sprintf("unsupported file type %q. supported: %s", ext, strings.Join(supported, ", "))
		log.Printf("[UPLOAD] id=%s ERROR %s", reqID, msg)
		http.Error(w, msg, http.StatusUnsupportedMediaType)
		return
	}

	// Write to a temp file preserving the original extension so extractors work
	tmp, err := os.CreateTemp("", "upload-*"+ext)
	if err != nil {
		log.Printf("[UPLOAD] id=%s ERROR creating temp file: %v", reqID, err)
		http.Error(w, "server error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	written, err := io.Copy(tmp, file)
	if err != nil {
		log.Printf("[UPLOAD] id=%s ERROR writing temp file: %v", reqID, err)
		http.Error(w, "server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[UPLOAD] id=%s temp_file=%s written=%d bytes type=%s", reqID, tmp.Name(), written, fileType)
	log.Printf("[UPLOAD] id=%s starting ingestion pipeline...", reqID)

	if err := h.Service.Ingest(tmp.Name(), ext); err != nil {
		log.Printf("[UPLOAD] id=%s ERROR ingesting: %v", reqID, err)
		http.Error(w, "ingestion error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[UPLOAD] id=%s ingestion complete elapsed=%s", reqID, time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "ok",
		"filename": origName,
		"type":     fileType,
	})
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
