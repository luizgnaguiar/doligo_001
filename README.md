# Doligo ERP/CRM (Go Native Port)

**Status**: MVP (Produção)
**Versão**: 1.0.0
**Arquitetura**: Clean Architecture (Hexagonal)

---

## 1. Visão Geral

O **Doligo** é um sistema ERP/CRM de alta performance, portado do PHP (Dolibarr) para **Go (Golang) Nativo**.
O objetivo deste projeto é fornecer uma fundação robusta, escalável e segura para gestão empresarial, focada em:
- **Performance**: Tempo de resposta de API < 50ms.
- **Concorrência**: Gestão segura de estoques e transações financeiras.
- **Portabilidade**: Binário estático único, sem dependências de sistema operacional.
- **Segurança**: Autenticação JWT Stateless, RBAC granular e Auditoria de Domínio.

---

## 2. Stack Técnica

- **Linguagem**: Go 1.22+ (Strict Standard Library usage)
- **Banco de Dados**: PostgreSQL (Driver `pgx`) / MySQL (Driver `go-sql-driver`)
- **ORM**: GORM v2 (com suporte a SQL Raw para relatórios)
- **HTTP Framework**: Echo v4
- **Containerização**: Docker Scratch (Multi-stage Build)
- **Logs**: ZeroLog (Estruturado JSON)
- **Documentos**: Maroto (Gerador PDF Pure Go)

---

## 3. Destaques de Engenharia

O projeto implementa rigorosamente critérios de engenharia de software para sistemas críticos:

### CT-01: Portabilidade Extrema
O binário final é **100% estático** (`CGO_ENABLED=0`), capaz de rodar em containers `scratch` de ~20MB, eliminando vulnerabilidades de SO base.

### CT-02: Concorrência Pessimista (Estoque)
Utiliza `SELECT ... FOR UPDATE` (Pessimistic Locking) para garantir integridade absoluta em movimentações de estoque simultâneas.
*Validado por testes de estresse em `internal/infrastructure/repository/stock_repository_ct02_test.go`.*

### CT-03: Cancelamento Assíncrono
Suporte nativo a `context.Context` em todas as camadas. Se um cliente cancela a requisição, o processamento (ex: geração de PDF pesado) é interrompido imediatamente para poupar recursos.

### CT-04: Graceful Shutdown
O sistema intercepta sinais `SIGTERM/SIGINT` e aguarda o término de tarefas críticas (workers de email, transações) antes de encerrar, garantindo zero perda de dados.

---

## 4. Quick Start

### Pré-requisitos
- Docker & Docker Compose
- Go 1.22+ (para desenvolvimento local)

### Rodando com Docker (Recomendado)

1. **Configurar Ambiente**
   ```bash
   cp .env.example .env
   # Ajuste as variáveis se necessário
   ```

2. **Build & Run**
   ```bash
   docker-compose up --build -d
   ```

3. **Acessar**
   - API: `http://localhost:8080`
   - Health Check: `http://localhost:8080/health`

### Compilando Localmente

```bash
# Baixar dependências
go mod download

# Rodar Migrations (Embutidas no binário, mas executadas na inicialização)
go run cmd/main.go migrate

# Iniciar Servidor
go run cmd/main.go server
```

---

## 5. Mapa da Documentação

Toda a documentação técnica detalhada encontra-se na pasta `docs/`:

| Categoria | Arquivo | Descrição |
|-----------|---------|-----------|
| **Arquitetura** | [database_schema.md](docs/database_schema.md) | Diagrama e detalhes do esquema SQL. |
| | [concurrency_control.md](docs/concurrency_control.md) | Estratégias de locking e transações. |
| | [env_vars.md](docs/env_vars.md) | Dicionário de variáveis de ambiente. |
| **Operação** | [runbook_operacional.md](docs/runbook_operacional.md) | Guia dia-a-dia para SRE/Ops. |
| | [backup_strategy.md](docs/backup_strategy.md) | Políticas de RPO/RTO e backup. |
| | [log_rotation.md](docs/log_rotation.md) | Configuração de logs e rotação. |
| **Segurança** | [audit_logs.md](docs/audit_logs.md) | Estrutura de logs de auditoria. |
| | [security_headers.md](docs/security_headers.md) | Headers HTTP de segurança implementados. |
| **Qualidade** | [technical_debt.md](docs/technical_debt.md) | **IMPORTANTE**: Lista de pendências e dívidas técnicas. |
| | [observability.md](docs/observability.md) | Métricas e monitoramento. |

---

## 6. Dívida Técnica Crítica (Atenção)

Consulte `docs/technical_debt.md` para a lista completa. Destacamos:

- **Custo Médio (CMP) em Reversões**: Devido à ausência de histórico de custo unitário na tabela `stock_movements`, reversões de saída (ex: devolução de venda) utilizam o **Custo Médio Atual** do item como base para a reentrada. Isso garante neutralidade matemática mas pode divergir do custo histórico real.
- **Validação de UUID**: Verificações de nulos foram implementadas em repositórios críticos (Role, Permission), mas recomenda-se expansão para todas as entidades.

---

**Licença**: Proprietária / Interna.
**Maintainer**: Equipe de Arquitetura Doligo.