package traefik_geoip_metrics_middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var geoIPAPIMock *httptest.Server

func TestMain(m *testing.M) {
	geoIPAPIMock = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("DE"))
	}))

	m.Run()

	geoIPAPIMock.Close()
}
