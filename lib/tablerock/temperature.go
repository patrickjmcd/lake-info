package tablerock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// DefaultTemperatureURL is the White River Sky endpoint for Table Rock Lake
// surface water temperature. Temperature is not available in the USACE tabular
// data, so it is sourced separately from here.
const DefaultTemperatureURL = "https://api.whiteriversky.net/v1/temperatures/sites/5/"

// GetTemperature fetches the current surface water temperature (°F) for Table
// Rock Lake. If url is empty, DefaultTemperatureURL is used. A token is
// required; the caller is expected to supply it from configuration rather than
// hard-coding it.
func GetTemperature(url, token string) (float64, error) {
	if url == "" {
		url = DefaultTemperatureURL
	}
	if token == "" {
		return 0, fmt.Errorf("no temperature API token configured")
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Token "+token)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status %d from temperature API", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// The API may report fahrenheit as either a JSON number or a string, so
	// accept both.
	var payload struct {
		Fahrenheit interface{} `json:"fahrenheit"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, fmt.Errorf("error parsing temperature response: %w", err)
	}

	switch v := payload.Fahrenheit.(type) {
	case float64:
		return v, nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("error parsing fahrenheit value %q: %w", v, err)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("unexpected fahrenheit type %T in temperature response", v)
	}
}
