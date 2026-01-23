package schema

import "net/http"

// statusRecorder wraps http.ResponseWriter to capture status code
// Centralisé pour usage middleware

type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *StatusRecorder) WriteHeader(code int) {
	r.Status = code
	r.ResponseWriter.WriteHeader(code)
}
