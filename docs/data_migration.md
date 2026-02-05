# Guia de Migração de Dados (Interoperabilidade)

Este documento descreve os requisitos técnicos e a estratégia para a migração de dados externos (produção) para o ecossistema Doligo. O objetivo é garantir a integridade referencial, a rastreabilidade via auditoria e a consistência dos saldos iniciais.

## 1. Ordem de Dependência (Hierarquia de Carga)

A migração deve seguir rigorosamente a ordem abaixo para respeitar as chaves estrangeiras (Foreign Keys) e a lógica de domínio:

1.  **Identity / Roles**:
    -   Tabelas: `users`, `roles`, `permissions`, `user_roles`.
    -   Nota: O usuário de sistema (ver seção Auditoria) deve ser o primeiro a ser criado.
2.  **ThirdParties**:
    -   Tabelas: `third_parties`.
    -   Dependência: `users` (criado por/editado por).
3.  **Items**:
    -   Tabelas: `items`.
    -   Dependência: `users`.
4.  **Stock (Estrutura e Saldo)**:
    -   Tabelas: `warehouses`, `bins`, `stocks`, `stock_movements`, `stock_ledger`.
    -   Dependência: `items`, `users`.

## 2. Mapeamento de UUIDs

A aplicação utiliza UUID v4 como chave primária em todas as entidades.

-   **Geração**: Recomenda-se que o processo de migração gere UUIDs determinísticos ou aleatórios (v4) antes da inserção.
-   **Integridade**: Se o sistema de origem utiliza IDs incrementais (INT), mantenha um mapeamento (De-Para) durante a execução do processo de carga para garantir que as relações entre `Items` e `Stocks`, por exemplo, sejam preservadas.
-   **Proibição**: Não é permitido desativar a validação de UUIDs ou chaves estrangeiras no banco de dados durante a carga.

## 3. Auditoria de Importação

Para distinguir dados migrados de dados criados operacionalmente, todos os registros importados devem ser associados a um usuário de sistema específico.

-   **System User ID**: `00000000-0000-0000-0000-000000000000` (ou outro UUID reservado definido na implantação).
-   **Audit Logs**: Cada inserção deve gerar um registro na tabela `audit_logs` com:
    -   `resource_name`: Nome da tabela (ex: `items`).
    -   `action`: `IMPORT` ou `MIGRATION`.
    -   `user_id`: ID do usuário de sistema.
    -   `new_values`: JSON contendo os dados importados.

## 4. Validação de Integridade e Campos Obrigatórios

Para que os módulos de faturamento e estoque funcionem corretamente, os seguintes campos são obrigatórios:

### ThirdParties
-   `name`: Nome completo ou Razão Social.
-   `email`: Deve ser único e válido.
-   `type`: Deve ser exatamente `'CUSTOMER'` ou `'SUPPLIER'`.

### Items
-   `name`: Nome descritivo do item.
-   `type`: `'STORABLE'` (para controle de estoque) ou `'SERVICE'`.
-   `cost_price` / `sale_price`: Devem ser informados (mesmo que 0.0) para cálculos de margem.

## 5. Migração de Saldo de Estoque (StockLedger)

A migração de saldos não deve ser feita apenas inserindo valores na tabela `stocks`. Para garantir a consistência do histórico de inventário:

1.  **Inserção em `stocks`**: Define a quantidade atual.
2.  **Criação de `stock_movements`**: Registre um movimento do tipo `'IN'` com a razão `'INITIAL_MIGRATION'`.
3.  **Registro em `stock_ledger`**: **OBRIGATÓRIO**. Deve ser criado um registro para cada item migrado, onde:
    -   `quantity_before`: `0.0`
    -   `quantity_change`: Quantidade migrada.
    -   `quantity_after`: Quantidade migrada.
    -   `movement_type`: `'IN'`.
    -   `reason`: `'INITIAL_MIGRATION'`.

Este procedimento garante que relatórios de movimentação histórica e auditorias de estoque sejam precisos desde o dia 1.
