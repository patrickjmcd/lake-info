package measurement

import (
	"errors"
	"testing"
)

func TestParseMeasurement_TimezoneAndFields(t *testing.T) {
	// Columns: date, time, level, tailwater, generation, turbine, spillway, total.
	row := []string{"10SEP2026", "0900", "912.79", "702.60", "0", "20", "0", "20"}

	m, err := ParseMeasurement(row, "tablerock")
	if err != nil {
		t.Fatalf("ParseMeasurement returned error: %v", err)
	}

	if m.Level != 912.79 {
		t.Errorf("level = %v, want 912.79", m.Level)
	}
	if m.Generation != 0 {
		t.Errorf("generation = %v, want 0", m.Generation)
	}
	if m.TurbineReleaseRate != 20 {
		t.Errorf("turbineReleaseRate = %v, want 20", m.TurbineReleaseRate)
	}
	if m.TotalReleaseRate != 20 {
		t.Errorf("totalReleaseRate = %v, want 20", m.TotalReleaseRate)
	}

	got := m.MeasuredAt.AsTime()
	// Timestamps are parsed in America/Chicago. September = CDT (UTC-5),
	// so 09:00 local is 14:00 UTC.
	if got.UTC().Hour() != 14 {
		t.Errorf("measuredAt UTC hour = %d, want 14 (09:00 CDT)", got.UTC().Hour())
	}
	if y, mo, d := got.UTC().Date(); y != 2026 || mo.String() != "September" || d != 10 {
		t.Errorf("measuredAt UTC date = %04d-%s-%02d, want 2026-September-10", y, mo, d)
	}
}

func TestParseMeasurement_2400RollsToNextDay(t *testing.T) {
	row := []string{"10SEP2026", "2400", "912.79", "702.60", "0", "20", "0", "20"}
	m, err := ParseMeasurement(row, "tablerock")
	if err != nil {
		t.Fatalf("ParseMeasurement returned error: %v", err)
	}
	// 2400 on the 10th is midnight at the start of the 11th (CDT), = 05:00 UTC.
	got := m.MeasuredAt.AsTime().UTC()
	if got.Day() != 11 || got.Hour() != 5 {
		t.Errorf("2400 rollover = %v, want day 11 hour 05 UTC", got)
	}
}

func TestParseMeasurement_WrongFieldCount(t *testing.T) {
	_, err := ParseMeasurement([]string{"10SEP2026", "0900", "912.79"}, "tablerock")
	if !errors.Is(err, ErrInvalidMeasurement) {
		t.Errorf("error = %v, want ErrInvalidMeasurement", err)
	}
}
