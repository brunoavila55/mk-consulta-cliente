package mk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCustomerByDocument(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/mk/WSMKConsultaDoc.rule" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if got := request.URL.Query().Get("sys"); got != "MK0" {
			t.Errorf("sys = %q", got)
		}
		if got := request.URL.Query().Get("token"); got != "secret-token" {
			t.Errorf("token = %q", got)
		}
		if got := request.URL.Query().Get("doc"); got != "52998224725" {
			t.Errorf("doc = %q", got)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{
			"CEP":"44700000", "CodigoPessoa":13565, "Email":"a@example.com",
			"Endereco":"Rua A", "Fone":"5511999999999", "Nome":"Cliente",
			"Outros":[{"CodigoPessoa":13726,"Nome":"Cliente","Situacao":"Ativo"}],
			"Situacao":"Ativo", "status":"OK"
		}`))
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := NewClient(baseURL, server.Client(), NewStaticTokenProvider("secret-token"))
	result, err := client.CustomerByDocument(context.Background(), "52998224725")
	if err != nil {
		t.Fatalf("CustomerByDocument() error = %v", err)
	}
	if result.Customer.PersonCode != 13565 || result.Customer.Name != "Cliente" {
		t.Errorf("customer = %+v", result.Customer)
	}
	if len(result.Others) != 1 || result.Others[0].PersonCode != 13726 {
		t.Errorf("others = %+v", result.Others)
	}
}

func TestFindAuthenticationToken(t *testing.T) {
	t.Parallel()
	payload := map[string]any{"data": map[string]any{"tokenRetornoAutenticacao": " generated-token "}}
	if got := findAuthenticationToken(payload); got != "generated-token" {
		t.Errorf("findAuthenticationToken() = %q", got)
	}
}
