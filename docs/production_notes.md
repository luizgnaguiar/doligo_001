# Production Notes & Infrastructure Recommendations

**Versão**: 1.1.0  
**Data**: 2026-02-05
**Status**: Documentação Técnica para Produção

---

## 1. Configuração de Reverse Proxy

A infraestrutura de Proxy Reverso (Nginx, Traefik, etc.) deve ser configurada para suportar as operações de longa duração da API Doligo, especificamente a geração de PDFs (timeout interno de 30s) e relatórios complexos.

### 1.1. Nginx

Para evitar erros `504 Gateway Timeout`, os valores de `proxy_read_timeout` devem exceder o timeout interno da aplicação.

```nginx
upstream doligo_backend {
    server backend:8080;
    # Manter conexões abertas com o backend para reduzir overhead de handshake TCP/TLS
    keepalive 32;
}

server {
    listen 80;
    server_name api.doligo.com;

    location / {
        proxy_pass http://doligo_backend;
        proxy_http_version 1.1;
        proxy_set_header Connection ""; # Necessário para keepalive do upstream
        
        # Timeouts de Conexão e Escrita
        proxy_connect_timeout 10s;
        proxy_send_timeout 30s;

        # Timeout de Leitura (Crítico para PDF/Relatórios)
        # Definido como 45s para acomodar o limite interno de 30s + latência
        proxy_read_timeout 45s;

        # Buffer de Proxy (Recomendado para respostas grandes)
        proxy_buffers 8 16k;
        proxy_buffer_size 32k;
    }
}
```

### 1.2. Traefik

Configuração recomendada via arquivo dinâmico ou labels, focando no `responseTimeout`.

```yaml
# Configuração via Dynamic File (YAML)
http:
  services:
    doligo-service:
      loadBalancer:
        servers:
          - url: "http://backend:8080"
        # Configuração de Transport para Keep-Alive
        proxyProtocol:
          version: 2
    
  middlewares:
    doligo-timeouts:
      forwarding:
        # Garante que o Traefik aguarde a geração do PDF (30s)
        responseTimeout: "45s"
        idleTimeout: "60s"

  routers:
    doligo-router:
      rule: "Host(`api.doligo.com`)"
      service: doligo-service
      middlewares:
        - doligo-timeouts
```

---

## 2. Gestão de Conexões e Concorrência

Para suportar ambientes de alta concorrência (validados na Fase 15), recomenda-se:

- **Keep-Alive**: Manter conexões persistentes entre o Proxy e a API para evitar o custo de "TCP Three-Way Handshake" em cada requisição.
  - **Nginx**: Usar `keepalive` no bloco `upstream`.
  - **Traefik**: Gerenciado automaticamente, mas ajustável via `transport`.
- **Limites de File Descriptors**: O servidor que hospeda o binário e o proxy deve ter o `ulimit -n` aumentado (mínimo 65535).

---

## 3. Compatibilidade com Ciclo de Vida da Aplicação

### 3.1. Graceful Shutdown vs Timeouts
A aplicação Doligo possui uma política de **Graceful Shutdown de 15 segundos**.

- **Conflito Teórico**: Se uma tarefa de PDF (30s) inicia e a aplicação recebe um `SIGTERM` logo em seguida, a tarefa será cancelada após 15s pela lógica de shutdown, mesmo que o Proxy esteja disposto a esperar 45s.
- **Resiliência**: O Proxy deve manter o timeout de 45s para garantir que, em condições normais de operação, o cliente receba o arquivo. O cancelamento prematuro durante o shutdown é um comportamento esperado para garantir a integridade do processo de parada.

### 3.2. Health Checks
O Proxy deve utilizar o endpoint `/health` para monitoramento de liveness e readiness:
- **Intervalo**: 5s a 10s.
- **Falhas consecutivas**: 3 (para marcar como unhealthy).

---

## 4. Segurança de Rede (Recomendações)

- **TLS Termination**: Sempre realize a terminação TLS no Reverse Proxy.
- **Headers de Segurança**: O Proxy deve injetar headers como `X-Content-Type-Options: nosniff` e `Strict-Transport-Security` se não forem providos pela aplicação.
- **HSTS**: Ativar para forçar conexões HTTPS.

---
**Nota Operacional**: Estas configurações são recomendações de arquitetura. A implementação real depende do orquestrador (Kubernetes, Docker Swarm) ou do sistema operacional host.
