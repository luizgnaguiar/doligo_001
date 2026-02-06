# Observabilidade e Monitoramento

**Versão**: 1.0.0  
**Data**: 2026-02-05  
**Escopo**: Estratégias de Log, Métricas e Rastreamento.

---

## 1. Estrutura de Logs

O sistema utiliza **Logs Estruturados (JSON)** via `log/slog` para facilitar a ingestão e análise por ferramentas como Loki, ElasticSearch ou Datadog.

### 1.1. Campos Padrão
Todo log de aplicação deve conter:
- `level`: Nível de severidade (`INFO`, `WARN`, `ERROR`).
- `msg`: Mensagem legível para humanos.
- `time`: Timestamp ISO-8601.
- `correlation_id`: ID único da requisição (propagado do Middleware).

### 1.2. Níveis de Severidade
- **INFO**: Operações normais (ex: "Request processed", "Job started").
- **WARN**: Situações anômalas mas recuperáveis (ex: "Retry sending email", "Stock low").
- **ERROR**: Falhas de operação (ex: "DB connection failed", "Validation error").
- **CRITICAL**: Auditoria de segurança ou integridade financeira (ex: "CMP Violation", "Login Brute-force").

---

## 2. Padrões de Busca (LogQL / Grafana)

Abaixo estão exemplos de queries para o **Loki** baseadas na estrutura atual.

### 2.1. Rastreamento de Requisição
Para acompanhar o ciclo de vida de uma requisição específica (API -> DB -> Worker):
```logql
{app="doligo"} |= "correlation_id":"<SEU-UUID-AQUI>"
```

### 2.2. Monitoramento de Erros
Filtrar apenas erros reais da aplicação:
```logql
{app="doligo"} |= "level":"ERROR"
```

### 2.3. Auditoria de Alta Severidade (CMP e Segurança)
Monitorar violações de Margem (CMP) ou ações críticas:
```logql
{app="doligo"} |= "severity":"CRITICAL"
```
*Alerta Recomendado*: Disparar notificação imediata se `count > 0` em 1 minuto.

---

## 3. Monitoramento de Infraestrutura

### 3.1. Banco de Dados (PostgreSQL)
Métricas essenciais para monitorar o `database/sql` connection pool:

- **Open Connections**: Deve ser monitorado contra o limite configurado (`DB_MAX_OPEN_CONNS`).
  - *Query Sugerida*: `go_sql_db_open_connections{db="postgres"}`
- **Wait Duration**: Tempo que as goroutines esperam por uma conexão livre. Altas latências aqui indicam necessidade de aumentar o pool ou otimizar queries.
  - *Query Sugerida*: `rate(go_sql_db_wait_duration_seconds_total[1m])`

### 3.2. Workers (PDF e E-mail)
O sistema utiliza canais buferizados para processamento assíncrono.

- **Queue Depth (Saturação)**: Monitorar se o canal está cheio, o que indica gargalo.
  - *Métrica*: Tamanho atual do buffer vs Capacidade total.
- **Worker Liveness**: Como os workers rodam em goroutines perenes, monitorar logs de "Worker panic" ou "Worker restart" é crucial.
- **Timeout Rate**: Monitorar ocorrências de context timeout (30s) na geração de PDFs.
  - *Log Search*: ` "msg"="PDF generation timed out" `

---

## 4. Health Checks

O sistema expõe endpoints para verificação de disponibilidade:
- `/health`: Liveness probe (O processo está rodando?).
- `/ready`: Readiness probe (Consegue conectar ao Banco de Dados?).

Estes endpoints devem ser configurados no Kubernetes/Load Balancer para gestão de tráfego.
