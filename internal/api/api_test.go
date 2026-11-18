package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bumlingo/internal/db"
)

type fakePointsReader struct {
	points []db.Point
	err    error
}

func (f fakePointsReader) ReadPoints(context.Context) ([]db.Point, error) {
	return f.points, f.err
}

func (f fakePointsReader) ReadPointsInBBox(ctx context.Context, minLon, minLat, maxLon, maxLat float64) ([]db.Point, error) {
	return f.points, f.err
}

func TestGetPoints(t *testing.T) {
	want := []db.Point{{Label: "New point", Lon: 18.0586, Lat: 59.33}}
	s := NewServer(fakePointsReader{points: want}).Handler()

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/points", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var got []db.Point
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(got) != len(want) || got[0] != want[0] {
		t.Errorf("points = %+v, want %+v", got, want)
	}
}

func TestGetPointsEmptyReturnsEmptyArray(t *testing.T) {
	s := NewServer(fakePointsReader{}).Handler()

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/points", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Errorf("body = %q, want []", body)
	}
}

func TestGetPointsError(t *testing.T) {
	s := NewServer(fakePointsReader{err: errors.New("db down")}).Handler()

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/points", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if strings.Contains(rec.Body.String(), "db down") {
		t.Error("response leaks internal error detail")
	}
}

func TestGetPointsMethodNotAllowed(t *testing.T) {
	s := NewServer(fakePointsReader{}).Handler()

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/points", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestHealthz(t *testing.T) {
	s := NewServer(fakePointsReader{}).Handler()

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
