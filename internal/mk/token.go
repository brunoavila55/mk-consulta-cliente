package mk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// TokenManager obtém e armazena em cache o token de acesso da MK
// (WSAutenticacao.rule), evitando uma autenticação nova a cada requisição.
// A autenticação usa duas credenciais fixas do sistema MK: a contra-senha
// (parâmetro "token" da própria WSAutenticacao) e a senha ("password").
type TokenManager struct {
	client      *http.Client
	baseURL     string
	sys         string
	contraSenha string
	password    string
	cdServico   string
	ttl         time.Duration
	maxAttempts int

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func NewTokenManager(client *http.Client, baseURL, sys, contraSenha, password, cdServico string, ttl time.Duration, maxAttempts int) *TokenManager {
	return &TokenManager{
		client:      client,
		baseURL:     baseURL,
		sys:         sys,
		contraSenha: contraSenha,
		password:    password,
		cdServico:   cdServico,
		ttl:         ttl,
		maxAttempts: maxAttempts,
	}
}

// Get devolve um token válido, reaproveitando o cache quando ainda não
// expirou e buscando um novo na MK quando necessário.
func (tm *TokenManager) Get(ctx context.Context) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.token != "" && time.Now().Before(tm.expiresAt) {
		return tm.token, nil
	}

	token, err := tm.fetch(ctx)
	if err != nil {
		return "", err
	}

	tm.token = token
	tm.expiresAt = time.Now().Add(tm.ttl)
	return tm.token, nil
}

// Invalidate descarta o token em cache, forçando nova autenticação na
// próxima chamada a Get.
func (tm *TokenManager) Invalidate() {
	tm.mu.Lock()
	tm.token = ""
	tm.mu.Unlock()
}

func (tm *TokenManager) fetch(ctx context.Context) (string, error) {
	q := url.Values{}
	q.Set("sys", tm.sys)
	q.Set("token", tm.contraSenha)
	q.Set("password", tm.password)
	q.Set("cd_servico", tm.cdServico)

	endpoint := tm.baseURL + "/mk/WSAutenticacao.rule?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("montar requisição de autenticação MK: %w", err)
	}

	res, err := doWithRetry(ctx, tm.client, req, tm.maxAttempts)
	if err != nil {
		return "", fmt.Errorf("chamar WSAutenticacao.rule: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("WSAutenticacao.rule retornou HTTP %d", res.StatusCode)
	}

	var out struct {
		Token string `json:"Token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decodificar resposta de WSAutenticacao.rule: %w", err)
	}
	if out.Token == "" {
		return "", fmt.Errorf("WSAutenticacao.rule não retornou token de acesso")
	}

	return out.Token, nil
}
