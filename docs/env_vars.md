# Variáveis de Ambiente

**Versão**: 1.0.1  
**Data**: 2026-02-03
**Escopo**: Aplicação completa

---

## 1. APP_ENV

- **Descrição**: Define o ambiente de execução da aplicação (e.g., "development", "production").
- **Tipo**: string
- **Obrigatório**: NÃO
- **Valor Default**: `development`
- **Impacto se Ausente**: A aplicação assume o ambiente de desenvolvimento, o que pode afetar o nível de log e outros comportamentos.
- **Exemplo**:
  ```
  APP_ENV=production
  ```

---

## 2. PORT

- **Descrição**: Define a porta TCP onde o servidor web irá escutar.
- **Tipo**: int
- **Obrigatório**: NÃO
- **Valor Default**: `8080`
- **Impacto se Ausente**: O servidor web tentará iniciar na porta 8080.
- **Exemplo**:
  ```
  PORT=3000
  ```

---

## 3. DB_TYPE

- **Descrição**: Especifica o dialeto do banco de dados a ser utilizado.
- **Tipo**: string
- **Obrigatório**: NÃO
- **Valor Default**: `postgres`
- **Impacto se Ausente**: O sistema tentará se conectar a um banco de dados PostgreSQL.
- **Exemplo**:
  ```
  DB_TYPE=mysql
  ```

---

## 4. DB_HOST

- **Descrição**: Endereço do servidor de banco de dados.
- **Tipo**: string
- **Obrigatório**: NÃO
- **Valor Default**: `localhost`
- **Impacto se Ausente**: O sistema tentará se conectar ao banco de dados no host local.
- **Exemplo**:
  ```
  DB_HOST=db.meu-servidor.com
  ```

---

## 5. DB_PORT

- **Descrição**: Porta do servidor de banco de dados.
- **Tipo**: int
- **Obrigatório**: NÃO
- **Valor Default**: `5432`
- **Impacto se Ausente**: O sistema tentará se conectar à porta padrão do PostgreSQL.
- **Exemplo**:
  ```
  DB_PORT=5433
  ```

---

## 6. DB_USER

- **Descrição**: Nome de usuário para autenticação no banco de dados.
- **Tipo**: string
- **Obrigatório**: NÃO
- **Valor Default**: `user`
- **Impacto se Ausente**: A conexão usará o nome de usuário "user".
- **Exemplo**:
  ```
  DB_USER=admin
  ```

---

## 7. DB_PASSWORD

- **Descrição**: Senha para autenticação no banco de dados.
- **Tipo**: string
- **Obrigatório**: NÃO
- **Valor Default**: `password`
- **Impacto se Ausente**: A conexão usará a senha "password".
- **Exemplo**:
  ```
  DB_PASSWORD=senha_super_segura
  ```

---

## 8. DB_NAME

- **Descrição**: Nome do banco de dados (database/schema) a ser utilizado.
- **Tipo**: string
- **Obrigatório**: NÃO
- **Valor Default**: `dolibarr`
- **Impacto se Ausente**: O sistema tentará se conectar ao banco de dados "dolibarr".
- **Exemplo**:
  ```
  DB_NAME=erp_producao
  ```

---

## 9. DB_SSLMODE

- **Descrição**: Modo de conexão SSL com o banco de dados.
- **Tipo**: string
- **Obrigatório**: NÃO
- **Valor Default**: `disable`
- **Impacto se Ausente**: A conexão com o banco de dados não usará SSL, o que é inseguro para ambientes de produção.
- **Exemplo**:
  ```
  DB_SSLMODE=require
  ```

---

## 10. DB_MAX_OPEN_CONNS

- **Descrição**: Número máximo de conexões abertas com o banco de dados.
- **Tipo**: int
- **Obrigatório**: NÃO
- **Valor Default**: `10`
- **Impacto se Ausente**: O pool de conexões terá no máximo 10 conexões ativas.
- **Exemplo**:
  ```
  DB_MAX_OPEN_CONNS=100
  ```

---

