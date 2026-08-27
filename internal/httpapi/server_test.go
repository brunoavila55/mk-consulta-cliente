package httpapi

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
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
