package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"consulta-cliente/internal/config"
	"consulta-cliente/internal/handler"
	"consulta-cliente/internal/mk"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuração inválida", "error", err)
		os.Exit(1)
	}

	httpClient := &http.Client{
		Timeout: cfg.HTTPTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	tokenMgr := mk.NewTokenManager(
		httpClient,
		cfg.MKBaseURL,
		cfg.MKSys,
		cfg.MKContraSenha,
		cfg.MKPassword,
		cfg.MKCdServico,
		cfg.MKTokenTTL,
		cfg.MKRetryAttempts,
	)

	mkClient := mk.NewClient(httpClient, cfg.MKBaseURL, cfg.MKSys, cfg.MKRetryAttempts)

	consultaCliente := &handler.ConsultaCliente{
		MK:             mkClient,
		Tokens:         tokenMgr,
		APIKey:         cfg.APIKey,
		RequestTimeout: cfg.RequestTimeout,
		Logger:         logger,
	}

	mux := http.NewServeMux()
	mux.Handle("/consulta-cliente", consultaCliente)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           withRequestLogging(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("servidor iniciado", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("erro no servidor", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	stop()
	logger.Info("encerrando servidor")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("falha ao encerrar servidor de forma graciosa", "error", err)
		os.Exit(1)
	}

	logger.Info("servidor encerrado")
}

// withRequestLogging registra método, rota, status e duração de cada
// requisição — sem expor parâmetros de query (podem conter doc/key).
func withRequestLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		logger.Info("requisição atendida",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