## 11. DB_MAX_IDLE_CONNS

- **Descrição**: Número máximo de conexões inativas no pool de conexões.
- **Tipo**: int
- **Obrigatório**: NÃO
- **Valor Default**: `5`
- **Impacto se Ausente**: O pool de conexões manterá no máximo 5 conexões inativas.
- **Exemplo**:
  ```
  DB_MAX_IDLE_CONNS=20
  ```

---

## 12. DB_CONN_MAX_LIFETIME

- **Descrição**: Tempo máximo que uma conexão pode ser reutilizada (formato: 5m, 1h).
- **Tipo**: duration
- **Obrigatório**: NÃO
- **Valor Default**: `5m`
- **Impacto se Ausente**: As conexões serão reutilizadas por no máximo 5 minutos.
- **Exemplo**:
  ```
  DB_CONN_MAX_LIFETIME=1h
  ```

---

## 13. LOG_LEVEL

- **Descrição**: Nível de verbosidade dos logs da aplicação.
- **Tipo**: string
- **Obrigatório**: NÃO
- **Valor Default**: `info`
- **Impacto se Ausente**: Apenas logs de nível "info" ou superior (warn, error, fatal) serão exibidos.
- **Exemplo**:
  ```
  LOG_LEVEL=debug
  ```

---

## 14. INTERNAL_WORKER_POOL_SIZE

- **Descrição**: Número de goroutines no pool de workers para tarefas em background.
- **Tipo**: int
- **Obrigatório**: NÃO
- **Valor Default**: `5`
- **Impacto se Ausente**: O worker pool será iniciado com 5 workers.
- **Exemplo**:
  ```
  INTERNAL_WORKER_POOL_SIZE=20
  ```

---

## 15. INTERNAL_WORKER_SHUTDOWN_TIMEOUT

- **Descrição**: Tempo máximo de espera para as tarefas em background finalizarem durante um graceful shutdown (formato: 15s, 1m).
- **Tipo**: duration
- **Obrigatório**: NÃO
- **Valor Default**: `15s`
- **Impacto se Ausente**: A aplicação aguardará até 15 segundos para os workers finalizarem antes de forçar o encerramento.
- **Exemplo**:
  ```
  INTERNAL_WORKER_SHUTDOWN_TIMEOUT=30s
  ```

---

## 16. JWT_SECRET

- **Descrição**: Chave secreta para assinar e verificar tokens JWT.
- **Tipo**: string
- **Obrigatório**: NÃO
- **Valor Default**: `super-secret-jwt-key`
- **Impacto se Ausente**: A segurança da autenticação estará comprometida se a chave default for usada em produção.
- **Exemplo**:
  ```
  JWT_SECRET=uma_chave_secreta_longa_e_dificil_de_adivinhar
  ```

---

## 17. RATE_LIMIT_ENABLED

- **Descrição**: Habilita ou desabilita o middleware de Rate Limiting.
- **Tipo**: bool
- **Obrigatório**: NÃO
- **Valor Default**: `false`
- **Impacto se Ausente**: Rate limiting estará desligado.
- **Exemplo**:
  ```
  RATE_LIMIT_ENABLED=true
  ```

---

## 18. RATE_LIMIT_REQUESTS_PER_SECOND

- **Descrição**: Número máximo de requisições permitidas por segundo (sustentado) por IP.
- **Tipo**: int
- **Obrigatório**: NÃO
- **Valor Default**: `10`
- **Impacto se Ausente**: Utilizará o valor padrão de 10 RPS.
- **Exemplo**:
  ```
  RATE_LIMIT_REQUESTS_PER_SECOND=20
  ```

---

## 19. RATE_LIMIT_BURST

- **Descrição**: Número máximo de requisições instantâneas permitidas (burst) por IP.
- **Tipo**: int
- **Obrigatório**: NÃO
- **Valor Default**: `20`
- **Impacto se Ausente**: Utilizará o valor padrão de burst 20.
- **Exemplo**:
  ```
  RATE_LIMIT_BURST=40
  ```
