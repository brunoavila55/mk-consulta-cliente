package httpapi

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"mk-consulta-cliente/internal/document"
	"mk-consulta-cliente/internal/mk"
)

type customerLookup func(*http.Request, string) (mk.CustomerResult, error)

func NewHandler(client *mk.Client, apiKey string, logger *slog.Logger) http.Handler {
	metrics := newAPIMetrics()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(writer http.ResponseWriter, _ *http.Request) {
		writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.Handle("GET /metrics", metrics.handler)
	mux.Handle("GET /v1/clientes", requireAPIKey(apiKey, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		handleCustomer(writer, request, func(request *http.Request, document string) (mk.CustomerResult, error) {
			return client.CustomerByDocument(request.Context(), document)
		}, logger)
	})))

	return metrics.instrument(requestLog(mux, logger))
}

func handleCustomer(writer http.ResponseWriter, request *http.Request, lookup customerLookup, logger *slog.Logger) {
	value := request.URL.Query().Get("documento")
	if value == "" {
		value = request.URL.Query().Get("cpf") // Compatibilidade com clientes existentes.
	}
	normalizedDocument, err := document.NormalizeDocument(value)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "documento_invalido", "Informe um CPF ou CNPJ válido.")
		return
	}

	result, err := lookup(request, normalizedDocument)
	if err != nil {
		logger.Error("falha na consulta ao MK", "error", err)
		var networkError net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &networkError) && networkError.Timeout()) {
			writeError(writer, http.StatusGatewayTimeout, "mk_timeout", "O ERP demorou demais para responder.")
			return
		}
		writeError(writer, http.StatusBadGateway, "mk_indisponivel", "Não foi possível consultar o ERP.")
		return
	}

	writeJSON(writer, http.StatusOK, map[string]any{"status": "ok", "dados": result})
}

func requireAPIKey(expected string, next http.Handler) http.Handler {
	expectedHash := sha256.Sum256([]byte(expected))
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		providedHash := sha256.Sum256([]byte(strings.TrimSpace(request.Header.Get("X-API-Key"))))
		if subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) != 1 {
			writeError(writer, http.StatusUnauthorized, "nao_autorizado", "Chave de acesso ausente ou inválida.")
			return
		}
		next.ServeHTTP(writer, request)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (recorder *statusRecorder) WriteHeader(status int) {
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}

func requestLog(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: writer, status: http.StatusOK}
		next.ServeHTTP(recorder, request)
		// Deliberately log only the path, never the query string containing a document.
		logger.Info("requisição HTTP", "method", request.Method, "path", request.URL.Path, "status", recorder.status, "duration", time.Since(started))
	})
}

func writeError(writer http.ResponseWriter, status int, code, message string) {
	writeJSONStatus(writer, status, map[string]any{
		"status": "erro",
		"erro":   map[string]string{"codigo": code, "mensagem": message},
	})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writeJSONStatus(writer, status, value)
}

func writeJSONStatus(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
