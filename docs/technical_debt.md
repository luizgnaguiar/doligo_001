# Dívidas Técnicas e Status Final

**Data**: 2026-02-05
**Status**: Fase 19.4 Concluída

---

## 1. Dívidas Técnicas Identificadas

Abaixo listamos os pontos de melhoria e dívidas técnicas conhecidas que não foram abordados no escopo atual, mas que devem ser priorizados em ciclos futuros de desenvolvimento.

### 1.1. Regras de Negócio
- **CMP em Reversões de Estoque**: Atualmente, a reversão de estoque ou produção não valida se a operação resultará em margem negativa ou inconsistência financeira profunda, apenas estorna a quantidade. Necessário implementar validação de custos no fluxo de reversão.
- **Validação Estrita de Nulos**: Reforçar a validação de `user_id` não nulo na camada de entrada (Middleware/Handler) para reduzir a dependência de "System Actions" (user_id NULL) nos logs de auditoria, garantindo que toda ação tenha um responsável humano sempre que possível.

### 1.2. Infraestrutura e Testes
- **Testes de Integração de Workers**: Aumentar a cobertura de testes automatizados focados especificamente nos cenários de falha e retry dos Workers de PDF e Email.
- **Otimização de Queries**: Revisar índices em tabelas de alto volume (`stock_ledger`, `audit_logs`) conforme o volume de dados em produção crescer.

---

## 2. Próximos Passos Sugeridos

1. **Revisão de Segurança**: Executar um pentest focado nas rotas administrativas e na validação de permissões RBAC.
2. **Dashboard Operacional**: Implementar dashboards no Grafana utilizando as queries definidas em `docs/observability.md`.
