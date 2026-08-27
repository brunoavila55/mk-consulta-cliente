# MK Consulta Cliente

API HTTP escrita em Go para receber o CPF informado em uma automação de chatbot, consultar o cadastro correspondente no MK Solutions ERP e devolver os dados em um JSON previsível.

Esta é uma implementação completamente nova. Ela não preserva rotas, variáveis, respostas ou comportamentos de versões anteriores deste repositório.

## O que a API faz

1. Recebe um CPF pelo endpoint `GET /v1/clientes`.
2. Remove a pontuação e valida os dígitos verificadores do CPF.
3. Obtém um token temporário no `WSAutenticacao.rule`, usando:
   - o token fixo do usuário autorizado;
   - a contrassenha gerada na criação do perfil de Webservice;
   - o código de serviço `6` (`Consulta documento`).
4. Consulta o CPF no `WSMKConsultaDoc.rule`.
5. Normaliza a resposta do MK e a devolve ao chatbot.

```text
Chatbot
   |
   | GET /v1/clientes?cpf=...
   | X-API-Key: ...
   v
API Go
   |
   | token fixo do usuário + contrassenha do Webservice
   v
WSAutenticacao.rule (serviço 6)
   |
   | token temporário
   v
WSMKConsultaDoc.rule
   |
   | dados do cliente
   v
Resposta normalizada para o chatbot
```

## Escopo atual

Incluído:

- consulta por CPF;
- validação completa do CPF;
- autenticação automática no MK;
- cache em memória do token temporário;
- suporte ao cadastro principal e à lista `Outros` retornada pelo MK;
- proteção do endpoint por chave de API;
- timeouts e erros JSON padronizados;
- logs estruturados sem CPF ou credenciais;
- graceful shutdown;
- execução local ou em Docker;
- healthcheck do container.

Fora do escopo desta primeira versão:

- CNPJ;
- consulta de contratos, conexões, faturas ou boletos;
- classificação entre lead e cliente;
- banco de dados ou armazenamento de consultas;
- painel administrativo;
- proxy reverso e certificado HTTPS da API pública;
- compatibilidade com a versão antiga do repositório.

## Integração utilizada no MK

