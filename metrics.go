package traefik_geoip_metrics_middleware

import (
	"net/http"
	"os"
	"text/template"
)

var metricsTemplateRaw = `
# HELP traefik_geoip_requests_total Counts total requests by country code.
# TYPE traefik_geoip_requests_total counter
{{- range $country, $value := .MetricRequestsPerCountry }}
traefik_geoip_requests_total{country_iso="{{ $country }}"} {{ $value }}
{{- end }}
`

var metricsTemplate *template.Template

func init() {
	var err error
	metricsTemplate = template.New("metrics")
	metricsTemplate, err = metricsTemplate.Parse(metricsTemplateRaw)
	if err != nil {
		panic("failed to load metrics template: " + err.Error())
	}
}

func (a *GeoIPMiddleware) MetricsHander(w http.ResponseWriter, _ *http.Request) {
	a.metricRequestsPerCountryLock.Lock()
	data := struct {
		MetricRequestsPerCountry map[string]uint64
	}{
		MetricRequestsPerCountry: make(map[string]uint64, len(a.metricRequestsPerCountry)),
	}
	for k, v := range a.metricRequestsPerCountry {
		data.MetricRequestsPerCountry[k] = v
	}
	a.metricRequestsPerCountryLock.Unlock()

	err := metricsTemplate.ExecuteTemplate(w, "metrics", data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		os.Stdout.WriteString("failed to template metrics: " + err.Error() + "\n")
		return
	}
}
