// Package metrics define as métricas Prometheus expostas em /metrics.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal conta requisições recebidas pela API, por rota,
	// método e status HTTP.
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total de requisições HTTP recebidas.",
	}, []string{"method", "path", "status"})

	// HTTPRequestDuration mede a latência de cada requisição recebida.
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duração das requisições HTTP recebidas, em segundos.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	// HTTPRequestsInFlight mede quantas requisições estão sendo
	// processadas neste instante.
	HTTPRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Número de requisições HTTP em processamento neste momento.",
	})

	// MKRequestsTotal conta chamadas feitas à API da MK Solutions, por
	// endpoint e resultado (sucesso/erro).
	MKRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mk_requests_total",
		Help: "Total de chamadas HTTP feitas à API da MK Solutions.",
	}, []string{"endpoint", "outcome"})

	// MKRequestDuration mede a latência de cada chamada à API da MK
	// Solutions, incluindo retries.
	MKRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "mk_request_duration_seconds",
		Help:    "Duração das chamadas à API da MK Solutions, em segundos.",
		Buckets: prometheus.DefBuckets,
	}, []string{"endpoint"})

	// MKRequestRetriesTotal conta quantas tentativas extras (além da
	// primeira) foram necessárias em chamadas à MK Solutions.
	MKRequestRetriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mk_request_retries_total",
		Help: "Total de retentativas feitas em chamadas à API da MK Solutions.",
	}, []string{"endpoint"})

	// MKTokenCacheTotal conta hits/misses do cache do token de acesso MK.
	MKTokenCacheTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mk_token_cache_total",
		Help: "Total de consultas ao cache do token de acesso MK, por resultado (hit/miss).",
	}, []string{"result"})

	// MKTokenRefreshTotal conta autenticações feitas contra
	// WSAutenticacao.rule, por resultado.
	MKTokenRefreshTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mk_token_refresh_total",
		Help: "Total de autenticações feitas contra WSAutenticacao.rule.",
	}, []string{"outcome"})

	// ConsultaClienteResultadoTotal conta o resultado de negócio de cada
	// consulta: tipo (Lead/Cliente) devolvido ao chatbot.
	ConsultaClienteResultadoTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "consulta_cliente_resultado_total",
		Help: "Total de consultas de cliente, por tipo de resultado devolvido.",
	}, []string{"tipo"})
)
