// Package cda is a spike of an alternative Table Rock data source: the USACE
// CWMS Data API (https://cwms-data.usace.army.mil/cwms-data), which serves the
// same measurements as the tab7d.htm HTML page but as structured JSON with
// units and quality flags. See SPIKE.md for the evaluation and field mapping.
package cda

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"

	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	DefaultBaseURL = "https://cwms-data.usace.army.mil/cwms-data"
	Office         = "SWL"
	LakeName       = "tablerock"
)

// Table Rock tab7d columns mapped to CWMS timeseries IDs. Each was confirmed
// against the live API and cross-checked with the HTML page. The version
// suffix differs per series (some CCP-Comp series are empty, so Regi-Comp is
// used where it is the populated one) — see SPIKE.md.
const (
	tsElevation  = "Table_Rock_Dam-Headwater.Elev.Inst.1Hour.0.Decodes-rev"
	tsGeneration = "Table_Rock_Dam.Energy-Gen_Plant.Total.1Hour.1Hour.CCP-Comp"
	tsTurbine    = "Table_Rock_Dam.Flow-Plant.Ave.1Hour.1Hour.CCP-Comp"
	tsSpillway   = "Table_Rock_Dam.Flow-Tainter Total.Ave.1Hour.1Hour.Regi-Comp"
	tsTotal      = "Table_Rock_Dam.Flow-Res Out.Ave.1Hour.1Hour.Regi-Comp"
)

// Client talks to the CWMS Data API.
type Client struct {
	baseURL string
	office  string
	http    *http.Client
}

// New returns a Client with sane defaults: the public CDA endpoint, the SWL
// office, proxy support from the environment, and normal TLS verification (the
// CDA endpoint has a valid certificate, unlike the workaround the HTML scraper
// uses).
func New() *Client {
	return &Client{
		baseURL: DefaultBaseURL,
		office:  Office,
		http: &http.Client{
			Timeout:   30 * time.Second,
			Transport: &http.Transport{Proxy: http.ProxyFromEnvironment},
		},
	}
}

// Point is a single timestamped value from a timeseries.
type Point struct {
	Time  time.Time
	Value float64
}

type cdaResponse struct {
	Name   string       `json:"name"`
	Units  string       `json:"units"`
	Values [][]*float64 `json:"values"`
}

// GetTimeseries fetches one timeseries between begin and end (inclusive),
// returning its populated points (nulls skipped) and its units.
func (c *Client) GetTimeseries(ctx context.Context, tsid string, begin, end time.Time) ([]Point, string, error) {
	q := url.Values{}
	q.Set("office", c.office)
	q.Set("name", tsid)
	q.Set("begin", begin.UTC().Format(time.RFC3339))
	q.Set("end", end.UTC().Format(time.RFC3339))
	endpoint := c.baseURL + "/timeseries?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Accept", "application/json;version=2")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, "", fmt.Errorf("unexpected status %d from CDA for %q: %s", resp.StatusCode, tsid, body)
	}

	var parsed cdaResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, "", fmt.Errorf("decoding CDA response for %q: %w", tsid, err)
	}

	var points []Point
	for _, row := range parsed.Values {
		// Each row is [epochMillis, value, quality]; value may be null.
		if len(row) < 2 || row[0] == nil || row[1] == nil {
			continue
		}
		points = append(points, Point{
			Time:  time.UnixMilli(int64(*row[0])).UTC(),
			Value: *row[1],
		})
	}
	return points, parsed.Units, nil
}

// GetMeasurements fetches the mapped timeseries for the window and aligns them
// by timestamp into LakeInfoMeasurement records, sorted ascending by time. A
// record is emitted for every timestamp that has an elevation reading; other
// fields are filled where a value exists for that same timestamp. Temperature
// is intentionally left unset — CDA only exposes tailwater temperature, not the
// lake-surface temperature the app tracks (still sourced separately).
func (c *Client) GetMeasurements(ctx context.Context, begin, end time.Time) ([]*lakeinfov1.LakeInfoMeasurement, error) {
	elev, _, err := c.GetTimeseries(ctx, tsElevation, begin, end)
	if err != nil {
		return nil, fmt.Errorf("elevation: %w", err)
	}

	byTime := map[int64]*lakeinfov1.LakeInfoMeasurement{}
	order := []int64{}
	for _, p := range elev {
		key := p.Time.UnixMilli()
		byTime[key] = &lakeinfov1.LakeInfoMeasurement{
			LakeName:   LakeName,
			Level:      p.Value,
			MeasuredAt: timestamppb.New(p.Time),
			CreatedAt:  timestamppb.Now(),
		}
		order = append(order, key)
	}

	// Fill the remaining fields onto existing (elevation-anchored) timestamps.
	fill := func(tsid string, set func(m *lakeinfov1.LakeInfoMeasurement, v float64)) error {
		pts, _, err := c.GetTimeseries(ctx, tsid, begin, end)
		if err != nil {
			return fmt.Errorf("%s: %w", tsid, err)
		}
		for _, p := range pts {
			if m, ok := byTime[p.Time.UnixMilli()]; ok {
				set(m, p.Value)
			}
		}
		return nil
	}

	if err := fill(tsGeneration, func(m *lakeinfov1.LakeInfoMeasurement, v float64) { m.Generation = v }); err != nil {
		return nil, err
	}
	if err := fill(tsTurbine, func(m *lakeinfov1.LakeInfoMeasurement, v float64) { m.TurbineReleaseRate = v }); err != nil {
		return nil, err
	}
	if err := fill(tsSpillway, func(m *lakeinfov1.LakeInfoMeasurement, v float64) { m.SpillwayReleaseRate = v }); err != nil {
		return nil, err
	}
	if err := fill(tsTotal, func(m *lakeinfov1.LakeInfoMeasurement, v float64) { m.TotalReleaseRate = v }); err != nil {
		return nil, err
	}

	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
	records := make([]*lakeinfov1.LakeInfoMeasurement, 0, len(order))
	for _, k := range order {
		records = append(records, byTime[k])
	}
	return records, nil
}

// GetLatestRecord returns the most recent aligned measurement in the window.
func (c *Client) GetLatestRecord(ctx context.Context, begin, end time.Time) (*lakeinfov1.LakeInfoMeasurement, error) {
	records, err := c.GetMeasurements(ctx, begin, end)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no records found in window")
	}
	return records[len(records)-1], nil
}
