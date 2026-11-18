package api

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"bumlingo/internal/db"
)

func parseBBox(s string) (minLon, minLat, maxLon, maxLat float64, err error) {
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return 0, 0, 0, 0, fmt.Errorf("bbox must be 4 comma-separated floats, got %d parts", len(parts))
	}
	vals := make([]float64, 4)
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("bbox part %s is not a number", p)
		}
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, 0, 0, 0, fmt.Errorf("bbox part %s is Nan or inf", p)
		}
		vals[i] = v
	}
	minLon, minLat, maxLon, maxLat = vals[0], vals[1], vals[2], vals[3]
	if minLon < -180 || maxLon > 180 || minLat < -90 || maxLat > 90 {
		return 0, 0, 0, 0, fmt.Errorf("bbox out of range: lon in [-180,180], lat in [-90,90]")
	}
	if minLon > maxLon || minLat > maxLat {
		return 0, 0, 0, 0, fmt.Errorf("bbox min must not exceed max (antimeridian-crossing boxes not supported)")
	}
	return minLon, minLat, maxLon, maxLat, nil
}

// getPoints handles GET /api/points and returns all points as a JSON array.
func (s *Server) getPoints(w http.ResponseWriter, r *http.Request) {
	var points []db.Point = nil
	var err error
	if raw := r.URL.Query().Get("bbox"); raw != "" {
		var minLon, minLat, maxLon, maxLat float64
		minLon, minLat, maxLon, maxLat, err = parseBBox(raw)
		if err != nil {
			writeClientError(w, http.StatusBadRequest, err.Error())
			return
		}
		points, err = s.points.ReadPointsInBBox(r.Context(), minLon, minLat, maxLon, maxLat)
	} else {
		points, err = s.points.ReadPoints(r.Context())
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if points == nil {
		points = []db.Point{} // encode as [], not null
	}
	writeJSON(w, http.StatusOK, points)
}
