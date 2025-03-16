package traefik_geoip_metrics_middleware_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	geoipmiddleware "github.com/Cubicroots-Playground/traefik-geoip-metrics-middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeoIPMiddleware(t *testing.T) {
	// Setup.
	cfg := geoipmiddleware.CreateConfig()
	cfg.GeoIPAPI = geoIPAPIMock.URL

	ctx := context.Background()
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	middleware, err := geoipmiddleware.New(ctx, next, cfg, "geoip-plugin")
	require.NoError(t, err)

	// Execute.
	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	require.NoError(t, err)

	middleware.ServeHTTP(recorder, req)

	// Assert & clean up.
	assert.Equal(t, "DE", req.Header.Get("Geoip_country_iso"))
	assertMetricForCountry(t, "DE")

	geoIPMiddleware := middleware.(*geoipmiddleware.GeoIPMiddleware)
	require.NoError(t, geoIPMiddleware.Close())
}

func TestGeoIPMiddleware_WithInvalidGeoIPAPI(t *testing.T) {
	// Setup.
	cfg := geoipmiddleware.CreateConfig()
	cfg.GeoIPAPI = "https://localhost:65001"

	ctx := context.Background()
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	middleware, err := geoipmiddleware.New(ctx, next, cfg, "geoip-plugin")
	require.NoError(t, err)

	// Execute.
	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	require.NoError(t, err)

	middleware.ServeHTTP(recorder, req)

	// Assert & clean up.
	assert.Empty(t, req.Header.Get("Geoip_country_iso"))

	geoIPMiddleware := middleware.(*geoipmiddleware.GeoIPMiddleware)
	require.NoError(t, geoIPMiddleware.Close())
}

func assertMetricForCountry(t *testing.T, country string) {
	t.Helper()

	resp, err := http.Get("http://127.0.0.1:2112/metrics")
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Contains(t, string(body), `traefik_geoip_requests_total{country_iso="`+country+`"} `)
}
