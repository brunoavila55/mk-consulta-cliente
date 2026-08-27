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

	"mk-consulta-cliente/internal/config"
	"mk-consulta-cliente/internal/httpapi"
	"mk-consulta-cliente/internal/mk"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	settings, err := config.Load()
	if err != nil {
		logger.Error("configuração inválida", "error", err)
		os.Exit(1)
	}

	httpClient := &http.Client{Timeout: settings.MKHTTPTimeout}
	var tokenProvider mk.TokenProvider
	if settings.MKTemporaryAuthToken != "" {
		tokenProvider = mk.NewStaticTokenProvider(settings.MKTemporaryAuthToken)
	} else {
		tokenProvider = mk.NewAuthenticationTokenProvider(
			settings.MKBaseURL, httpClient, settings.MKUserAccessToken,
			settings.MKWebserviceCounterPassword, settings.MKTemporaryAuthTokenTTL,
		)
	}

	mkClient := mk.NewClient(settings.MKBaseURL, httpClient, tokenProvider)
	server := &http.Server{
		Addr:              ":" + settings.Port,
		Handler:           httpapi.NewHandler(mkClient, settings.ChatbotAPIKey, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("API iniciada", "address", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("servidor encerrado com erro", "error", err)
			os.Exit(1)
		}
	case <-shutdownSignal.Done():
		logger.Info("encerrando API")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			logger.Error("falha ao encerrar API", "error", err)
			os.Exit(1)
		}
	}
}
