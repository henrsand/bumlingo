package api

import (
	"net/http"

	"henrik_gis/internal/db"
)

// getPoints handles GET /api/points and returns all points as a JSON array.
func (s *Server) getPoints(w http.ResponseWriter, r *http.Request) {
	points, err := s.points.ReadPoints(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if points == nil {
		points = []db.Point{} // encode as [], not null
	}
	writeJSON(w, http.StatusOK, points)
}
