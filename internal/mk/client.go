package mk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

const maxResponseBody = 2 << 20

type Person struct {
	CEP          string `json:"cep"`
	PersonCode   int64  `json:"codigo_pessoa"`
	Email        string `json:"email"`
	Address      string `json:"endereco"`
	Phone        string `json:"fone"`
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
	Name         string `json:"nome"`
	Registration string `json:"situacao,omitempty"`
}

type CustomerResult struct {
	Customer Person   `json:"cliente"`
	Others   []Person `json:"outros"`
}

type upstreamPerson struct {
	CEP          string `json:"CEP"`
	PersonCode   int64  `json:"CodigoPessoa"`
	Email        string `json:"Email"`
	Address      string `json:"Endereco"`
	Phone        string `json:"Fone"`
	Latitude     string `json:"Latitude"`
	Longitude    string `json:"Longitude"`
	Name         string `json:"Nome"`
	Registration string `json:"Situacao"`
}

type upstreamCustomer struct {
	upstreamPerson
	Others []upstreamPerson `json:"Outros"`
	Status string           `json:"status"`
}

type UpstreamError struct {
	Operation string
	Status    string
	Code      int
	Err       error
}

func (errorValue *UpstreamError) Error() string {
	if errorValue.Err != nil {
		return fmt.Sprintf("MK %s: %v", errorValue.Operation, errorValue.Err)
	}
	return fmt.Sprintf("MK %s retornou status %q", errorValue.Operation, errorValue.Status)
}

func (errorValue *UpstreamError) Unwrap() error { return errorValue.Err }

type TokenProvider interface {
	Token(context.Context) (string, error)
	Invalidate()
}

type Client struct {
	baseURL       *url.URL
	httpClient    *http.Client
	tokenProvider TokenProvider
}

func NewClient(baseURL *url.URL, httpClient *http.Client, tokenProvider TokenProvider) *Client {
	return &Client{baseURL: baseURL, httpClient: httpClient, tokenProvider: tokenProvider}
}

func (client *Client) CustomerByDocument(ctx context.Context, document string) (CustomerResult, error) {
	for attempt := 0; attempt < 2; attempt++ {
		token, err := client.tokenProvider.Token(ctx)
		if err != nil {
			return CustomerResult{}, err
		}

		result, statusCode, err := client.queryCustomer(ctx, token, document)
		if err == nil {
			return result, nil
		}
		if attempt == 0 && (statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden) {
			client.tokenProvider.Invalidate()
			continue
		}
		return CustomerResult{}, err
	}

	return CustomerResult{}, errors.New("não foi possível consultar o MK")
}

func (client *Client) queryCustomer(ctx context.Context, token, document string) (CustomerResult, int, error) {
	endpoint := endpointURL(client.baseURL, "mk/WSMKConsultaDoc.rule")
	query := endpoint.Query()
	query.Set("sys", "MK0")
	query.Set("token", token)
	query.Set("doc", document)
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return CustomerResult{}, 0, err
	}
	request.Header.Set("Accept", "application/json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return CustomerResult{}, 0, &UpstreamError{Operation: "consulta", Err: err}
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseBody))
		return CustomerResult{}, response.StatusCode, &UpstreamError{Operation: "consulta", Code: response.StatusCode, Status: response.Status}
	}

	var upstream upstreamCustomer
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseBody))
	if err := decoder.Decode(&upstream); err != nil {
		return CustomerResult{}, response.StatusCode, &UpstreamError{Operation: "consulta", Code: response.StatusCode, Err: fmt.Errorf("JSON inválido: %w", err)}
	}
	if upstream.Status != "" && !strings.EqualFold(upstream.Status, "OK") {
		return CustomerResult{}, response.StatusCode, &UpstreamError{Operation: "consulta", Code: response.StatusCode, Status: upstream.Status}
	}

	result := CustomerResult{Customer: mapPerson(upstream.upstreamPerson), Others: make([]Person, 0, len(upstream.Others))}
	for _, person := range upstream.Others {
		result.Others = append(result.Others, mapPerson(person))
	}
	return result, response.StatusCode, nil
}

func mapPerson(person upstreamPerson) Person {
	return Person{
		CEP: person.CEP, PersonCode: person.PersonCode, Email: person.Email,
		Address: person.Address, Phone: person.Phone, Latitude: person.Latitude,
		Longitude: person.Longitude, Name: person.Name, Registration: person.Registration,
	}
}

type StaticTokenProvider struct{ value string }

func NewStaticTokenProvider(value string) *StaticTokenProvider {
	return &StaticTokenProvider{value: value}
}
func (provider *StaticTokenProvider) Token(context.Context) (string, error) {
	return provider.value, nil
}
func (provider *StaticTokenProvider) Invalidate() {}

type AuthenticationTokenProvider struct {
	baseURL         *url.URL
	httpClient      *http.Client
	userAccessToken string
	counterPassword string
	ttl             time.Duration

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func NewAuthenticationTokenProvider(baseURL *url.URL, httpClient *http.Client, userAccessToken, counterPassword string, ttl time.Duration) *AuthenticationTokenProvider {
	return &AuthenticationTokenProvider{
		baseURL: baseURL, httpClient: httpClient, userAccessToken: userAccessToken,
		counterPassword: counterPassword, ttl: ttl,
	}
}

func (provider *AuthenticationTokenProvider) Token(ctx context.Context) (string, error) {
	provider.mu.Lock()
	defer provider.mu.Unlock()

	if provider.token != "" && time.Now().Before(provider.expiresAt) {
		return provider.token, nil
	}

	endpoint := endpointURL(provider.baseURL, "mk/WSAutenticacao.rule")
	query := endpoint.Query()
	query.Set("sys", "MK0")
	query.Set("token", provider.userAccessToken)
	query.Set("password", provider.counterPassword)
	query.Set("cd_servico", "6")
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/json")

	response, err := provider.httpClient.Do(request)
	if err != nil {
		return "", &UpstreamError{Operation: "autenticação", Err: err}
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseBody))
		return "", &UpstreamError{Operation: "autenticação", Code: response.StatusCode, Status: response.Status}
	}

	var payload any
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseBody))
	if err := decoder.Decode(&payload); err != nil {
		return "", &UpstreamError{Operation: "autenticação", Err: fmt.Errorf("JSON inválido: %w", err)}
	}
	token := findAuthenticationToken(payload)
	if token == "" {
		return "", &UpstreamError{Operation: "autenticação", Err: errors.New("token não encontrado na resposta")}
	}

	provider.token = token
	provider.expiresAt = time.Now().Add(provider.ttl)
	return provider.token, nil
}

func (provider *AuthenticationTokenProvider) Invalidate() {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	provider.token = ""
	provider.expiresAt = time.Time{}
}

func findAuthenticationToken(payload any) string {
	object, ok := payload.(map[string]any)
	if !ok {
		return ""
	}

	preferred := []string{"tokenretornoautenticacao", "tokenautenticacao", "token"}
	for _, wanted := range preferred {
		for key, value := range object {
			normalized := strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.ToLower(key))
			if normalized == wanted {
				if token, ok := value.(string); ok {
					return strings.TrimSpace(token)
				}
			}
		}
	}
	for _, value := range object {
		if token := findAuthenticationToken(value); token != "" {
			return token
		}
	}
	return ""
}

func endpointURL(baseURL *url.URL, endpointPath string) *url.URL {
	copyURL := *baseURL
	copyURL.Path = path.Join(copyURL.Path, endpointPath)
	copyURL.RawQuery = ""
	copyURL.Fragment = ""
	return &copyURL
}
