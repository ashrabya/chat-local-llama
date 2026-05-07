# LocalAI — Local RAG with Ollama + Qdrant

A fully local Retrieval-Augmented Generation (RAG) system built in Go.  
Upload your documents, ask questions, get answers — no data ever leaves your machine.

```
┌─────────────────────────────────────────────────────────────┐
│                        Browser UI                           │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP
┌────────────────────────▼────────────────────────────────────┐
│                   Go HTTP Server (:8080)                     │
│                                                             │
│   POST /upload          POST /chat                          │
│        │                     │                              │
│   ┌────▼──────┐         ┌────▼──────┐                       │
│   │  Ingest   │         │   Ask     │                       │
│   │  Service  │         │  Service  │                       │
│   └────┬──────┘         └────┬──────┘                       │
│        │                     │                              │
│   ┌────▼──────┐   embed  ┌───▼───────┐                      │
│   │  Loader   │          │ Qdrant    │  similarity search   │
│   │ PDF/DOCX/ │──────────▶ VectorDB  │◀─────────────────    │
│   │  TXT      │          └───────────┘                      │
│   └───────────┘                │                            │
│                           ┌────▼──────┐                     │
│                           │  Ollama   │  llama3 / llms.LLM  │
│                           │   LLM     │                     │
│                           └───────────┘                     │
└─────────────────────────────────────────────────────────────┘
```

---

## Features

- **100% local** — Ollama for LLM + embeddings, Qdrant for vector search
- **Multi-format ingestion** — PDF, Word (`.docx`), and plain text (`.txt`)
- **langchaingo-powered** — no manual vector math; uses the standard `VectorStore`, `documentloaders`, and `textsplitter` interfaces
- **Structured logs** — every request is logged with component, level, and latency
- **Embedded UI** — the chat interface is compiled into the binary (`//go:embed frontend`)
- **Auto-collection** — Qdrant collection is created at startup if it doesn't exist

---

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| [Go](https://go.dev/dl/) | ≥ 1.22 | Build the server |
| [Ollama](https://ollama.com) | latest | Run local LLM & embeddings |
| [Qdrant](https://qdrant.tech) | latest | Vector database |
| [Docker](https://docs.docker.com/get-docker/) | optional | Run Qdrant easily |

---

## Quick Start

```bash
# 1. Clone
git clone https://github.com/ashrabya/chat-local-llama
cd chat-local-llama

# 2. Pull required Ollama models
make pull-models

# 3. Start Qdrant (Docker)
make qdrant-up

# 4. Run the server
make run
```

Open **http://localhost:8080**, upload a document, and start asking questions.

---


---

## API Reference

### `POST /upload`

Upload a document to the knowledge base.

**Request:** `multipart/form-data` with a `file` field.

**Accepted formats:** PDF (`.pdf`), Word (`.docx`), Plain text (`.txt`)

```bash
curl -X POST http://localhost:8080/upload \
  -F "file=@report.pdf"

curl -X POST http://localhost:8080/upload \
  -F "file=@notes.docx"

curl -X POST http://localhost:8080/upload \
  -F "file=@readme.txt"
```

**Response:**
```json
{ "status": "ok", "message": "\"report.pdf\" ingested successfully" }
```

---

### `POST /chat`

Ask a question about the ingested documents.

**Request:**
```json
{ "message": "What are the key findings?" }
```

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What are the key findings?"}'
```

**Response:**
```json
{ "answer": "The key findings are …" }
```

---

## Configuration

All tunable constants live in `rag/service.go`:

| Constant | Default | Description |
|----------|---------|-------------|
| `chatModel` | `llama3` | Ollama model used to answer questions |
| `embedModel` | `nomic-embed-text` | Ollama model used for embeddings |
| `qdrantHTTPURL` | `http://localhost:6333` | Qdrant REST endpoint |
| `collectionName` | `docs` | Qdrant collection name |
| `chunkSize` | `800` | Characters per chunk |
| `chunkOverlap` | `80` | Overlap between adjacent chunks |
| `topK` | `3` | Number of chunks retrieved per query |

---

## Logs

Every stage of the pipeline is logged to stdout:

```
2025/05/06 12:00:00 [INFO]  [main]  LocalAI — local RAG with Ollama + Qdrant
2025/05/06 12:00:00 [INFO]  [startup] connecting to Ollama LLM  model=llama3
2025/05/06 12:00:00 [INFO]  [startup] connecting to Qdrant  collection=docs
2025/05/06 12:00:00 [INFO]  [startup] RAG service ready ✓
2025/05/06 12:00:05 [INFO]  [http] → POST /upload  remote=127.0.0.1:54321
2025/05/06 12:00:05 [INFO]  [ingest] → received file  name="report.pdf"
2025/05/06 12:00:05 [INFO]  [loader] loading PDF: report.pdf
2025/05/06 12:00:05 [INFO]  [ingest] split into 42 chunks
2025/05/06 12:00:07 [INFO]  [ingest] ✓ stored 42 vectors
2025/05/06 12:00:07 [INFO]  [http] ← POST /upload  status=200  latency=2.1s
2025/05/06 12:00:10 [INFO]  [http] → POST /chat  remote=127.0.0.1:54321
2025/05/06 12:00:10 [INFO]  [ask] → question received  preview="What are the key findings?"
2025/05/06 12:00:10 [INFO]  [ask] retrieved 3 chunks from Qdrant
2025/05/06 12:00:11 [INFO]  [ask] ✓ answer generated  latency=980ms
2025/05/06 12:00:11 [INFO]  [http] ← POST /chat  status=200  latency=1.2s
```

---

## Make Commands

```bash
make help          # Show all available commands
make run           # Build and run the server
make build         # Compile to ./bin/localai
make dev           # Run with live-reload (requires air)
make test          # Run all tests
make lint          # Run golangci-lint
make pull-models   # Pull required Ollama models
make qdrant-up     # Start Qdrant in Docker
make qdrant-down   # Stop Qdrant Docker container
make clean         # Remove build artifacts
```

---

## Troubleshooting

**`connection refused` on startup**  
Make sure Ollama is running (`ollama serve`) and Qdrant is up (`make qdrant-up`).

**Empty answers / "no context found"**  
Upload a document first via `/upload` or the UI before asking questions.

**`unsupported file type` error**  
Only `.pdf`, `.docx`, and `.txt` are accepted.

**Slow first response**  
The first request after startup may be slow as Ollama loads the model weights into memory. Subsequent requests are faster.

---

## Tech Stack

- **[Go](https://go.dev)** — HTTP server, business logic
- **[langchaingo](https://github.com/tmc/langchaingo)** — `vectorstores/qdrant`, `documentloaders`, `textsplitter`, `embeddings`
- **[Ollama](https://ollama.com)** — local LLM (`llama3`) and embeddings (`nomic-embed-text`)
- **[Qdrant](https://qdrant.tech)** — vector database
- **[gorilla/mux](https://github.com/gorilla/mux)** — HTTP router
- **[nguyenthenguyen/docx](https://github.com/nguyenthenguyen/docx)** — DOCX text extraction


![alt text](https://file%2B.vscode-resource.vscode-cdn.net/Users/shrab/Desktop/Screenshot%202026-05-07%20at%2012.25.36.png?version%3D1778171166682)