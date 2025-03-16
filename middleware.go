// Package plugindemo a demo plugin.
package traefik_geoip_metrics_middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

// Config the plugin configuration.
type Config struct {
	GeoIPAPI    string `json:"geoipApi"`
	MetricsPort int    `json:"metricsPort"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		GeoIPAPI:    "http://geoip-api:8080",
		MetricsPort: 2112,
	}
}

// GeoIPMiddleware a GeoIPMiddleware plugin.
type GeoIPMiddleware struct {
	next     http.Handler
	name     string
	template *template.Template
	config   *Config
}

// New created a new Demo plugin.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	a := &GeoIPMiddleware{
		next:     next,
		name:     name,
		template: template.New("demo").Delims("[[", "]]"),
		config:   config,
	}

	return a, nil
}

func (a *GeoIPMiddleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Get geo data.
	headers, err := a.getGeoIPHeaders(req)
	if err != nil {
		os.Stdout.WriteString("failed to assemble geoip API request: " + err.Error() + "\n")
	}

	// Write headers.
	for key, value := range headers {
		tmpl, err := a.template.Parse(value)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		writer := &bytes.Buffer{}

		err = tmpl.Execute(writer, req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		req.Header.Set(key, writer.String())
	}

	a.next.ServeHTTP(rw, req)
}

func (a *GeoIPMiddleware) getGeoIPHeaders(req *http.Request) (map[string]string, error) {
	headers := map[string]string{}
	remoteAddr := strings.Split(req.RemoteAddr, ":")[0]

	reqCtx, cancel := context.WithTimeout(req.Context(), time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, a.config.GeoIPAPI, bytes.NewReader([]byte(remoteAddr)))
	if err != nil {
		return headers, fmt.Errorf("failed to assemble geo IP request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return headers, fmt.Errorf("failed to make geo IP request: %w", err)
	}
	defer resp.Body.Close()

	country, err := io.ReadAll(resp.Body)
	if err != nil {
		return headers, fmt.Errorf("failed to read geo IP response: %w", err)
	}

	if len(string(country)) > 0 {
		headers["GEOIP_COUNTRY_ISO"] = string(country)
	}
	headers["GEOIP_IP"] = remoteAddr

	return headers, nil
}
