# Configuração de Timeout para Reverse Proxy

**Versão**: 1.0.0  
**Data**: 2026-02-03
**Escopo**: Recomendações técnicas para produção

---

## 1. Contexto Técnico

Este documento define valores mínimos de timeout para reverse proxies baseando-se em operações longas conhecidas da aplicação. A configuração inadequada de timeouts em proxies pode levar a erros `504 Gateway Timeout` prematuros, interrompendo operações válidas.

### Operações Longas Identificadas

| Endpoint | Operação | Timeout Interno (Go) | Timeout Mínimo Recomendado (Proxy) |
|----------|----------|----------------------|-------------------------------------|
| `/api/v1/invoices/:id/pdf` | Geração de PDF de Fatura | 30s | 45s |
| `/api/v1/margin` | Agregação SQL de Margens | N/A | 60s |
| `/api/v1/boms/produce` | Processamento de Ordem de Produção | N/A | 60s |

**Justificativa dos Valores**:
- **PDF**: O timeout interno da aplicação para geração de PDF é de 30 segundos. Adicionamos uma margem de segurança de 15 segundos (50%) para acomodar latência de rede e overhead do proxy, resultando em 45 segundos.
- **Margin**: A agregação de dados para relatórios de margem pode ser computacionalmente intensiva, envolvendo queries complexas. Na ausência de um timeout interno explícito, um valor conservador de 60 segundos é recomendado para permitir a conclusão de relatórios sobre grandes volumes de dados.
- **BOM Production**: A produção a partir de uma Bill of Materials é uma operação transacional que pode envolver múltiplos bloqueios de banco de dados (pessimistic locking) para garantir a consistência do estoque. Um timeout de 60 segundos é um ponto de partida seguro para evitar interrupções durante a transação.

---

## 2. Nginx

### Configuração Mínima Recomendada

A configuração pode ser aplicada globalmente ou de forma granular por `location`. A abordagem granular é preferível para evitar manter conexões abertas desnecessariamente em endpoints rápidos.

```nginx
# Configuração granular para o endpoint de PDF
location /api/v1/invoices/ {
    proxy_pass http://backend; # Substituir 'backend' pelo upstream real
    proxy_read_timeout 45s;
    proxy_connect_timeout 10s;
    proxy_send_timeout 10s;
}

# Configuração para endpoints com transações e relatórios longos
location ~ ^/api/v1/(margin|boms/produce) {
    proxy_pass http://backend;
    proxy_read_timeout 60s;
    proxy_connect_timeout 10s;
    proxy_send_timeout 10s;
}

# Configuração global alternativa (menos granular, aplicar no bloco http, server ou location)
# proxy_read_timeout 60s;
```

**Parâmetros Críticos**:
- `proxy_read_timeout`: Tempo máximo que o Nginx aguardará por uma resposta do serviço Go após o envio da requisição. Este é o parâmetro mais crítico para operações longas.
- `proxy_connect_timeout`: Tempo máximo para estabelecer uma conexão com o serviço Go.
- `proxy_send_timeout`: Tempo máximo para enviar a requisição completa ao serviço Go.

---

## 3. Traefik

### Configuração via Labels (Docker)

Esta abordagem é comum em ambientes de contêineres, aplicando a configuração diretamente no serviço.

```yaml
# labels do serviço no docker-compose.yml
labels:
  # Roteador principal
  - "traefik.http.routers.app.rule=Host(`seu-dominio.com`)"
  - "traefik.http.routers.app.service=app-service"

  # Middleware de timeout para o serviço
  - "traefik.http.routers.app.middlewares=app-timeout"
  
  # Definição do middleware de timeout
  # Define um timeout de resposta de 60 segundos para todas as rotas
  - "traefik.http.middlewares.app-timeout.forwarding.responseTimeout=60s"

  # Definição do serviço
  - "traefik.http.services.app-service.loadbalancer.server.port=8080"
```

### Configuração via Arquivo Dinâmico (YAML)

Para uma configuração mais centralizada, pode-se usar um arquivo de configuração dinâmico.

```yaml
# config/dynamic.yml
http:
  routers:
    app-router:
      rule: "Host(`seu-dominio.com`)"
      service: "app-service"
      middlewares:
        - "app-timeout-middleware"

  services:
    app-service:
      loadBalancer:
        servers:
          - url: "http://backend:8080" # Aponta para o container da aplicação Go

  middlewares:
    app-timeout-middleware:
      forwarding:
        responseTimeout: "60s" # Timeout para a resposta do backend
```

**Nota**: Em Traefik, `responseTimeout` é o tempo que o Traefik aguarda pela resposta completa do backend. Ajustar este valor é crucial para endpoints de longa duração.

---

## 4. Restrições e Avisos

### O que NÃO está incluído
- Benchmarks de performance ou testes de carga.
- Configuração de cache, rate limiting ou circuit breakers no nível do proxy.
- Otimização de throughput ou ajuste fino de keep-alive.

### Riscos Conhecidos Aceitos
- Os valores são baseados em análise estática e margens de segurança teóricas, não em medições de carga real. Ambientes de produção com alto volume podem exigir ajustes.
- Faturas com um número extremamente grande de itens (>1000) podem, teoricamente, exceder o timeout de 30s da aplicação, e consequentemente o do proxy.
- A configuração global de timeout pode afetar a resiliência de endpoints que deveriam falhar rapidamente.

### Quando Revisar
- Após qualquer incidente de `504 Gateway Timeout` em produção.
- Ao adicionar novos endpoints que executem operações de longa duração (ex: relatórios complexos, exportações de dados).
- Após mudanças significativas no volume de dados que possam impactar o tempo de execução das queries.

---

## 5. Validação

Para uma verificação básica se os timeouts estão sendo respeitados pelo proxy:

```bash
# Simular uma requisição longa para o endpoint de margem
# O comando deve demorar, mas concluir antes do timeout do proxy (e.g., < 60s)
time curl -X GET "http://localhost/api/v1/margin?startDate=2020-01-01&endDate=2024-12-31"

# Verificar os logs de erro do proxy para identificar timeouts prematuros
# Nginx: tipicamente em /var/log/nginx/error.log
# Traefik: no stdout do container ou no arquivo de log configurado
```

Se um erro `504 Gateway Timeout` ocorrer antes do tempo configurado (`proxy_read_timeout` ou `responseTimeout`), a configuração do proxy é o ponto de partida para a investigação.

---

**Documento produzido em conformidade com FASE 12.2**  
**Complementado por `docs/production_notes.md` (FASE 17.2)**  
**Nenhuma configuração real foi aplicada**
