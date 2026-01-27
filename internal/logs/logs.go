package logs

import (
	"log"
	"net/http"
	"go-dsc-pull/internal/schema"
)

// loggingMiddleware logs all HTTP requests with method, path, remote addr, and status
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &schema.StatusRecorder{ResponseWriter: w, Status: 200}
		next.ServeHTTP(rec, r)
		log.Printf("[HTTP] %s %s %s %d", r.Method, r.URL.Path, r.RemoteAddr, rec.Status)
	})
}
