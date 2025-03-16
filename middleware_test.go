package traefik_geoip_metrics_middleware_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	geoipmiddleware "github.com/Cubicroots-Playground/traefik-geoip-metrics-middleware"
)

func TestGeoIPMiddleware(t *testing.T) {
	// Setup.
	cfg := geoipmiddleware.CreateConfig()
	cfg.GeoIPAPI = geoIPAPIMock.URL

	ctx := context.Background()
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	middleware, err := geoipmiddleware.New(ctx, next, cfg, "geoip-plugin")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Execute.
	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	middleware.ServeHTTP(recorder, req)

	// Assert & clean up.
	if req.Header.Get("Geoip_country_iso") != "DE" {
		t.Errorf("expected DE got '%s'", req.Header.Get("Geoip_country_iso"))
	}
	assertMetricForCountry(t, "DE")

	geoIPMiddleware := middleware.(*geoipmiddleware.GeoIPMiddleware)
	err = geoIPMiddleware.Close()
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestGeoIPMiddleware_WithInvalidGeoIPAPI(t *testing.T) {
	// Setup.
	cfg := geoipmiddleware.CreateConfig()
	cfg.GeoIPAPI = "https://localhost:65001"

	ctx := context.Background()
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	middleware, err := geoipmiddleware.New(ctx, next, cfg, "geoip-plugin")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Execute.
	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	middleware.ServeHTTP(recorder, req)

	// Assert & clean up.
	if req.Header.Get("Geoip_country_iso") != "" {
		t.Errorf("expected empty got '%s'", req.Header.Get("Geoip_country_iso"))
	}

	geoIPMiddleware := middleware.(*geoipmiddleware.GeoIPMiddleware)
	err = geoIPMiddleware.Close()
	if err != nil {
		t.Fatal(err.Error())
	}
}

func assertMetricForCountry(t *testing.T, country string) {
	t.Helper()

	resp, err := http.Get("http://127.0.0.1:2112/metrics")
	if err != nil {
		t.Fatal(err.Error())
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer resp.Body.Close()

	if !strings.Contains(string(body), `traefik_geoip_requests_total{country_iso="`+country+`"} `) {
		t.Errorf("missing metric for country '%s'", country)
	}
}
