package dal_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/patrickjmcd/lake-info/dal"
	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Runs against a real MongoDB (time-series collections need 5.0+). Skipped
// unless LAKE_INFO_TEST_MONGO_URI is set, e.g.:
//
//	docker run -d --rm -p 27017:27017 mongo:8
//	LAKE_INFO_TEST_MONGO_URI=mongodb://localhost:27017 go test ./dal/
//
// Use a throwaway server: the test creates and fills the lake-info database.
func newTestClient(t *testing.T) *dal.Client {
	t.Helper()
	uri := os.Getenv("LAKE_INFO_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("LAKE_INFO_TEST_MONGO_URI not set")
	}
	t.Setenv("ATLAS_CONNECTION_URI", uri)
	c, err := dal.New()
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func measurement(lake string, at time.Time, level float64) *lakeinfov1.LakeInfoMeasurement {
	return &lakeinfov1.LakeInfoMeasurement{
		LakeName:            lake,
		Level:               level,
		Temperature:         71.5,
		Generation:          12.25,
		TurbineReleaseRate:  3000,
		SpillwayReleaseRate: 150,
		TotalReleaseRate:    3150,
		MeasuredAt:          timestamppb.New(at),
		CreatedAt:           timestamppb.New(at.Add(5 * time.Minute)),
	}
}

func TestStoreAndGetLakeInfo(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	if err := c.Setup(ctx); err != nil {
		t.Fatalf("Setup: %v", err)
	}

	base := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	in := []*lakeinfov1.LakeInfoMeasurement{
		measurement("tablerock", base, 915.10),
		measurement("tablerock", base.Add(time.Hour), 915.20),
		measurement("tablerock", base.Add(2*time.Hour), 915.30),
		measurement("beaver", base.Add(time.Hour), 1120.40),
	}
	if err := c.StoreLakeInfo(ctx, in); err != nil {
		t.Fatalf("StoreLakeInfo: %v", err)
	}
	// Storing the same measurements again must not duplicate them.
	if err := c.StoreLakeInfo(ctx, in); err != nil {
		t.Fatalf("StoreLakeInfo (repeat): %v", err)
	}

	all, err := c.GetLakeInfo(ctx, "tablerock", time.Time{}, time.Time{}, false)
	if err != nil {
		t.Fatalf("GetLakeInfo: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("got %d tablerock measurements, want 3 (duplicates or wrong lake?)", len(all))
	}
	// Newest first, and every field round-trips.
	for i, want := range []*lakeinfov1.LakeInfoMeasurement{in[2], in[1], in[0]} {
		got := all[i]
		if got.LakeName != want.LakeName || got.Level != want.Level || got.Temperature != want.Temperature ||
			got.Generation != want.Generation || got.TurbineReleaseRate != want.TurbineReleaseRate ||
			got.SpillwayReleaseRate != want.SpillwayReleaseRate || got.TotalReleaseRate != want.TotalReleaseRate {
			t.Errorf("measurement %d: got %+v, want %+v", i, got, want)
		}
		if !got.MeasuredAt.AsTime().Equal(want.MeasuredAt.AsTime()) {
			t.Errorf("measurement %d: measuredAt %v, want %v", i, got.MeasuredAt.AsTime(), want.MeasuredAt.AsTime())
		}
		if !got.CreatedAt.AsTime().Equal(want.CreatedAt.AsTime()) {
			t.Errorf("measurement %d: createdAt %v, want %v", i, got.CreatedAt.AsTime(), want.CreatedAt.AsTime())
		}
	}

	latest, err := c.GetLakeInfo(ctx, "tablerock", time.Time{}, time.Time{}, true)
	if err != nil {
		t.Fatalf("GetLakeInfo latest: %v", err)
	}
	if len(latest) != 1 || latest[0].Level != 915.30 {
		t.Fatalf("latest: got %v, want the 915.30 reading", latest)
	}

	ranged, err := c.GetLakeInfo(ctx, "tablerock", base.Add(30*time.Minute), base.Add(90*time.Minute), false)
	if err != nil {
		t.Fatalf("GetLakeInfo range: %v", err)
	}
	if len(ranged) != 1 || ranged[0].Level != 915.20 {
		t.Fatalf("range: got %v, want only the 915.20 reading", ranged)
	}

	none, err := c.GetLakeInfo(ctx, "no-such-lake", time.Time{}, time.Time{}, false)
	if err != nil || len(none) != 0 {
		t.Fatalf("unknown lake: got %v, %v; want no results", none, err)
	}
}

func TestGetLakeInfoValidation(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	now := time.Now()

	for name, tc := range map[string]struct {
		lake       string
		start, end time.Time
		want       error
	}{
		"empty lake":      {"", time.Time{}, time.Time{}, dal.ErrLakeNameEmpty},
		"only start":      {"tablerock", now, time.Time{}, dal.ErrIncompleteTimeRange},
		"only end":        {"tablerock", time.Time{}, now, dal.ErrIncompleteTimeRange},
		"start after end": {"tablerock", now, now.Add(-time.Hour), dal.ErrInvalidTimeRange},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := c.GetLakeInfo(ctx, tc.lake, tc.start, tc.end, false); !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}
