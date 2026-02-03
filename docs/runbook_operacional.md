# Runbook Operacional Básico

Este documento fornece procedimentos operacionais para resposta a incidentes básicos. Siga os passos estritamente. Não tente ações não documentadas.

---

## 1. API não responde

### Sintoma
- Aplicações cliente (frontend, apps) recebem timeouts ou erros de conexão.
- Chamadas diretas via `curl` ou Postman para os endpoints da API falham.

### Como detectar
1.  **Verificar o processo:**
    ```bash
    ps aux | grep "doligo" 
    ```
    Confirme se o processo do binário da aplicação está em execução.

2.  **Verificar a porta:**
    ```bash
    netstat -tuln | grep ":<PORTA_DA_API>"
    ```
    Confirme se a porta configurada (ex: 8080) está sendo ouvida (`LISTEN`) pelo processo correto.

### Ação imediata
1.  **Coletar logs:**
    Capture as últimas 100 linhas do log de aplicação para análise posterior.
    ```bash
    tail -n 100 app.log > /tmp/incidente_api_down.log
    ```

2.  **Restart controlado:**
    Se o processo não estiver rodando ou a porta não estiver sendo ouvida, execute um restart da aplicação usando o script ou comando padrão de inicialização.
    ```bash
    # Exemplo:
    ./start_server.sh
    ```

### O que NÃO fazer
- **NÃO** reinicie o servidor ou o banco de dados sem antes coletar os logs.
- **NÃO** altere configurações de rede ou firewall sem um diagnóstico claro.

### Quando escalar
- Se o processo não iniciar após o restart.
- Se a API responder, mas continuar instável ou com erros 5xx.
- Se os logs indicarem um `panic` ou um erro fatal de inicialização (ex: falha ao conectar ao DB).

---

## 2. Erro na geração de PDF

### Sintoma
- O endpoint de geração de fatura (ou outro documento) retorna erro 500.
- O usuário relata que o download do PDF falha.

### Como detectar
1.  **Analisar logs:**
    Procure por logs de erro no `app.log` correlacionados ao `request_id` da requisição falha. Filtre por mensagens contendo `ERROR` e `invoice_generator`.
    ```bash
    grep "invoice_generator" app.log | grep "ERROR"
    ```

### Ação imediata
1.  **Diferenciar o tipo de erro:**
    - **Erro de Dados:** Se o log indicar um problema com os dados de entrada (ex: "campo X não pode ser nulo", "formato de data inválido"), o problema está nos dados da fatura. **Ação:** Registre o ID da fatura e informe à equipe de suporte ou produto para correção dos dados.
    - **Erro de Execução:** Se o log indicar um problema interno da biblioteca `maroto` ou `context deadline exceeded`, é uma falha no processamento. **Ação:** Nenhuma ação corretiva imediata é possível.

### O que NÃO fazer
- **NÃO** tente reenviar a requisição para o mesmo documento se for um erro de dados.
- **NÃO** reinicie o serviço. A geração de PDF é uma operação isolada e um restart não corrigirá um erro de lógica ou de dados.

### Quando escalar
- Se múltiplos PDFs de diferentes faturas/clientes estiverem falhando (sugere um problema sistêmico).
- Se o log mostrar `context deadline exceeded` com frequência (pode indicar um problema de performance).
- Se o log apresentar um `panic`.

---

## 3. Falha no envio de e-mail

### Sintoma
- Usuários não recebem e-mails transacionais (ex: confirmação de cadastro, recuperação de senha).

### Como detectar
1.  **Analisar logs:**
    Procure por erros relacionados ao `email.sender` ou `worker`.
    ```bash
    grep "email" app.log | grep "ERROR"
    ```

2.  **Verificar Circuit Breaker:**
    A aplicação possui um retry simples. Se os logs mostrarem falhas repetidas para o mesmo serviço de e-mail, o "circuit breaker" (implementação manual) pode estar "aberto", cessando as tentativas temporariamente. Procure por mensagens como "serviço de e-mail indisponível, nova tentativa em X segundos".

### Ação imediata
1.  **Aguardar:**
    Se o circuit breaker estiver ativo, o comportamento esperado é que o sistema se recupere sozinho. Aguarde o tempo indicado no log antes de uma nova tentativa.

2.  **Verificar serviço externo:**
    Se as falhas persistirem, verifique o status page do provedor de e-mail (ex: SendGrid, Mailgun).

