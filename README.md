# Serviço de Empresas (Go + MongoDB + RabbitMQ)

API RESTful para cadastro e gestão de empresas. Persistência em MongoDB e publicação de eventos em RabbitMQ.

- Linguagem: Go 1.25
- SDK/Dependências: Go Modules
- Armazenamento: MongoDB
- Mensageria: RabbitMQ
- Status das mensagens: "Cadastro/Edição/Exclusão da EMPRESA <nome_fantasia>"

## Sumário
- [Arquitetura e Decisões](#arquitetura-e-decisões)
- [Requisitos](#requisitos)
- [Configuração (Variáveis de Ambiente)](#configuração-variáveis-de-ambiente)
- [Rodando com Docker](#rodando-com-docker)
- [Execução Local](#execução-local)
- [Endpoints](#endpoints)
- [WebSocket de Eventos](#websocket-de-eventos)
- [Exemplos de Requisição](#exemplos-de-requisição)
- [Modelo de Erros](#modelo-de-erros)
- [Observabilidade](#observabilidade)
- [Testes](#testes)
- [Troubleshooting](#troubleshooting)

## Arquitetura e Decisões
- API HTTP simples usando chi.
- Unicidade de CNPJ garantida por índice único no MongoDB.
- Publicação de eventos em RabbitMQ após operações CRUD (quando configurado).

## Requisitos
- Docker e Docker Compose (para rodar com containers)
- OU Go 1.25, MongoDB e RabbitMQ disponíveis localmente

## Configuração (Variáveis de Ambiente)
Crie um arquivo `.env` na raiz (opcional). Exemplos de variáveis comuns:

- HTTP_ADDR: endereço do servidor HTTP (ex.: :8080)
- MONGO_URI: URI de conexão (ex.: mongodb://localhost:27017)
- MONGO_DB: nome do banco (ex.: empresa)
- MONGO_COLLECTION: nome da coleção (ex.: empresas)
- RABBITMQ_URL: URL do RabbitMQ (ex.: amqp://guest:guest@localhost:5672/)
- API_BASE_PATH: base path da API (ex.: /api) — mantenha consistente com as rotas

Observação: Garanta consistência entre o base path documentado (ex.: /api/empresas) e o configurado na aplicação.

## Rodando com Docker
1) docker compose up --build
2) API disponível em http://localhost:8080/api/empresas (ajuste se não usar base path /api)
3) WebSocket de eventos em ws://localhost:8090/ws/events
4) MongoDB: localhost:27017
5) RabbitMQ Management: http://localhost:15672 (guest/guest)

## Execução Local
- API: go run ./cmd/server
- WS:  go run ./cmd/wsserver
- Defina as variáveis de ambiente (use `.env` opcionalmente)

## Endpoints
Base Path: /api (ajuste se diferente)

- POST   /api/empresas — cria uma empresa
  - Content-Type: application/x-www-form-urlencoded ou multipart/form-data
  - Body (campos): cnpj, nome_fantasia, razao_social, endereco, num_funcionarios, num_min_pcd
  - Respostas:
    - 201 Created: {"id": "<novo_id>"}
    - 400 Bad Request: {"error": "<mensagem>"}
- GET    /api/empresas — lista empresas
  - Respostas:
    - 200 OK: [ { empresa }, ... ]
    - 500 Internal Server Error: {"error": "<mensagem>"}
- GET    /api/empresas/{id} — obtém empresa por ID
  - Respostas:
    - 200 OK: { empresa }
    - 404 Not Found: {"error": "não encontrado"}
- PUT    /api/empresas/{id} — atualiza empresa
  - Content-Type: application/x-www-form-urlencoded ou multipart/form-data
  - Body (campos): cnpj, nome_fantasia, razao_social, endereco, num_funcionarios, num_min_pcd
  - Respostas:
    - 200 OK: {"status": "ok"}
    - 400 Bad Request: {"error": "<mensagem>"}
- DELETE /api/empresas/{id} — remove empresa
  - Respostas:
    - 200 OK: {"status": "ok"}
    - 500 Internal Server Error: {"error": "<mensagem>"}

Campos da Empresa (modelo):
- id (string, somente resposta)
- cnpj (string, único, obrigatório)
- nome_fantasia (string)
- razao_social (string)
- endereco (string)
- num_funcionarios (int)
- num_min_pcd (int)

## Exemplos de Requisição

Criar empresa (x-www-form-urlencoded):

curl -X POST http://localhost:8080/api/empresas \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data-urlencode "cnpj=12345678000195" \
  --data-urlencode "nome_fantasia=Acme" \
  --data-urlencode "razao_social=Acme LTDA" \
  --data-urlencode "endereco=Rua X, 123" \
  --data-urlencode "num_funcionarios=50" \
  --data-urlencode "num_min_pcd=5"

Atualizar empresa:

curl -X PUT http://localhost:8080/api/empresas/{id} \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data-urlencode "cnpj=12345678000195" \
  --data-urlencode "nome_fantasia=Acme 2" \
  --data-urlencode "razao_social=Acme LTDA" \
  --data-urlencode "endereco=Rua Y, 999" \
  --data-urlencode "num_funcionarios=80" \
  --data-urlencode "num_min_pcd=8"

Listar empresas:

curl -X GET http://localhost:8080/api/empresas

Obter por ID:

curl -X GET http://localhost:8080/api/empresas/{id}

Remover:

curl -X DELETE http://localhost:8080/api/empresas/{id}

## Modelo de Erros
Todas as respostas de erro seguem o formato JSON:

{"error": "<mensagem>"}

Principais códigos:
- 400: Erros de validação/negócio (ex.: CNPJ inválido, CNPJ já cadastrado)
- 404: Recurso não encontrado
- 500: Erro interno (falhas de repositório/infra)

## Observabilidade
- Logs: o serviço escreve logs padrão na saída do processo.
- RabbitMQ: quando configurado, publica mensagens como "Cadastro/Edição/Exclusão da EMPRESA <nome_fantasia>".
- MongoDB: dados persistidos na coleção configurada (ex.: empresas).

## WebSocket de Eventos
- Serviço dedicado que consome a fila RabbitMQ (padrão: logs.empresas) e transmite cada mensagem para todos os clientes conectados.
- Endpoint: ws://localhost:8090/ws/events
- Protocolo: WebSocket (texto). Cada evento do RabbitMQ é enviado como uma mensagem de texto ao cliente.
- Variáveis de ambiente (WS):
  - WS_HTTP_ADDR (default :8090)
  - RABBITMQ_URL (default amqp://guest:guest@rabbitmq:5672/)
  - RABBITMQ_QUEUE (default logs.empresas)
- Teste rápido:
  - Conecte com: `wscat -c ws://localhost:8090/ws/events` ou use qualquer cliente WebSocket.
  - Realize operações na API (criar/editar/excluir empresa) e observe as mensagens chegarem em tempo real.

## Testes
- Rode os testes: go test ./...
- Testes principais em: internal/httpapi/handlers_test.go

## Troubleshooting
- Conexão MongoDB falhando: verifique MONGO_URI e se o Mongo está acessível.
- CNPJ único: se houver erro de duplicidade, remova o documento duplicado ou ajuste seu dado de teste.
- RabbitMQ indisponível: as operações CRUD funcionam mesmo sem publisher; apenas a publicação de eventos não ocorrerá.
