# ─────────────────────────────────────────────────────────────────────────────
#  LocalAI Makefile
# ─────────────────────────────────────────────────────────────────────────────

BINARY     := localai
BUILD_DIR  := ./bin
MAIN       := ./backend
DOCKER_IMG := qdrant/qdrant

OLLAMA_CHAT_MODEL  := llama3
OLLAMA_EMBED_MODEL := nomic-embed-text

# Detect OS for open command
UNAME := $(shell uname -s)
ifeq ($(UNAME), Darwin)
  OPEN := open
else
  OPEN := xdg-open
endif

.DEFAULT_GOAL := help

# ─────────────────────────────────────────────────────────────────────────────
#  Help
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: help
help: ## Show this help message
	@echo ""
	@echo "  LocalAI — Local RAG with Ollama + Qdrant"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

# ─────────────────────────────────────────────────────────────────────────────
#  Build & Run
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: run
run: ## Build and run the server (http://localhost:8080)
	@echo "→ Starting LocalAI..."
	go run $(MAIN)

.PHONY: build
build: ## Compile the binary to ./bin/localai
	@mkdir -p $(BUILD_DIR)
	@echo "→ Building $(BINARY)..."
	go build -o $(BUILD_DIR)/$(BINARY) $(MAIN)
	@echo "✓ Binary: $(BUILD_DIR)/$(BINARY)"

.PHONY: start
start: build ## Build then run the compiled binary
	$(BUILD_DIR)/$(BINARY)

.PHONY: dev
dev: ## Run with live-reload using 'air' (go install github.com/air-verse/air@latest)
	@which air > /dev/null || (echo "Install air: go install github.com/air-verse/air@latest" && exit 1)
	air -c .air.toml

# ─────────────────────────────────────────────────────────────────────────────
#  Dependencies
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: deps
deps: ## Download and tidy Go dependencies
	@echo "→ Tidying Go modules..."
	cd backend && go mod download && go mod tidy
	@echo "✓ Done"

.PHONY: pull-models
pull-models: ## Pull required Ollama models (llama3 + nomic-embed-text)
	@echo "→ Pulling Ollama models..."
	ollama pull $(OLLAMA_CHAT_MODEL)
	ollama pull $(OLLAMA_EMBED_MODEL)
	@echo "✓ Models ready"

# ─────────────────────────────────────────────────────────────────────────────
#  Qdrant
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: qdrant-up
qdrant-up: ## Start Qdrant vector DB in Docker (ports 6333 REST, 6334 gRPC)
	@echo "→ Starting Qdrant..."
	docker run -d \
		--name qdrant \
		--restart unless-stopped \
		-p 6333:6333 \
		-p 6334:6334 \
		-v qdrant_storage:/qdrant/storage \
		$(DOCKER_IMG)
	@echo "✓ Qdrant running at http://localhost:6333"
	@echo "  Dashboard: http://localhost:6333/dashboard"

.PHONY: qdrant-down
qdrant-down: ## Stop and remove the Qdrant Docker container
	@echo "→ Stopping Qdrant..."
	docker stop qdrant && docker rm qdrant
	@echo "✓ Qdrant stopped"

.PHONY: qdrant-logs
qdrant-logs: ## Tail Qdrant container logs
	docker logs -f qdrant

.PHONY: qdrant-reset
qdrant-reset: ## Delete all data in Qdrant (drops the 'docs' collection)
	@echo "⚠  Deleting Qdrant 'docs' collection..."
	curl -s -X DELETE http://localhost:6333/collections/docs | python3 -m json.tool
	@echo "✓ Collection deleted"

# ─────────────────────────────────────────────────────────────────────────────
#  Testing
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run all tests
	cd backend && go test ./... -v -count=1

.PHONY: test-short
test-short: ## Run tests, skipping slow integration tests
	cd backend && go test ./... -short

.PHONY: coverage
coverage: ## Run tests and open HTML coverage report
	cd backend && go test ./... -coverprofile=coverage.out
	cd backend && go tool cover -html=coverage.out -o coverage.html
	$(OPEN) backend/coverage.html

# ─────────────────────────────────────────────────────────────────────────────
#  Code Quality
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: lint
lint: ## Run golangci-lint (go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@which golangci-lint > /dev/null || (echo "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	cd backend && golangci-lint run ./...

.PHONY: fmt
fmt: ## Format all Go source files
	cd backend && gofmt -w .

.PHONY: vet
vet: ## Run go vet
	cd backend && go vet ./...

# ─────────────────────────────────────────────────────────────────────────────
#  Manual API smoke tests
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: smoke-upload-pdf
smoke-upload-pdf: ## Upload a test PDF (FILE=path/to/file.pdf)
	@test -n "$(FILE)" || (echo "Usage: make smoke-upload-pdf FILE=path/to/file.pdf" && exit 1)
	curl -s -X POST http://localhost:8080/upload \
		-F "file=@$(FILE)" | python3 -m json.tool

.PHONY: smoke-upload-docx
smoke-upload-docx: ## Upload a test DOCX (FILE=path/to/file.docx)
	@test -n "$(FILE)" || (echo "Usage: make smoke-upload-docx FILE=path/to/file.docx" && exit 1)
	curl -s -X POST http://localhost:8080/upload \
		-F "file=@$(FILE)" | python3 -m json.tool

.PHONY: smoke-chat
smoke-chat: ## Send a test chat message (MSG="your question")
	@test -n "$(MSG)" || (echo "Usage: make smoke-chat MSG=\"your question\"" && exit 1)
	curl -s -X POST http://localhost:8080/chat \
		-H "Content-Type: application/json" \
		-d "{\"message\": \"$(MSG)\"}" | python3 -m json.tool

.PHONY: smoke-health
smoke-health: ## Check if the server is up
	@curl -sf http://localhost:8080 > /dev/null \
		&& echo "✓ Server is up at http://localhost:8080" \
		|| echo "✗ Server is not responding"

# ─────────────────────────────────────────────────────────────────────────────
#  Cleanup
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: clean
clean: ## Remove build artifacts and temp files
	@echo "→ Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f backend/coverage.out backend/coverage.html
	@echo "✓ Clean"

.PHONY: clean-all
clean-all: clean qdrant-down ## Clean build artifacts AND stop Qdrant