Documentação de referência: [APIs gerais do MK Solutions](https://mkloud.atlassian.net/wiki/spaces/MK30/pages/48699908/APIs%2Bgerais).

### 1. Autenticação

```text
GET {MK_BASE_URL}/mk/WSAutenticacao.rule
    ?sys=MK0
    &token={token fixo do usuário}
    &password={contrassenha do perfil de Webservice}
    &cd_servico=6
```

Nomenclatura adotada no projeto:

| Conceito no MK | Variável | Duração | Uso |
|---|---|---:|---|
| Token de acesso do usuário autorizado | `MK_USER_ACCESS_TOKEN` | Fixo | Enviado como parâmetro `token` ao `WSAutenticacao.rule` |
| Contrassenha do perfil de Webservice | `MK_WEBSERVICE_COUNTER_PASSWORD` | Fixa até ser alterada no MK | Enviada como parâmetro `password` |
| Token retornado pela autenticação | Gerenciado internamente | Temporário | Enviado ao `WSMKConsultaDoc.rule` |

O perfil de Webservice precisa estar ativo, permitir o usuário escolhido e liberar o serviço **Consulta documento**, código `6`. Se o perfil limitar o consumo por IP, o IP de saída do servidor da API também precisa ser autorizado no MK.

### 2. Consulta do documento

```text
GET {MK_BASE_URL}/mk/WSMKConsultaDoc.rule
    ?sys=MK0
    &token={token temporário}
    &doc={CPF sem pontuação}
```

## Configuração

Todas as configurações são recebidas por variáveis de ambiente. O arquivo [.env.example](.env.example) serve como modelo e não contém credenciais reais.

| Variável | Obrigatória | Padrão | Descrição |
|---|:---:|---|---|
| `HTTP_PORT` | Não | `8080` | Porta interna da API |
| `CHATBOT_API_KEY` | Sim | - | Chave com no mínimo 24 caracteres exigida no header `X-API-Key` |
| `MK_BASE_URL` | Sim | - | Endereço base do MK, atualmente `http://177.72.80.20:8080` |
| `MK_ALLOW_INSECURE_HTTP` | Sim para esse MK | `false` | Deve ser `true` para confirmar conscientemente o uso de HTTP sem TLS |
| `MK_USER_ACCESS_TOKEN` | Sim no fluxo normal | - | Token fixo do usuário liberado no perfil de Webservice |
| `MK_WEBSERVICE_COUNTER_PASSWORD` | Sim no fluxo normal | - | Contrassenha criada junto com o perfil de Webservice |
| `MK_TEMPORARY_AUTH_TOKEN_TTL` | Não | `5m` | Tempo de cache local do token temporário |
| `MK_TEMPORARY_AUTH_TOKEN` | Não | - | Alternativa de diagnóstico; ignora a autenticação automática enquanto estiver preenchida |
| `MK_HTTP_TIMEOUT` | Não | `10s` | Timeout de cada chamada HTTP ao MK |
| `API_BIND_ADDRESS` | Não | `127.0.0.1` | Endereço no qual o Docker publica a porta |
| `API_PUBLIC_PORT` | Não | `8080` | Porta publicada pelo Docker no host |
| `API_ENV_FILE` | Não | `.env` | Arquivo de ambiente usado pelo Compose |

### Exemplo de `.env`

```dotenv
HTTP_PORT=8080
CHATBOT_API_KEY=substitua-por-uma-chave-longa-e-aleatoria

MK_BASE_URL=http://177.72.80.20:8080
MK_ALLOW_INSECURE_HTTP=true
MK_USER_ACCESS_TOKEN=preencher-no-servidor
MK_WEBSERVICE_COUNTER_PASSWORD=preencher-no-servidor
MK_TEMPORARY_AUTH_TOKEN_TTL=5m
MK_HTTP_TIMEOUT=10s

API_BIND_ADDRESS=127.0.0.1
API_PUBLIC_PORT=8080
```

Nunca envie o `.env` ao Git. Ele está listado no `.gitignore` e no `.dockerignore`.

## Executar com Docker

### Pré-requisitos

- Docker Engine ou Docker Desktop;
- Docker Compose;
- conectividade do servidor até `177.72.80.20:8080`;
- credenciais do perfil de Webservice do MK.

### Primeira execução

No Linux:

```bash
cp .env.example .env
nano .env
docker compose up -d --build
```

No PowerShell:

```powershell
Copy-Item .env.example .env
notepad .env
docker compose up -d --build
```

Consulte o estado:

```bash
docker compose ps
docker compose logs --tail=100 api
```

O estado deve mudar para `healthy` depois que `GET /health` começar a responder.

### Atualização

```bash
git pull --ff-only
docker compose up -d --build
docker image prune -f
```

### Reinício e encerramento

```bash
docker compose restart api
docker compose down
```

O Compose usa `restart: unless-stopped`, sistema de arquivos somente leitura, nenhuma capability Linux adicional, limite de processos e rotação dos logs locais.

### Publicação da API

Por padrão, o Compose publica a API apenas em:

```text
127.0.0.1:8080
```

Essa configuração é apropriada para instalar um proxy reverso HTTPS na mesma máquina. O proxy deve encaminhar o tráfego público para `http://127.0.0.1:8080`.

Não é recomendado publicar a API diretamente em `0.0.0.0` sem HTTPS. Caso isso seja necessário temporariamente para diagnóstico, altere `API_BIND_ADDRESS`, restrinja a porta no firewall e mantenha `X-API-Key` habilitado.

## Executar sem Docker

O binário lê variáveis do ambiente; ele não carrega `.env` automaticamente.

Exemplo no PowerShell:

```powershell
$env:HTTP_PORT = "8080"
$env:CHATBOT_API_KEY = "uma-chave-longa-com-pelo-menos-24-caracteres"
$env:MK_BASE_URL = "http://177.72.80.20:8080"
$env:MK_ALLOW_INSECURE_HTTP = "true"
$env:MK_USER_ACCESS_TOKEN = "token-fixo-do-usuario"
$env:MK_WEBSERVICE_COUNTER_PASSWORD = "contrassenha-do-webservice"
go run ./cmd/api
```

## Contrato HTTP

### `GET /health`

Healthcheck público, sem autenticação e sem chamada ao MK.

Resposta `200 OK`:

```json
{
  "status": "ok"
}
```

Esse endpoint confirma que o processo está respondendo. Ele não confirma que as credenciais ou o MK estão funcionando.

### `GET /v1/clientes`

Consulta um CPF no MK.

Parâmetros:

| Local | Nome | Obrigatório | Exemplo |
|---|---|:---:|---|
| Query string | `cpf` | Sim | `529.982.247-25` |
| Header | `X-API-Key` | Sim | valor de `CHATBOT_API_KEY` |

O CPF pode ser enviado com ou sem pontos e traço. Letras, quantidade incorreta de dígitos, CPFs repetidos como `11111111111` e dígitos verificadores incorretos são rejeitados antes de qualquer chamada ao MK.

Exemplo com `curl`:

```bash
curl --get 'https://SEU-DOMINIO/v1/clientes' \
  --data-urlencode 'cpf=529.982.247-25' \
  --header 'X-API-Key: SUA_CHAVE_DO_CHATBOT'
```

Exemplo no PowerShell:

```powershell
$headers = @{ "X-API-Key" = "SUA_CHAVE_DO_CHATBOT" }
Invoke-RestMethod `
  -Uri "https://SEU-DOMINIO/v1/clientes?cpf=529.982.247-25" `
  -Headers $headers
```

Resposta `200 OK`:

```json
{
  "status": "ok",
  "dados": {
    "cliente": {
      "cep": "44700000",
      "codigo_pessoa": 13565,
      "email": "cliente@exemplo.com.br",
      "endereco": "Rua A, 0 - Centro",
      "fone": "5511999999999",
      "latitude": "",
      "longitude": "",
      "nome": "Cliente Exemplo",
      "situacao": "Ativo"
    },
    "outros": [
      {
        "cep": "96930000",
        "codigo_pessoa": 13778,
        "email": "",
        "endereco": "Rua B, 0 - Centro",
        "fone": "5511999999999",
        "latitude": "",
        "longitude": "",
        "nome": "Cliente Exemplo",
        "situacao": "Ativo"
      }
    ]
  }
}
```

Campos úteis para mapear no chatbot:

| Informação | Caminho JSON |
|---|---|
| Nome principal | `dados.cliente.nome` |
| Código no MK | `dados.cliente.codigo_pessoa` |
| Situação | `dados.cliente.situacao` |
| Telefone | `dados.cliente.fone` |
| E-mail | `dados.cliente.email` |
| Endereço | `dados.cliente.endereco` |
| Outros cadastros | `dados.outros` |

### Erros

Todos os erros produzidos pela API seguem o mesmo formato:

```json
{
  "status": "erro",
  "erro": {
    "codigo": "cpf_invalido",
    "mensagem": "Informe um CPF válido com 11 dígitos."
  }
}
```

| HTTP | Código | Significado |
|---:|---|---|
| `400` | `cpf_invalido` | CPF ausente ou inválido |
| `401` | `nao_autorizado` | Header `X-API-Key` ausente ou incorreto |
| `502` | `mk_indisponivel` | O MK recusou a autenticação, retornou erro ou uma resposta inválida |
| `504` | `mk_timeout` | A chamada ao MK ultrapassou o timeout |

Detalhes internos e credenciais não são devolvidos ao consumidor.

## Configuração no chatbot

No bloco **Conecte a outro sistema**:

| Campo | Valor |
|---|---|
| Método | `GET` |
| URL | Domínio HTTPS público da API, sem path e sem credenciais |
| Path | `/v1/clientes` |
| Params | Chave `cpf`, com o valor da variável coletada no fluxo |
| Headers | Chave `X-API-Key`, com a chave configurada no servidor |

A sintaxe exata da variável do CPF depende da plataforma do chatbot. Exemplos comuns são `{{cpf}}`, `#cpf` ou a seleção da variável pela interface.

O token fixo do MK e a contrassenha nunca devem ser cadastrados no chatbot. O chatbot conhece somente a chave desta API.

## Segurança e privacidade

### Limitação conhecida do MK

O MK informado está disponível apenas em:

```text
http://177.72.80.20:8080
```

Esse trecho não possui TLS. Por isso, CPF, credenciais do Webservice e dados retornados pelo MK trafegam sem criptografia entre o servidor da API e o ERP. `MK_ALLOW_INSECURE_HTTP=true` existe para tornar essa decisão explícita e evitar ativação acidental em outro ambiente.

Recomendações:

- hospedar a API em infraestrutura confiável;
- restringir o acesso ao MK pelo IP de saída do servidor, se o perfil permitir;
- usar VPN ou túnel privado quando possível;
- sempre oferecer HTTPS entre o chatbot e esta API;
- manter `.env`, backups e logs acessíveis somente a administradores;
- trocar imediatamente qualquer credencial que tenha sido exposta.

### Proteções implementadas

- CPF não aparece nos logs porque a query string inteira é omitida;
- tokens, contrassenha e chave do chatbot não são registrados;
- a chave do chatbot é comparada em tempo constante;
- respostas do MK têm limite de tamanho;
- chamadas externas possuem timeout;
- o container executa sem root e sem capabilities Linux;
- o sistema de arquivos do container é somente leitura.

## Desenvolvimento e verificação

Requer Go 1.26 ou superior.

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
```

Validar o Compose sem usar credenciais reais:

No PowerShell:

```powershell
$env:API_ENV_FILE = ".env.example"
docker compose config --quiet
Remove-Item Env:API_ENV_FILE
```

No Linux:

```bash
API_ENV_FILE=.env.example docker compose config --quiet
```

## Estrutura do projeto

```text
.
|-- cmd/
|   |-- api/                 # inicialização e ciclo de vida do servidor
|   `-- healthcheck/         # healthcheck usado pela imagem Docker
|-- internal/
|   |-- config/              # leitura e validação do ambiente
|   |-- document/            # normalização e validação do CPF
|   |-- httpapi/             # rotas, autenticação e resposta HTTP
|   `-- mk/                  # autenticação e cliente do MK Solutions
|-- .env.example             # modelo sem segredos
|-- Dockerfile               # build multi-stage e runtime sem root
|-- compose.yaml             # execução e proteções do container
|-- context.md               # contexto técnico para manutenção futura
`-- README.md
```

## Diagnóstico

### A API encerra imediatamente

Consulte:

```bash
docker compose logs api
```

O processo rejeita configurações ausentes, chave do chatbot curta, URL inválida e HTTP sem a confirmação explícita.

### `401 nao_autorizado`

- confira se o header se chama exatamente `X-API-Key`;
- confira se o valor é o mesmo de `CHATBOT_API_KEY`;
- não coloque a chave no parâmetro da URL.

### `400 cpf_invalido`

- confirme se o chatbot está enviando a variável correta;
- confirme se o valor tem 11 dígitos;
- pontos e traço são aceitos, mas letras não.

### `502 mk_indisponivel`

- valide a conectividade com `177.72.80.20:8080` a partir do servidor;
- confirme que `MK_USER_ACCESS_TOKEN` é o token fixo do usuário, não a contrassenha;
- confirme que `MK_WEBSERVICE_COUNTER_PASSWORD` é a contrassenha do perfil;
- libere o serviço `6` no perfil de Webservice;
- verifique a restrição de IPs no MK;
- ajuste `MK_TEMPORARY_AUTH_TOKEN_TTL` para um valor inferior ao prazo real do token.

### Container `unhealthy`

```bash
docker inspect --format='{{json .State.Health}}' mk-consulta-cliente-api-1
docker compose logs --tail=100 api
```

O healthcheck usa `HTTP_PORT=8080` dentro do container. O Compose mantém essa porta interna fixa mesmo quando `API_PUBLIC_PORT` é alterada.

## Checklist de produção

- [ ] Criar perfil de Webservice no MK.
- [ ] Liberar o serviço `6` (`Consulta documento`).
- [ ] Liberar o usuário que possui o token fixo.
- [ ] Autorizar o IP de saída da API, se houver restrição.
- [ ] Criar `.env` diretamente no servidor.
- [ ] Gerar `CHATBOT_API_KEY` aleatória com pelo menos 24 caracteres.
- [ ] Configurar token fixo e contrassenha sem colocá-los no Git.
- [ ] Subir o Compose e confirmar o estado `healthy`.
- [ ] Publicar a API por HTTPS.
- [ ] Configurar URL, path, parâmetro e header no chatbot.
- [ ] Testar CPF válido, CPF inválido, chave incorreta e indisponibilidade do MK.
