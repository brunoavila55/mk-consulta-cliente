# consulta-cliente

API em Go que consulta cadastro de cliente na MK Solutions, usada pelo chatbot. Porta em Docker, exposta em `api.newlifefibra.com.br/consulta-cliente`.

## Deploy na VM (Traefik)

Assume que a VM já tem o stack do Traefik rodando na rede externa `traefik-public` (veja instruções à parte). Se ainda não tiver, crie a rede antes:

```bash
docker network create traefik-public
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

## Desenvolvimento local

```bash
go build ./...
go vet ./...
```
