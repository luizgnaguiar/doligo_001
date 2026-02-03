# Variáveis de Ambiente

Este documento detalha todas as variáveis de ambiente utilizadas pela aplicação.

---

### `APP_ENV`
- **Descrição**: Define o ambiente em que a aplicação está rodando (ex: `development`, `production`).
- **Tipo**: `String`
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `development`
- **Impacto se Ausente ou Inválida**: A aplicação usará o valor padrão `development`. Pode afetar a configuração de logs ou outros comportamentos dependentes do ambiente.

---

### `PORT`
- **Descrição**: Especifica a porta na qual o servidor HTTP irá escutar.
- **Tipo**: `String`
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `8080`
- **Impacto se Ausente ou Inválida**: A aplicação usará a porta padrão `8080`. Se a porta estiver em uso, a aplicação falhará ao iniciar.

---

### `LOG_LEVEL`
- **Descrição**: Define o nível mínimo de severidade para os logs que serão registrados.
- **Tipo**: `String`
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `info`
- **Impacto se Ausente ou Inválida**: A aplicação usará o nível de log padrão `info`. Um valor inválido pode fazer com que nenhum log seja exibido.

---

### `JWT_SECRET`
- **Descrição**: Chave secreta utilizada para assinar e verificar os tokens JWT (JSON Web Tokens).
- **Tipo**: `String`
- **Obrigatoriedade**: Opcional (altamente recomendado para produção)
- **Valor Padrão**: `super-secret-jwt-key`
- **Impacto se Ausente ou Inválida**: A aplicação usará a chave padrão, o que é inseguro para ambientes de produção. A autenticação de usuários não funcionará corretamente se a chave for alterada enquanto tokens ainda são válidos.

---

### `DB_TYPE`
- **Descrição**: Define o tipo do banco de dados a ser utilizado.
- **Tipo**: `String`
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `postgres`
- **Impacto se Ausente ou Inválida**: A aplicação tentará se conectar a um banco de dados PostgreSQL. Um tipo de banco de dados não suportado causará falha na inicialização.

---

### `DB_HOST`
- **Descrição**: Endereço do servidor de banco de dados.
- **Tipo**: `String`
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `localhost`
- **Impacto se Ausente ou Inválida**: A aplicação tentará se conectar ao `localhost`. Se o banco de dados estiver em outro host, a conexão falhará.

---

### `DB_PORT`
- **Descrição**: Porta do servidor de banco de dados.
- **Tipo**: `Integer`
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `5432`
- **Impacto se Ausente ou Inválida**: A aplicação usará a porta padrão `5432`. Uma porta incorreta impedirá a conexão com o banco de dados.

---

### `DB_USER`
- **Descrição**: Nome de usuário para autenticação no banco de dados.
- **Tipo**: `String`
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `user`
- **Impacto se Ausente ou Inválida**: A conexão com o banco de dados falhará se o usuário for inválido ou não tiver as permissões necessárias.

---

### `DB_PASSWORD`
- **Descrição**: Senha para autenticação no banco de dados.
- **Tipo**: `String`
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `password`
- **Impacto se Ausente ou Inválida**: A autenticação no banco de dados falhará.

---

### `DB_NAME`
- **Descrição**: Nome do banco de dados (database/schema) a ser utilizado.
- **Tipo**: `String`
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `dolibarr`
- **Impacto se Ausente ou Inválida**: A aplicação não conseguirá selecionar o banco de dados correto, resultando em falha na conexão ou tabelas não encontradas.

---

### `DB_SSLMODE`
- **Descrição**: Modo de conexão SSL com o banco de dados (ex: `disable`, `require`, `verify-full`).
- **Tipo**: `String`
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `disable`
- **Impacto se Ausente ou Inválida**: A aplicação pode não conseguir se conectar ao banco de dados se ele exigir um modo SSL específico.

---

### `DB_MAX_OPEN_CONNS`
- **Descrição**: Número máximo de conexões abertas com o banco de dados.
- **Tipo**: `Integer`
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `10`
- **Impacto se Ausente ou Inválida**: Um valor muito baixo pode degradar a performance sob carga. Um valor muito alto pode sobrecarregar o banco de dados.

---

### `DB_MAX_IDLE_CONNS`
- **Descrição**: Número máximo de conexões inativas no pool de conexões.
- **Tipo**: `Integer`
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `5`
- **Impacto se Ausente ou Inválida**: Afeta a reutilização de conexões e a performance da aplicação.

---

### `DB_CONN_MAX_LIFETIME`
- **Descrição**: Tempo máximo de vida de uma conexão com o banco de dados.
- **Tipo**: `Duration` (ex: `5m`, `1h`)
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `5m`
- **Impacto se Ausente ou Inválida**: Conexões podem se tornar obsoletas, causando erros em ambientes com firewalls ou proxies.

---

### `INTERNAL_WORKER_POOL_SIZE`
- **Descrição**: Número de workers (goroutines) no pool de tarefas internas.
- **Tipo**: `Integer`
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `5`
- **Impacto se Ausente ou Inválida**: Afeta a concorrência e a vazão do processamento de tarefas em segundo plano.

---

### `INTERNAL_WORKER_SHUTDOWN_TIMEOUT`
- **Descrição**: Tempo máximo de espera para as tarefas do worker pool concluírem durante o graceful shutdown.
- **Tipo**: `Duration` (ex: `15s`, `1m`)
- **Obrigatoriedade**: Opcional
- **Valor Padrão**: `15s`
- **Impacto se Ausente ou Inválida**: Tarefas em andamento podem ser interrompidas abruptamente se o tempo for muito curto.
