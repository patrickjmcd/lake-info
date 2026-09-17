package tablerock

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTemperature(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		status  int
		token   string
		want    float64
		wantErr bool
	}{
		{name: "numeric fahrenheit", body: `{"fahrenheit": 72.5}`, status: 200, token: "tok", want: 72.5},
		{name: "string fahrenheit", body: `{"fahrenheit": "68.2"}`, status: 200, token: "tok", want: 68.2},
		{name: "non-200", body: `{}`, status: 500, token: "tok", wantErr: true},
		{name: "bad json", body: `not json`, status: 200, token: "tok", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("Authorization"); got != "Token "+tt.token {
					t.Errorf("Authorization header = %q, want %q", got, "Token "+tt.token)
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			got, err := GetTemperature(srv.URL, tt.token)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (value %v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("temperature = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTemperature_NoToken(t *testing.T) {
	if _, err := GetTemperature("https://example.com", ""); err == nil {
		t.Fatal("expected an error when no token is configured, got nil")
	}
}
