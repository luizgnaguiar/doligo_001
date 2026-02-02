# Limites de Observabilidade do Sistema ERP/CRM (Dolibarr-Go)

Este documento descreve os limites de observabilidade do sistema ERP/CRM, conforme o estado atual de sua implementação.

## Escopo Atual de Observabilidade

O sistema implementa as seguintes capacidades de observabilidade:

*   **Logging Estruturado (9.1):** Utiliza um mecanismo para geração de logs estruturados em formato chave-valor. O `correlation_id` é injetado no contexto para correlação.
*   **Métricas Internas (9.2):** Métricas de aplicação são coletadas internamente. Não há menção de um sistema de exportação ou visualização dessas métricas para ferramentas externas.
*   **Correlação por `correlation_id` (9.3):** O `correlation_id` é utilizado para correlacionar eventos dentro de uma única transação ou requisição.

É declarado que essas capacidades são locais ao sistema. Não há observabilidade distribuída implementada.

## O que NÃO é Observável por Design

Os seguintes aspectos não são observáveis pelo design atual do sistema:

*   **Observabilidade Externa:** O sistema não expõe métricas ou logs de forma padronizada para consumo por sistemas de monitoramento externos (e.g., Prometheus, Grafana, ELK Stack).
*   **Tracing Distribuído:** Não há implementação de tracing distribuído (e.g., OpenTelemetry, Jaeger) para rastrear requisições através de múltiplos serviços ou componentes.
*   **Métricas de Negócio:** Métricas específicas de negócio (e.g., número de vendas concluídas, tempo médio de processamento de fatura) não são coletadas ou expostas de forma dedicada.
*   **SLA/SLO Definidos:** O sistema não possui Service Level Agreements (SLAs) ou Service Level Objectives (SLOs) definidos e, portanto, não há mecanismos para medir ou reportar a conformidade com eles.

## Decisões Arquiteturais Deliberadas

As seguintes decisões arquiteturais foram tomadas, impactando os limites de observabilidade:

*   **Foco em Capacidades Locais:** A implementação inicial priorizou o estabelecimento de capacidades de observabilidade internas (logs e métricas) para diagnóstico e depuração local.
*   **Ausência de Integração com Ferramentas de Mercado:** Não foi realizada integração com ferramentas de observabilidade de mercado para evitar acoplamento prematuro.
*   **Minimização de Dependências Externas:** A arquitetura busca minimizar dependências externas, o que inclui a não adoção de frameworks ou agentes de observabilidade que exijam integração profunda.

## Riscos Conhecidos Aceitos

A abordagem atual de observabilidade acarreta os seguintes riscos conhecidos, que são aceitos no estágio atual do projeto:

*   **Dificuldade de Diagnóstico em Ambiente Distribuído:** A ausência de tracing distribuído e observabilidade externa pode dificultar a identificação da causa raiz de problemas em um ambiente de produção com múltiplos serviços.
*   **Visibilidade Limitada do Comportamento do Sistema:** Sem métricas de negócio e integração com ferramentas de monitoramento, a visibilidade sobre o desempenho e a saúde geral do sistema em tempo real é limitada.
*   **Esforço Manual para Análise de Logs:** A análise de logs estruturados requer acesso direto aos logs e ferramentas de parse, sem uma interface centralizada de busca e visualização.

## Itens Explicitamente Fora de Escopo

Os seguintes itens estão explicitamente fora do escopo da fase atual de desenvolvimento no que tange à observabilidade:

*   Implementação de qualquer ferramenta ou integração de observabilidade de terceiros.
*   Definição de SLAs/SLOs ou mecanismos de reporte associados.
*   Criação de dashboards ou painéis de visualização de métricas/logs.
*   Desenvolvimento de um sistema de tracing distribuído.