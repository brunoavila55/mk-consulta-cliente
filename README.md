# consulta-cliente

API em Go que consulta cadastro de cliente na MK Solutions, usada pelo chatbot. Porta em Docker, exposta em `api.newlifefibra.com.br/consulta-cliente`.

## Deploy na VM (Traefik)

Assume que a VM já tem o stack do Traefik rodando na rede externa `proxy` (veja instruções à parte). Se ainda não tiver, crie a rede antes:

```bash
docker network create proxy
```

Depois:

```bash
git clone https://github.com/brunoavila55/mk-consulta-cliente.git
cd mk-consulta-cliente
cp .env.example .env
# edite o .env com as credenciais reais (ver abaixo)
docker compose up -d --build
```

## Variáveis de ambiente (`.env`)

| Variável | Descrição |
|---|---|
| `MK_CONTRA_SENHA` | Contra-senha (parâmetro `token`) da autenticação MK Solutions |
| `MK_PASSWORD` | Senha da autenticação MK Solutions |
| `API_KEY` | Chave que o chatbot deve enviar no parâmetro `key` |
| `MK_BASE_URL` | Base da API MK Solutions (padrão `https://sac.newlifefibra.com.br`) |
| `MK_SYS`, `MK_CD_SERVICO` | Parâmetros fixos da MK (padrão `MK0` / `9999`) |
| `MK_TOKEN_TTL` | Tempo de cache do token de acesso MK (padrão `20m`) |
| `MK_RETRY_ATTEMPTS` | Tentativas em falha transitória ao chamar a MK (padrão `2`) |
| `HTTP_TIMEOUT` | Timeout por chamada HTTP à MK (padrão `15s`) |
| `REQUEST_TIMEOUT` | Timeout total de uma requisição recebida (padrão `20s`) |
| `SHUTDOWN_TIMEOUT` | Tempo de graceful shutdown ao receber SIGTERM (padrão `15s`) |
| `PORT` | Porta interna do servidor (padrão `8080`) |

`MK_CONTRA_SENHA`, `MK_PASSWORD` e `API_KEY` são obrigatórias — o processo não sobe sem elas.

## Endpoint

```
GET /consulta-cliente?doc={cpf-ou-cnpj}&key={chave-do-chatbot}
```

`/healthz` responde `200 ok` para health check do orquestrador.

## Métricas (Prometheus)

`/metrics` expõe métricas no formato Prometheus. Ele **não** é roteado publicamente pelo Traefik (não há router para ele nas labels do `docker-compose.yml`) — só é alcançável de dentro da rede Docker. Aponte o Prometheus para `consulta-cliente:8080/metrics` na rede `proxy` (ou coloque o Prometheus nessa mesma rede), por exemplo:

```yaml
scrape_configs:
  - job_name: consulta-cliente
    static_configs:
      - targets: ["consulta-cliente:8080"]
```

Métricas expostas:

| Métrica | Tipo | Descrição |
|---|---|---|
| `http_requests_total{method,path,status}` | Counter | Requisições recebidas pela API |
| `http_request_duration_seconds{method,path,status}` | Histogram | Latência das requisições recebidas |
| `http_requests_in_flight` | Gauge | Requisições sendo processadas agora |
| `mk_requests_total{endpoint,outcome}` | Counter | Chamadas feitas à API da MK Solutions (sucesso/erro) |
| `mk_request_duration_seconds{endpoint}` | Histogram | Latência das chamadas à MK Solutions |
| `mk_request_retries_total{endpoint}` | Counter | Retentativas feitas em chamadas à MK Solutions |
| `mk_token_cache_total{result}` | Counter | Hits/misses do cache do token de acesso MK |
| `mk_token_refresh_total{outcome}` | Counter | Autenticações contra `WSAutenticacao.rule` |
| `consulta_cliente_resultado_total{tipo}` | Counter | Resultado de negócio de cada consulta (`Lead`/`Cliente`) |

Métricas padrão de runtime Go e processo (`go_*`, `process_*`) também são expostas automaticamente pelo `client_golang`.

## Desenvolvimento local

```bash
go build ./...
go vet ./...
```
