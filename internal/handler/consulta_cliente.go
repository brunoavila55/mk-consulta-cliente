package handler

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"consulta-cliente/internal/mk"
)

// maxOutrosIterados e maxCadastros replicam os limites da implementação
// original: no máximo 3 cadastros com conexão ativa, olhando no máximo os
// primeiros 11 registros de "Outros" (índices 0 a 10).
const (
	maxCadastros      = 3
	maxOutrosIterados = 10
)

// docSanitizer mantém apenas dígitos, cobrindo CPF (pontos e traço) e CNPJ
// (pontos, barra e traço), além de qualquer erro de digitação do cliente.
var docSanitizer = regexp.MustCompile(`\D`)

type cadastroResposta struct {
	CodCliente mk.FlexibleID `json:"codCliente"`
	Nome       string        `json:"nome"`
	Endereco   string        `json:"endereco"`
}

type consultaClienteResposta struct {
	Tipo      string             `json:"tipo"`
	Cadastros []cadastroResposta `json:"cadastros"`
}

func placeholders() []cadastroResposta {
	return []cadastroResposta{
		{CodCliente: mk.EmptyID()},
		{CodCliente: mk.EmptyID()},
		{CodCliente: mk.EmptyID()},
	}
}

// ConsultaCliente expõe GET /consulta-cliente.
type ConsultaCliente struct {
	MK             *mk.Client
	Tokens         *mk.TokenManager
	APIKey         string
	RequestTimeout time.Duration
	Logger         *slog.Logger
}

func (h *ConsultaCliente) logger() *slog.Logger {
	if h.Logger != nil {
		return h.Logger
	}
	return slog.Default()
}

func (h *ConsultaCliente) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.RequestTimeout)
	defer cancel()

	q := r.URL.Query()
	key := q.Get("key")
	doc := q.Get("doc")

	if key == "" || subtle.ConstantTimeCompare([]byte(key), []byte(h.APIKey)) != 1 {
		writeError(w, http.StatusUnauthorized, "Consulta não autorizada, parâmetro key incorreto!")
		return
	}
	if doc == "" {
		writeError(w, http.StatusBadRequest, "Parâmetro doc é obrigatório")
		return
	}
	docTratado := docSanitizer.ReplaceAllString(doc, "")

	token, err := h.Tokens.Get(ctx)
	if err != nil {
		h.logger().ErrorContext(ctx, "falha ao obter token MK", "error", err)
		writeError(w, http.StatusBadGateway, "Falha ao autenticar com o sistema MK")
		return
	}

	cadastro, err := h.MK.ConsultaDoc(ctx, token, docTratado)
	if err != nil {
		h.logger().ErrorContext(ctx, "falha ao consultar documento", "error", err)
		writeError(w, http.StatusBadGateway, "Falha ao consultar documento")
		return
	}

	resposta := consultaClienteResposta{Tipo: "Lead", Cadastros: []cadastroResposta{}}

	switch cadastro.Status {
	case "OK":
		if cadastro.Situacao == "Ativo" {
			conexoes, err := h.MK.ConexoesPorCliente(ctx, token, cadastro.CodigoPessoa.String())
			if err != nil {
				h.logger().ErrorContext(ctx, "falha ao consultar conexões", "error", err)
				writeError(w, http.StatusBadGateway, "Falha ao consultar conexões")
				return
			}
			if conexoes.Status == "OK" && len(conexoes.Conexoes) > 0 {
				resposta.Tipo = "Cliente"
				resposta.Cadastros = append(resposta.Cadastros, cadastroResposta{
					CodCliente: cadastro.CodigoPessoa,
					Nome:       cadastro.Nome,
					Endereco:   cadastro.Endereco,
				})
			}
		}

		for i, outro := range cadastro.Outros {
			conexoes, err := h.MK.ConexoesPorCliente(ctx, token, outro.CodigoPessoa.String())
			if err != nil {
				h.logger().ErrorContext(ctx, "falha ao consultar conexões de outro cadastro", "error", err)
				writeError(w, http.StatusBadGateway, "Falha ao consultar conexões")
				return
			}
			if conexoes.Status == "OK" && len(conexoes.Conexoes) > 0 {
				resposta.Tipo = "Cliente"
				resposta.Cadastros = append(resposta.Cadastros, cadastroResposta{
					CodCliente: outro.CodigoPessoa,
					Nome:       outro.Nome,
					Endereco:   outro.Endereco,
				})
			}
			if len(resposta.Cadastros) == maxCadastros || i == maxOutrosIterados {
				break
			}
		}
	case "ERRO":
		// status ERRO da MK indica documento não cadastrado: segue como Lead vazio.
	default:
		writeError(w, http.StatusUnauthorized, "Consulta não autorizada")
		return
	}

	resposta.Cadastros = append(resposta.Cadastros, placeholders()...)

	writeJSON(w, http.StatusOK, resposta)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("erro ao serializar resposta", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"message": message})
}
