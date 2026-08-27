package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type apiMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inFlight prometheus.Gauge
	handler  http.Handler
}

func newAPIMetrics() *apiMetrics {
	registry := prometheus.NewRegistry()
	metrics := &apiMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "mk_consulta_cliente",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total de requisicoes HTTP recebidas pela API.",
		}, []string{"method", "route", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "mk_consulta_cliente",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Duracao das requisicoes HTTP da API em segundos.",
			Buckets:   []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}, []string{"method", "route", "status"}),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "mk_consulta_cliente",
			Subsystem: "http",
			Name:      "in_flight_requests",
			Help:      "Quantidade de requisicoes HTTP atualmente em processamento.",
		}),
	}
	registry.MustRegister(metrics.requests, metrics.duration, metrics.inFlight)
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	metrics.handler = promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	return metrics
}

func (metrics *apiMetrics) instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// A coleta do proprio endpoint nao entra nas metricas da aplicacao.
		if request.URL.Path == "/metrics" {
			next.ServeHTTP(writer, request)
			return
		}

		started := time.Now()
		metrics.inFlight.Inc()
		defer metrics.inFlight.Dec()

		recorder := &statusRecorder{ResponseWriter: writer, status: http.StatusOK}
		next.ServeHTTP(recorder, request)
		status := strconv.Itoa(recorder.status)
		route := metricRoute(request)
		metrics.requests.WithLabelValues(request.Method, route, status).Inc()
		metrics.duration.WithLabelValues(request.Method, route, status).Observe(time.Since(started).Seconds())
	})
}

func metricRoute(request *http.Request) string {
	if pattern := request.Pattern; pattern != "" {
		return pattern
	}
	return "unmatched"
}