### O que NÃO fazer
- **NÃO** reinicie a aplicação. O worker de e-mail e o estado do circuit breaker são gerenciados em memória e um restart pode perder tarefas enfileiradas ou resetar o backoff, causando mais carga no serviço externo.
- **NÃO** tente disparar e-mails manualmente pela API para contornar o problema.

### Quando escalar
- Se o serviço de e-mail externo estiver operacional, mas a aplicação continuar registrando falhas de conexão/autenticação.
- Se o worker de tarefas internas (que processa e-mails) parar de funcionar (verificado via logs ou métricas internas, se disponíveis).

---

## 4. Erro de banco de dados

### Sintoma
- API retorna erros 500 em múltiplos endpoints.
- Logs mostram erros de `SQLSTATE`, `connection refused`, ou `migration`.

### Como detectar
1.  **Logs de Conexão:**
    Filtre o `app.log` por erros do driver de banco de dados (`pgx`, `mysql`).
    ```bash
    grep "database" app.log | grep "ERROR"
    ```
    Procure por "connection refused" ou "failed to connect".

2.  **Logs de Migração:**
    Na inicialização da aplicação, verifique se há erros com `golang-migrate`.
    ```bash
    grep "migration" app.log | grep "ERROR"
    ```
    Procure por "migration file not found" ou "dirty migration".

### Ação imediata
- **Falha de conexão:**
  1. Verifique a conectividade de rede entre o servidor da aplicação e o servidor de banco de dados (`ping`, `telnet <db_host> <db_port>`).
  2. Verifique o status do serviço do banco de dados no servidor correspondente.
  **Ação segura:** Nenhuma. Apenas diagnóstico.

- **Migração Pendente/Falha:**
  A aplicação não deve iniciar com migrações falhas. Se isso ocorrer, a implantação falhou.
  **Ação segura:** Reverter o deploy para a versão anterior.

### O que NÃO fazer
- **NÃO** aplique migrações SQL manualmente no banco de dados de produção.
- **NÃO** altere as credenciais de conexão na aplicação em execução.
- **NÃO** reinicie o banco de dados sem coordenação com a equipe de infraestrutura.

### Quando escalar
- **IMEDIATAMENTE** para qualquer erro de banco de dados (conexão, migração, query). Problemas de DB são críticos e requerem análise especializada.

---

## 5. Lentidão geral

### Sintoma
- A API responde, mas com alta latência em vários endpoints.
- Usuários relatam que a aplicação está "lenta".

### Como detectar
1.  **Analisar Métricas (se existentes):**
    Consulte o endpoint `/metrics` (se implementado e exposto) para verificar latências de requisição (`http_request_duration_seconds`) e uso de recursos.

2.  **Analisar Logs:**
    Procure por logs de requisições lentas. O `request_logger` deve registrar a duração de cada chamada. Identifique se a lentidão está concentrada em endpoints específicos.
    ```bash
    # Exemplo hipotético de análise de log
    cat app.log | grep "request completed" | grep "duration_ms=[1-9][0-9]{3,}"
    ```

3.  **Diferenciar Causa:**
    - **Lentidão em endpoints de PDF:** Isole se a lentidão ocorre apenas na geração de documentos. Isso aponta para um gargalo de CPU na geração do PDF.
    - **Lentidão em endpoints de DB:** Se múltiplos endpoints que acessam o banco estão lentos, pode ser uma sobrecarga no DB ou queries ineficientes.

### Ação imediata
- Nenhuma ação corretiva direta é segura. O objetivo é coletar dados para o diagnóstico.
- Colete os logs das requisições lentas.
- Anote quais endpoints são os mais afetados.

### O que NÃO fazer
- **NÃO** reinicie a aplicação. Um restart pode mascarar a causa raiz (ex: limpando um cache de query lento) e não resolve o problema subjacente.
- **NÃO** aumente o número de réplicas da aplicação sem saber a causa. Se o gargalo for o banco de dados, mais réplicas irão piorar a situação.

### Quando escalar
- Se a latência média da API exceder um limite aceitável (ex: > 1 segundo) por mais de 5 minutos.
- Se a lentidão estiver correlacionada a um aumento no uso de CPU ou memória do servidor da aplicação ou do banco de dados.
- Se a causa não for claramente identificável como um endpoint de PDF específico.
