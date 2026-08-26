package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config reúne toda a configuração da aplicação, lida a partir de variáveis
// de ambiente. Nenhuma credencial fica hardcoded no código-fonte.
type Config struct {
	Port string

	MKBaseURL     string
	MKSys         string
	MKContraSenha string // credencial "token" da WSAutenticacao.rule
	MKPassword    string
	MKCdServico   string
	MKTokenTTL    time.Duration

	APIKey string // key exigida do chatbot que dispara a API

	HTTPTimeout     time.Duration
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration

	MKRetryAttempts int
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		MKBaseURL:   getEnv("MK_BASE_URL", "https://sac.newlifefibra.com.br"),
		MKSys:       getEnv("MK_SYS", "MK0"),
		MKCdServico: getEnv("MK_CD_SERVICO", "9999"),
	}

	var err error
	cfg.MKContraSenha, err = requireEnv("MK_CONTRA_SENHA")
	if err != nil {
		return nil, err
	}
	cfg.MKPassword, err = requireEnv("MK_PASSWORD")
	if err != nil {
		return nil, err
	}
	cfg.APIKey, err = requireEnv("API_KEY")
	if err != nil {
		return nil, err
	}

	cfg.MKTokenTTL, err = getEnvDuration("MK_TOKEN_TTL", 20*time.Minute)
	if err != nil {
		return nil, err
	}
	cfg.HTTPTimeout, err = getEnvDuration("HTTP_TIMEOUT", 15*time.Second)
	if err != nil {
		return nil, err
	}
	cfg.RequestTimeout, err = getEnvDuration("REQUEST_TIMEOUT", 20*time.Second)
	if err != nil {
		return nil, err
	}
	cfg.ShutdownTimeout, err = getEnvDuration("SHUTDOWN_TIMEOUT", 15*time.Second)
	if err != nil {
		return nil, err
	}

	cfg.MKRetryAttempts, err = getEnvInt("MK_RETRY_ATTEMPTS", 2)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("variável de ambiente obrigatória não definida: %s", key)
	}
	return v, nil
}

func getEnvInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("variável de ambiente %s inválida: %w", key, err)
	}
	return n, nil
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("variável de ambiente %s inválida: %w", key, err)
	}
	return d, nil
}
