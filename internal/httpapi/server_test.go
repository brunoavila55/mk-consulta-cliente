package httpapi

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"mk-consulta-cliente/internal/mk"
)

func TestHandlerRejectsInvalidCPF(t *testing.T) {
	t.Parallel()

	baseURL, _ := url.Parse("http://mk.invalid")
	client := mk.NewClient(baseURL, http.DefaultClient, mk.NewStaticTokenProvider("token"))
	handler := NewHandler(client, "a-valid-api-key-with-24-chars", slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	request := httptest.NewRequest(http.MethodGet, "/v1/clientes?cpf=123", nil)
	request.Header.Set("X-API-Key", "a-valid-api-key-with-24-chars")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestMetricsEndpointExposesRequestMetrics(t *testing.T) {
	t.Parallel()

	baseURL, _ := url.Parse("http://mk.invalid")
	client := mk.NewClient(baseURL, http.DefaultClient, mk.NewStaticTokenProvider("token"))
	handler := NewHandler(client, "a-valid-api-key-with-24-chars", slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	healthRequest := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthResponse := httptest.NewRecorder()
	handler.ServeHTTP(healthResponse, healthRequest)

	metricsRequest := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsResponse := httptest.NewRecorder()
	handler.ServeHTTP(metricsResponse, metricsRequest)

	if metricsResponse.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", metricsResponse.Code, http.StatusOK)
	}
	body := metricsResponse.Body.String()
	if !strings.Contains(body, `mk_consulta_cliente_http_requests_total{method="GET",route="GET /health",status="200"} 1`) {
		t.Fatalf("metrica de requisicao nao encontrada em:\n%s", body)
	}
	if !strings.Contains(body, "mk_consulta_cliente_http_request_duration_seconds_bucket") {
		t.Fatal("histograma de duracao nao encontrado")
	}
	if !strings.Contains(body, "go_goroutines") || !strings.Contains(body, "process_resident_memory_bytes") {
		t.Fatal("metricas de runtime ou processo nao encontradas")
	}
}

func TestHandlerRequiresAPIKey(t *testing.T) {
	t.Parallel()

	baseURL, _ := url.Parse("http://mk.invalid")
	client := mk.NewClient(baseURL, http.DefaultClient, mk.NewStaticTokenProvider("token"))
	handler := NewHandler(client, "a-valid-api-key-with-24-chars", slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	request := httptest.NewRequest(http.MethodGet, "/v1/clientes?cpf=52998224725", nil).WithContext(context.Background())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}
