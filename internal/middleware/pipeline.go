package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// responseWriterDelegator wraps http.ResponseWriter to capture response status codes
type responseWriterDelegator struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterDelegator) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware records entry timestamps, status codes, execution latency, and client IPs
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedWriter := &responseWriterDelegator{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default status code if WriteHeader isn't called explicitly
		}

		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start)
		log.Printf("[HTTP] %s %s | Status: %d | Duration: %v | Client: %s",
			r.Method, r.URL.Path, wrappedWriter.statusCode, duration, r.RemoteAddr)
	})
}

// RecoveryMiddleware captures runtime panics, logs stack traces, and prevents server crashes
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC RECOVERED] Error: %v\nStack Trace:\n%s", err, string(debug.Stack()))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"internal server error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles cross-origin policies and intercepts preflight OPTIONS requests
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Short-circuit browser preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Chain applies multiple middleware functions to an http.Handler in outer-to-inner order
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}