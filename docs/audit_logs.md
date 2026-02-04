# Auditoria de Domínio (Audit Logs)

**Versão**: 1.0.0  
**Data**: 2026-02-04
**Escopo**: Auditoria de alterações em entidades de domínio

---

## 1. Estrutura da Tabela `audit_logs`

O sistema rastreia alterações críticas em entidades de domínio (como Itens, Estoque, Faturas) na tabela `audit_logs`.

| Coluna | Tipo | Descrição |
| :--- | :--- | :--- |
| `id` | UUID | Identificador único do log de auditoria. |
| `timestamp` | TIMESTAMPTZ | Data e hora em que a ação ocorreu. |
| `user_id` | UUID | Identificador do usuário que realizou a ação. |
| `resource_name` | VARCHAR | Nome da entidade (ex: `items`, `invoices`). |
| `resource_id` | VARCHAR | Identificador único da instância da entidade. |
| `action` | VARCHAR | Ação realizada (`create`, `update`, `delete`). |
| `old_values` | JSONB | Estado anterior da entidade (apenas para `update` e `delete`). |
| `new_values` | JSONB | Novo estado da entidade (apenas para `create` e `update`). |
| `correlation_id` | VARCHAR | ID único da requisição para vincular logs de texto e banco. |

---

## 2. Rastreabilidade (Traceability)

Para uma auditoria completa, o sistema utiliza o `correlation_id` (também conhecido como `request_id` nos logs de aplicação).

### 2.1. Vinculação de Logs

Ao investigar um incidente, siga estes passos:

1.  **Localize a entrada no banco**:
    ```sql
    SELECT * FROM audit_logs WHERE resource_id = 'uuid-da-entidade';
    ```
2.  **Identifique o `correlation_id`**: Obtenha o valor da coluna `correlation_id` da linha encontrada.
3.  **Busque nos logs de aplicação**:
    Utilize ferramentas como `grep`, `Kibana` ou `CloudWatch Logs` para buscar por esse ID.
    ```bash
    grep "correlation_id=valor-obtido" app.log
    ```

Isso permite ver tanto o log estruturado do banco (quem mudou o quê) quanto o contexto técnico da execução (erros de rede, queries SQL intermediárias, latência).

---

## 3. Implementação Técnica

A auditoria é orquestrada pelo `AuditService` (`internal/usecase/audit_service.go`) e persistida pelo `AuditRepository`.

### 3.1. Exemplo de Uso no Usecase

```go
func (u *itemUsecase) Update(ctx context.Context, item *domain.Item) error {
    oldItem, _ := u.repo.GetByID(ctx, item.ID)
    
    if err := u.repo.Update(ctx, item); err != nil {
        return err
    }

    return u.auditService.Log(ctx, "items", item.ID.String(), "update", oldItem, item)
}
```

O `AuditService` extrai automaticamente o `user_id` e o `correlation_id` do `context.Context`, garantindo que o rastro seja mantido sem poluir a assinatura dos métodos de negócio.
