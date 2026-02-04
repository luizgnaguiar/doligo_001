# Política de Headers de Segurança

**Versão**: 1.1.0
**Data**: 2026-02-04
**Status**: Implementado (Fase 13.6)

---

Este documento detalha os headers HTTP de segurança injetados automaticamente em todas as respostas da API para mitigar vetores de ataque comuns baseados em browser, bem como a política de sanitização de dados.

## Headers Implementados

### 1. X-Frame-Options
- **Valor**: `DENY`
- **Propósito**: Previne ataques de **Clickjacking**. Impede que o site seja renderizado dentro de um `<frame>`, `<iframe>`, `<embed>` ou `<object>` em qualquer outro site.
- **Referência**: [MDN X-Frame-Options](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options)

### 2. X-Content-Type-Options
- **Valor**: `nosniff`
- **Propósito**: Previne ataques de **MIME Sniffing**. Força o browser a respeitar o header `Content-Type` enviado pelo servidor, impedindo a execução de arquivos disfarçados (ex: imagem contendo script).
- **Referência**: [MDN X-Content-Type-Options](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options)

### 3. X-XSS-Protection
- **Valor**: `1; mode=block`
- **Propósito**: Ativa o filtro de **Cross-Site Scripting (XSS)** nativo dos browsers mais antigos. Se um ataque for detectado, a renderização da página é bloqueada.
- **Referência**: [MDN X-XSS-Protection](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-XSS-Protection)

### 4. Strict-Transport-Security (HSTS)
- **Valor**: `max-age=31536000; includeSubDomains`
- **Propósito**: Força o browser a se comunicar com o servidor apenas via **HTTPS** por 1 ano (31536000 segundos), incluindo todos os subdomínios. Mitiga ataques de **Man-in-the-Middle (MitM)** e **Protocol Downgrade**.
- **Referência**: [MDN Strict-Transport-Security](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security)

### 5. Content-Security-Policy (CSP)
- **Valor**: `default-src 'self'`
- **Propósito**: Define uma política restritiva onde o browser só pode carregar recursos (scripts, estilos, imagens, etc.) da mesma origem (domínio) da página. Mitiga drasticamente riscos de **XSS** e **Data Injection**.
- **Observação**: Esta é uma política básica. Para aplicações frontend complexas, ela deve ser refinada.
- **Referência**: [MDN Content-Security-Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy)

## Sanitização de Input (Defesa em Profundidade)

Além dos headers de segurança, o sistema implementa uma camada automática de sanitização de entrada para prevenir Persisted XSS e Injection Attacks:

- **Automação**: O processo de *Binding* dos dados da requisição invoca automaticamente métodos de sanitização para campos de texto (string).
- **Política**: Utiliza a biblioteca `bluemonday` com política estrita (`StrictPolicy`), removendo todas as tags HTML e prevenindo a injeção de scripts maliciosos.
- **Escopo**: Todos os DTOs de entrada (Requests) que implementam a interface `Sanitizable`.

## Configuração

A injeção destes headers é controlada pela variável de ambiente:

```bash
SECURITY_HEADERS_ENABLED=true  # Default: true
```

## Middleware

A implementação reside em `internal/api/middleware/security_headers.go` e é aplicada globalmente no entrypoint da aplicação (`cmd/main.go`). A sanitização é aplicada via Custom Binder configurado na inicialização do servidor Echo.
