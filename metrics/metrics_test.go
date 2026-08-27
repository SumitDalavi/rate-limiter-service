package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCounter(t *testing.T) {
	c := &Counter{}

	if c.Value() != 0 {
		t.Errorf("expected 0, got %d", c.Value())
	}

	c.Inc()
	if c.Value() != 1 {
		t.Errorf("expected 1, got %d", c.Value())
	}

	c.Inc()
	if c.Value() != 2 {
		t.Errorf("expected 2, got %d", c.Value())
	}
}

func TestMetricsHandler(t *testing.T) {
	handler := MetricsHandler()

	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
