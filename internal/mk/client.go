package mk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"consulta-cliente/internal/metrics"
)

// Client encapsula as chamadas HTTP à API da MK usadas pela consulta de
// cliente.
type Client struct {
	httpClient  *http.Client
	baseURL     string
	sys         string
	maxAttempts int
}

func NewClient(httpClient *http.Client, baseURL, sys string, maxAttempts int) *Client {
	return &Client{httpClient: httpClient, baseURL: baseURL, sys: sys, maxAttempts: maxAttempts}
}

// Cadastro é a resposta de WSMKConsultaDoc.rule.
type Cadastro struct {
	Status       string          `json:"status"`
	Situacao     string          `json:"Situacao"`
	CodigoPessoa FlexibleID      `json:"CodigoPessoa"`
	Endereco     string          `json:"Endereco"`
	Nome         string          `json:"Nome"`
	Outros       []OutroCadastro `json:"Outros"`
}

type OutroCadastro struct {
	CodigoPessoa FlexibleID `json:"CodigoPessoa"`
	Endereco     string     `json:"Endereco"`
	Nome         string     `json:"Nome"`
}

// Conexoes é a resposta de WSMKConexoesPorCliente.rule.
type Conexoes struct {
	Status   string `json:"status"`
	Conexoes []struct {
		CodConexao string `json:"codconexao"`
		Endereco   string `json:"endereco"`
		Bloqueada  string `json:"bloqueada"`
	} `json:"Conexoes"`
}

func (c *Client) ConsultaDoc(ctx context.Context, token, doc string) (*Cadastro, error) {
	q := url.Values{}
	q.Set("sys", c.sys)
	q.Set("token", token)
	q.Set("doc", doc)

	var out Cadastro
	if err := c.get(ctx, "WSMKConsultaDoc", "/mk/WSMKConsultaDoc.rule", q, &out); err != nil {
		return nil, fmt.Errorf("WSMKConsultaDoc.rule: %w", err)
	}
	return &out, nil
}

func (c *Client) ConexoesPorCliente(ctx context.Context, token, codCliente string) (*Conexoes, error) {
	q := url.Values{}
	q.Set("sys", c.sys)
	q.Set("token", token)
	q.Set("cd_cliente", codCliente)

	var out Conexoes
	if err := c.get(ctx, "WSMKConexoesPorCliente", "/mk/WSMKConexoesPorCliente.rule", q, &out); err != nil {
		return nil, fmt.Errorf("WSMKConexoesPorCliente.rule: %w", err)
	}
	return &out, nil
}

func (c *Client) get(ctx context.Context, metricName, path string, q url.Values, out any) error {
	endpoint := c.baseURL + path + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("montar requisição: %w", err)
	}

	start := time.Now()
	res, attempts, err := doWithRetry(ctx, c.httpClient, req, c.maxAttempts)
	metrics.MKRequestDuration.WithLabelValues(metricName).Observe(time.Since(start).Seconds())
	if attempts > 1 {
		metrics.MKRequestRetriesTotal.WithLabelValues(metricName).Add(float64(attempts - 1))
	}
	if err != nil {
		metrics.MKRequestsTotal.WithLabelValues(metricName, "erro").Inc()
		return fmt.Errorf("executar requisição: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		metrics.MKRequestsTotal.WithLabelValues(metricName, "erro").Inc()
		return fmt.Errorf("resposta HTTP inesperada: %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		metrics.MKRequestsTotal.WithLabelValues(metricName, "erro").Inc()
		return fmt.Errorf("decodificar resposta: %w", err)
	}

	metrics.MKRequestsTotal.WithLabelValues(metricName, "sucesso").Inc()
	return nil
}
