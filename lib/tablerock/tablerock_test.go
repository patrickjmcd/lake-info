package tablerock

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func fixtureServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(body))
	}))
}

func TestGetAllRecords_RealPage(t *testing.T) {
	body, err := os.ReadFile("testdata/tab7d.htm")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	srv := fixtureServer(t, string(body))
	defer srv.Close()

	records, err := GetAllRecords(srv.URL)
	if err != nil {
		t.Fatalf("GetAllRecords returned error: %v", err)
	}
	if len(records) == 0 {
		t.Fatal("expected at least one record, got 0")
	}

	// Every parsed record should have the lake name set and a sane elevation
	// for Table Rock (power pool ~917 ft, flood pool ~931 ft).
	for i, r := range records {
		if r.LakeName != LakeName {
			t.Errorf("record %d: lake name = %q, want %q", i, r.LakeName, LakeName)
		}
		if r.Level < 800 || r.Level > 1000 {
			t.Errorf("record %d: implausible level %v", i, r.Level)
		}
		if r.MeasuredAt == nil {
			t.Errorf("record %d: measuredAt is nil", i)
		}
	}
	// The fixture has 168 data rows; all should parse (the old fixed-offset
	// skip dropped the oldest row).
	if len(records) != 168 {
		t.Errorf("parsed %d records, want 168", len(records))
	}
	// The oldest row (10SEP2026 0900) must be present, at level 912.79.
	if records[0].Level != 912.79 {
		t.Errorf("first (oldest) record level = %v, want 912.79 (oldest row must not be dropped)", records[0].Level)
	}

	t.Logf("parsed %d records; first level=%v, last level=%v",
		len(records), records[0].Level, records[len(records)-1].Level)
}

func TestGetLatestRecord_RealPage(t *testing.T) {
	body, err := os.ReadFile("testdata/tab7d.htm")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	srv := fixtureServer(t, string(body))
	defer srv.Close()

	all, err := GetAllRecords(srv.URL)
	if err != nil {
		t.Fatalf("GetAllRecords returned error: %v", err)
	}
	latest, err := GetLatestRecord(srv.URL)
	if err != nil {
		t.Fatalf("GetLatestRecord returned error: %v", err)
	}
	want := all[len(all)-1]
	if latest.MeasuredAt.AsTime() != want.MeasuredAt.AsTime() || latest.Level != want.Level {
		t.Errorf("latest record = (%v, %v), want (%v, %v)",
			latest.MeasuredAt.AsTime(), latest.Level, want.MeasuredAt.AsTime(), want.Level)
	}
}

func TestGetAllRecords_MalformedPage(t *testing.T) {
	// A page with no <hr> separator must return an error, not panic.
	srv := fixtureServer(t, "<html><body>maintenance</body></html>")
	defer srv.Close()

	_, err := GetAllRecords(srv.URL)
	if err == nil {
		t.Fatal("expected an error for a page without <hr>, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected page format") {
		t.Errorf("unexpected error message: %v", err)
	}
}
