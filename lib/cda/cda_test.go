package cda

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient points a Client at a mock CDA server.
func newTestClient(srv *httptest.Server) *Client {
	return &Client{baseURL: srv.URL, office: Office, http: srv.Client()}
}

// serveSeries returns a handler that responds per timeseries name with the
// given [epochMillis, value] rows.
func serveSeries(t *testing.T, series map[string][][2]float64) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		rows, ok := series[name]
		w.Header().Set("Content-Type", "application/json")
		if !ok {
			// Unknown series: valid but empty (mirrors CDA's empty CCP-Comp).
			_ = json.NewEncoder(w).Encode(map[string]any{"name": name, "units": "cfs", "values": [][]*float64{}})
			return
		}
		vals := make([][]*float64, 0, len(rows))
		for _, row := range rows {
			ts, v := row[0], row[1]
			q := 0.0
			vals = append(vals, []*float64{&ts, &v, &q})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"name": name, "units": "cfs", "values": vals})
	}
}

func TestGetMeasurements_AlignsByTimestamp(t *testing.T) {
	h0 := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC).UnixMilli()
	h1 := time.Date(2026, 9, 17, 13, 0, 0, 0, time.UTC).UnixMilli()

	srv := httptest.NewServer(serveSeries(t, map[string][][2]float64{
		tsElevation:  {{float64(h0), 912.30}, {float64(h1), 912.31}},
		tsGeneration: {{float64(h0), 5.0}, {float64(h1), 6.0}},
		tsTurbine:    {{float64(h0), 100.0}, {float64(h1), 110.0}},
		tsSpillway:   {{float64(h0), 0.0}, {float64(h1), 0.0}},
		// Total lags: only the earlier timestamp is published.
		tsTotal: {{float64(h0), 100.0}},
	}))
	defer srv.Close()

	c := newTestClient(srv)
	records, err := c.GetMeasurements(context.Background(),
		time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 17, 23, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetMeasurements: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
	// Sorted ascending; first row fully populated.
	first := records[0]
	if first.Level != 912.30 || first.TurbineReleaseRate != 100 || first.TotalReleaseRate != 100 {
		t.Errorf("first row = %+v, want level 912.30 turbine 100 total 100", first)
	}
	if !first.MeasuredAt.AsTime().Before(records[1].MeasuredAt.AsTime()) {
		t.Error("records are not sorted ascending by time")
	}
	// Second row: total lagged, so it stays at the zero default (gotcha #2).
	if records[1].TurbineReleaseRate != 110 {
		t.Errorf("second row turbine = %v, want 110", records[1].TurbineReleaseRate)
	}
	if records[1].TotalReleaseRate != 0 {
		t.Errorf("second row total = %v, want 0 (lagged series left unset)", records[1].TotalReleaseRate)
	}
}

func TestGetTimeseries_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, _, err := c.GetTimeseries(context.Background(), "whatever", time.Now().Add(-time.Hour), time.Now())
	if err == nil || !strings.Contains(err.Error(), "unexpected status 500") {
		t.Fatalf("expected a 500 error, got %v", err)
	}
}
