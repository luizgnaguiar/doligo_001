# Política de Headers de Segurança

**Versão**: 1.0.0
**Data**: 2026-02-03
**Status**: Implementado (Fase 13.3)

---

Este documento detalha os headers HTTP de segurança injetados automaticamente em todas as respostas da API para mitigar vetores de ataque comuns baseados em browser.

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

## Configuração

A injeção destes headers é controlada pela variável de ambiente:

```bash
SECURITY_HEADERS_ENABLED=true  # Default: true
```

## Middleware

A implementação reside em `internal/api/middleware/security_headers.go` e é aplicada globalmente no entrypoint da aplicação (`cmd/main.go`).
