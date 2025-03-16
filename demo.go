// Package plugindemo a demo plugin.
package traefik_geoip_metrics_middleware

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"text/template"
)

// Config the plugin configuration.
type Config struct {
	Headers map[string]string `json:"headers,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Headers: make(map[string]string),
	}
}

// Demo a Demo plugin.
type Demo struct {
	next     http.Handler
	headers  map[string]string
	name     string
	template *template.Template

	requestCnt atomic.Uint64
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.Headers) == 0 {
		return nil, fmt.Errorf("headers cannot be empty")
	}

	a := &Demo{
		headers:  config.Headers,
		next:     next,
		name:     name,
		template: template.New("demo").Delims("[[", "]]"),
	}

	http.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("reqs: %d", a.requestCnt.Load())))
	}))
	go func() {
		err := http.ListenAndServe(":2112", nil)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			os.Stdout.WriteString("failed to serve metrics: " + err.Error())
		}
	}()

	return a, nil
}

func (a *Demo) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	a.requestCnt.Add(1)
	headers := a.headers

	remoteAddr := strings.Split(req.RemoteAddr, ":")[0]
	headers["GEOIP_IP"] = remoteAddr

	resp, err := http.Post("http://geoip-api:8080", "", bytes.NewReader([]byte(remoteAddr)))
	if err != nil {
		os.Stdout.WriteString("failed to query geoip API: " + err.Error())
	} else {
		defer resp.Body.Close()
		country, err := io.ReadAll(resp.Body)
		if err != nil {
			os.Stdout.WriteString("failed to read geoip API response: " + err.Error())
		} else if len(string(country)) > 0 {
			headers["GEOIP_COUNTRY_ISO"] = string(country)
		}
	}

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
