# Runbook de Operação e Resposta a Incidentes (Dolibarr-Go)

Este documento descreve os procedimentos operacionais padrão para diagnóstico e resolução de incidentes no sistema ERP/CRM Dolibarr-Go.

**Mecanismos de Observabilidade Relacionados:**
- Logs Estruturados (JSON)
- Níveis de Severidade (INFO, WARN, ERROR, FATAL)
- Correlation ID (rastreabilidade entre camadas)
- Tabela `audit_logs` para registro persistente de ações críticas

---

## Cenário A: Falha no Banco de Dados

**Sintomas:**
- Logs da aplicação indicam erros de conexão (`connection refused`, `timeout`).
- Endpoint `/health` retorna status não saudável ou erro 500.
- Aplicação falha ao iniciar.

**Procedimento de Diagnóstico:**
1. Verifique os logs da aplicação buscando por erros fatais na inicialização ou durante transações.
   ```bash
   grep "FATAL" app.log | grep "database"
   ```
2. Teste a conectividade básica com o banco de dados a partir do host da aplicação.
   ```bash
   # Exemplo para PostgreSQL
   pg_isready -h $DB_HOST -p $DB_PORT
   ```

**Procedimento de Recuperação:**
1. **Restabelecimento de Serviço:**
   - Se o banco estiver parado, inicie o serviço do banco de dados.
   - Verifique se as credenciais no arquivo `.env` ou variáveis de ambiente estão corretas.

2. **Aplicação de Migrações:**
   - Se o erro indicar tabelas inexistentes ou incompatibilidade de schema, execute as migrações manualmente (se o auto-migrate falhou).
   - O binário possui migrações embutidas. Reiniciar a aplicação geralmente tenta reaplicar as migrações pendentes.
   - Verifique os logs de migração na inicialização:
     ```bash
     grep "Migração" app.log
     ```

---

## Cenário B: Worker Pool Travado (Processamento de PDF)

**Sintomas:**
- Solicitações de geração de PDF (Faturas/Relatórios) ficam pendentes indefinidamente.
- Clientes não recebem e-mails com anexos esperados.
- Logs de auditoria não mostram conclusão de tarefas.

**Procedimento de Diagnóstico:**
1. **Verificação de Logs de Auditoria e Severidade:**
   - Consulte a tabela `audit_logs` ou os logs da aplicação filtrando por severidade `ERROR` no componente de Worker.
   - Busque por mensagens de timeout ou erros no processamento de filas.
   ```sql
   SELECT * FROM audit_logs WHERE action = 'PDF_GENERATION' AND severity = 'ERROR' ORDER BY timestamp DESC LIMIT 10;
   ```
2. **Identificação de Tarefas Travadas:**
   - Verifique se há goroutines bloqueadas ou vazamento de recursos nos logs de métricas (se habilitados).

**Procedimento de Recuperação:**
1. **Reinício Controlado:**
   - A aplicação deve ser reiniciada para limpar o estado do Worker Pool em memória.
   - Envie um sinal `SIGTERM` para um *Graceful Shutdown* (o sistema tentará processar o buffer por até 15s).
   ```bash
   kill -SIGTERM <PID_DA_APLICACAO>
   ```
2. **Monitoramento Pós-Reinício:**
   - Acompanhe os logs para garantir que o Worker Pool reiniciou e está consumindo novas tarefas.

---

## Cenário C: Excesso de Rate Limiting

**Sintomas:**
- Usuários legítimos ou integrações recebem erros `429 Too Many Requests`.
- Logs de acesso mostram múltiplos bloqueios para mesmos IPs.

**Procedimento de Diagnóstico:**
1. **Identificação de IPs Bloqueados:**
   - Analise os logs buscando pelo middleware de Rate Limiting.
   ```bash
   grep "Rate limit exceeded" app.log | jq '.ip'
   ```
2. **Verificação de Configuração:**
   - Confirme os valores atuais de `RATE_LIMIT_REQUESTS` e `RATE_LIMIT_DURATION` nas variáveis de ambiente.

**Procedimento de Recuperação:**
1. **Ajuste de Configuração (ENV):**
   - Se o tráfego for legítimo (ex: campanha de vendas, integração em batch), aumente os limites no arquivo de configuração ou variáveis de ambiente.
   - Exemplo `.env`:
     ```
     RATE_LIMIT_REQUESTS=100
     RATE_LIMIT_DURATION=1m
     ```
2. **Aplicação:**
   - Reinicie a aplicação para que as novas configurações de Rate Limiting entrem em vigor.

---

## Diagnóstico via Correlação (Correlation ID)

O sistema injeta um `X-Correlation-ID` em cada requisição HTTP e o propaga para logs e contexto de execução (incluindo tarefas assíncronas).

**Como Utilizar:**

1. **Obter o ID:**
   - Solicite ao usuário o ID retornado no cabeçalho da resposta HTTP (`X-Correlation-ID`) ou na mensagem de erro JSON.
   
2. **Rastrear no Log da Aplicação:**
   - Utilize o ID para filtrar *todos* os logs relacionados àquela transação específica, desde a entrada na API até a execução no banco ou worker.
   ```bash
   grep "correlation_id_aqui" app.log
   ```

3. **Correlacionar com Auditoria:**
   - Utilize o mesmo ID para buscar registros na tabela `audit_logs`. O campo `details` ou metadados podem conter o ID de correlação se persistido (dependendo da implementação de log de auditoria específica).
   
Este mecanismo permite isolar uma falha específica em meio a logs concorrentes, facilitando a identificação da causa raiz (ex: erro de validação vs erro de banco de dados).
