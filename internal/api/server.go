// Package api implements the HTTP API exposing the GIS data.
package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"bumlingo/internal/db"
)

// PointsReader fetches all points. It is implemented by an adapter over the
// database in cmd/api and faked in tests.
type PointsReader interface {
	ReadPoints(ctx context.Context) ([]db.Point, error)
	ReadPointsInBBox(ctx context.Context, minLon, minLat, maxLon, maxLat float64) ([]db.Point, error)
}

// Server holds the dependencies of the HTTP handlers.
type Server struct {
	points PointsReader
}

// NewServer returns a Server that reads points via r.
func NewServer(r PointsReader) *Server {
	return &Server{points: r}
}

// Handler returns the HTTP handler with all routes registered.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /api/points", s.getPoints)
	return logRequests(mux)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

// writeJSON sends v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

// writeError logs the real error server-side and sends only a generic
// message to the client.
func writeError(w http.ResponseWriter, status int, err error) {
	log.Printf("%s: %v", http.StatusText(status), err)
	writeJSON(w, status, map[string]string{"error": http.StatusText(status)})
}

func writeClientError(w http.ResponseWriter, status int, msg string) {
	log.Printf("%s: %v", http.StatusText(status), msg)
	writeJSON(w, status, map[string]string{"error": msg})
}

// responseWriter captures the status code for logging.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// logRequests logs one line per request: method, path, status, duration.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start).Round(time.Microsecond))
	})
}
