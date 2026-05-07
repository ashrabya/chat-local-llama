package api

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// newRequestID generates a short random hex request ID.
func newRequestID() string {
	return fmt.Sprintf("%08x", rand.Uint32())
}

// LoggingMiddleware logs every request: method, path, request ID, status, latency, bytes.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := newRequestID()

		// Attach request ID header so clients can correlate
		w.Header().Set("X-Request-ID", reqID)

		log.Printf("[REQ ] id=%s method=%s path=%s remote=%s",
			reqID, r.Method, r.URL.Path, r.RemoteAddr)

		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		latency := time.Since(start)
		log.Printf("[RESP] id=%s status=%d bytes=%d latency=%s",
			reqID, wrapped.status, wrapped.size, latency)
	})
}
