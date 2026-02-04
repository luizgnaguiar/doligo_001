# Rate Limiting

**Versão**: 1.1.0
**Data**: 2026-02-04
**Status**: Implementado (Fase 13.6)

## Visão Geral

O sistema implementa um mecanismo de **Rate Limiting (Limitação de Taxa)** para proteger a API contra abuso, negação de serviço (DoS) e tráfego excessivo. A implementação atual é **in-memory** e baseada no endereço IP do cliente.

## Implementação Técnica

- **Biblioteca**: `golang.org/x/time/rate` (Token Bucket algorithm).
- **Escopo**: Global (aplicado a todas as rotas da API).
- **Armazenamento**: Mapa em memória (`map[string]*clientLimiter`), onde a chave é o IP do cliente.
- **Identificação**: O IP é obtido via `c.RealIP()`, que respeita headers como `X-Forwarded-For` e `X-Real-IP`.

## Configuração

A configuração é realizada via variáveis de ambiente.

| Variável | Descrição | Padrão |
| :--- | :--- | :--- |
| `RATE_LIMIT_ENABLED` | Habilita/Desabilita o rate limiting | `false` |
| `RATE_LIMIT_REQUESTS_PER_SECOND` | Taxa de requisições permitidas por segundo (r/s) | `10` |
| `RATE_LIMIT_BURST` | Capacidade máxima de "explosão" de requisições | `20` |

### Exemplo de Comportamento

Com `RPS=10` e `BURST=20`:
1. Um cliente pode fazer até 20 requisições instantaneamente.
2. Após consumir o burst, ele fica limitado a 10 requisições por segundo.
3. Se exceder, recebe `429 Too Many Requests`.

## Resposta de Bloqueio

Quando o limite é excedido, a API retorna:

- **Status Code**: `429 Too Many Requests`
- **Header**: `Retry-After: 60` (sugestão de tempo de espera em segundos)
- **Body**:
  ```json
  {
    "error": "rate limit exceeded"
  }
  ```

## Logs

Requisições bloqueadas são registradas com nível `WARN`:

```
level=WARN msg="Rate limit exceeded" ip=203.0.113.1 path=/api/v1/invoices correlation_id=...
```

## Gestão de Recursos (Cleanup)

Para prevenir vazamento de memória (memory leaks) decorrente do armazenamento de limitadores para IPs que não interagem mais com o sistema:

- **Rastreamento**: Cada limitador armazena a data/hora da última requisição (`lastSeen`).
- **Limpeza Automática**: Uma goroutine executa a cada 10 minutos.
- **Política**: Limitadores de IPs inativos há mais de 10 minutos são removidos da memória.

## Considerações para Proxies Reversos

O middleware utiliza `echo.Context.RealIP()` para determinar o endereço IP do cliente. Esta função é capaz de ler headers como `X-Forwarded-For`.

**Importante**: Se o sistema estiver atrás de um Proxy Reverso (Nginx, Cloudflare, AWS ALB, Traefik), certifique-se de que o proxy está configurado para encaminhar o IP real do cliente nestes headers. Caso contrário, todos os clientes podem ser identificados com o IP do proxy, causando bloqueio global indevido.

## Limitações (Dívida Técnica)

1. **Memória Distribuída**: Como o armazenamento é em memória local, em um ambiente com múltiplas instâncias (Kubernetes com N réplicas), o limite efetivo será `N * Limite Configurado` e não haverá sincronismo entre nós (ex: Redis).
2. **Persistência**: O estado dos limitadores é perdido ao reiniciar a aplicação.
