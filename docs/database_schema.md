# Dicionário de Dados e Esquema

**Versão**: 1.0.0  
**Data**: 2026-02-05  
**Motor**: PostgreSQL

---

## 1. Visão Geral

O banco de dados utiliza UUIDs (v4) para todas as chaves primárias e `TIMESTAMPTZ` para campos temporais. A integridade referencial é aplicada rigorosamente via Foreign Keys.

---

## 2. Tabelas e Relacionamentos

### 2.1. Identidade e Acesso (Identity)

| Tabela | PK | Descrição | Relacionamentos Chave |
| :--- | :--- | :--- | :--- |
| `users` | `id` | Usuários do sistema. | 1:N com `audit_logs`, `invoices`, etc. |
| `roles` | `id` | Papéis de acesso (ex: ADMIN). | N:N com `users` via `user_roles`. |
| `permissions` | `id` | Permissões granulares. | N:N com `roles` via `role_permissions`. |
| `user_roles` | (`user_id`, `role_id`) | Tabela associativa. | `ON DELETE CASCADE`. |
| `role_permissions` | (`role_id`, `permission_id`) | Tabela associativa. | `ON DELETE CASCADE`. |

### 2.2. Núcleo (Core)

| Tabela | PK | Descrição | Relacionamentos Chave |
| :--- | :--- | :--- | :--- |
| `third_parties` | `id` | Clientes e Fornecedores. | Usado em `invoices`. |
| `items` | `id` | Produtos e Serviços. | Usado em `stocks`, `invoice_lines`, `bom`. |

### 2.3. Estoque (Inventory)

| Tabela | PK | Descrição | Relacionamentos Chave |
| :--- | :--- | :--- | :--- |
| `warehouses` | `id` | Armazéns físicos. | 1:N com `bins`, `stocks`. |
| `bins` | `id` | Localizações dentro do armazém. | `UNIQUE(warehouse_id, name)`. |
| `stocks` | (`item_id`, `warehouse_id`, `bin_id`) | Quantidade atual (Snapshot). | FKs para `items`, `warehouses`, `bins`. |
| `stock_movements` | `id` | Registro volátil de movimento. | Base para o `stock_ledger`. |
| `stock_ledger` | `id` | Histórico imutável (Audit Trail). | Rastreabilidade total de estoque. |

### 2.4. Manufatura (BOM)

| Tabela | PK | Descrição | Relacionamentos Chave |
| :--- | :--- | :--- | :--- |
| `bill_of_materials` | `id` | Cabeçalho da Lista Técnica. | 1:1 com `items` (Produto final). |
| `bill_of_materials_components` | `id` | Componentes da receita. | N:1 com `bill_of_materials`, `items`. |
| `production_records` | `id` | Registro de produção realizada. | Vincula BOM, Produto e Armazém. |

### 2.5. Faturamento (Billing)

| Tabela | PK | Descrição | Relacionamentos Chave |
| :--- | :--- | :--- | :--- |
| `invoices` | `id` | Cabeçalho da Fatura. | N:1 com `third_parties`. |
| `invoice_lines` | `id` | Itens da Fatura. | N:1 com `invoices`, `items`. |

### 2.6. Sistema

| Tabela | PK | Descrição | Relacionamentos Chave |
| :--- | :--- | :--- | :--- |
| `audit_logs` | `id` | Log de Auditoria Técnica/Negócio. | `user_id` (Nullable), `correlation_id`. |

---

## 3. Notas de Integridade

- **Soft Deletes**: Tabelas principais (`users`, `items`, `invoices`, etc.) possuem `deleted_at`. O sistema deve filtrar registros onde `deleted_at IS NOT NULL`.
- **Prevenção de Orfãos**:
    - Tabelas associativas (`user_roles`) usam `ON DELETE CASCADE`.
    - Tabelas transacionais (`invoices`, `stock_ledger`) usam `ON DELETE RESTRICT` para impedir a exclusão de dados mestres que possuem histórico.
