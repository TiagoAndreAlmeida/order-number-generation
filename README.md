# Gopher-ID: High-Performance ID Generation Service

Este projeto é um microserviço especializado na geração de identificadores únicos, inteiros e persistentes, projetado para cenários de altíssima concorrência. Desenvolvido em **Go**, o sistema atua como um **IDaaS (ID as a Service)** para garantir a integridade e a velocidade na criação de registros em sistemas distribuídos.

## 🚀 O Desafio Técnico

Em sistemas de pedidos em larga escala, a geração de identificadores de negócio (como o número do pedido) enfrenta desafios críticos:
1.  **Unicidade Garantida**: Impedir duplicidade em ambientes com múltiplos servidores.
2.  **Alta Performance**: A geração do ID não pode ser um gargalo (bottleneck) no fluxo de criação do pedido.
3.  **Persistência**: A sequência deve ser mantida mesmo após reinicializações do serviço.
4.  **Independência do Banco de Dados**: O ID de negócio deve ser desacoplado da chave primária (PK) técnica do banco de dados de pedidos.

## 🛠️ Arquitetura: Segmented Range Allocation

Para resolver o conflito entre **Persistência** (escrita em disco) e **Alta Performance** (memória RAM), o Gopher-ID utiliza a técnica de **Alocação por Faixas (Segments)**, implementada sobre o **BadgerDB** (um motor chave-valor embarcado de alto desempenho).

### Como o sistema funciona internamente:
*   **Leasing de Segmentos**: O microserviço reserva no BadgerDB uma "faixa" de números (ex: 1.000 a 2.000). O banco grava no disco apenas o ponto de controle (checkpoint).
*   **Gerenciamento Atômico**: Uma vez reservada, a faixa é servida diretamente da RAM. Cada requisição incrementa o contador usando instruções atômicas de hardware (`sync/atomic`), garantindo que não existam duplicatas mesmo com milhares de pedidos simultâneos.
*   **Persistência e Crash Recovery**: O BadgerDB utiliza **WAL (Write-Ahead Log)**. Se o sistema cair, os IDs não utilizados da faixa atual são descartados para garantir a segurança, e a próxima inicialização começará a partir do próximo checkpoint persistido no disco.

## 📡 Como Utilizar

O microserviço expõe uma API REST minimalista e veloz:

### Obter Próximo ID
Retorna um identificador único e sequencial.

*   **URL**: `GET /next-id`
*   **Exemplo de Resposta**:
    ```json
    {
      "id": 5001
    }
    ```

## ⚡ Vantagens Técnicas e Requisitos

Este projeto foi desenvolvido utilizando a versão **Go 1.62.2**, com foco em **Zero Dependências de infraestrutura externa** (o banco de dados é embarcado no próprio binário).

1.  **Operações Atômicas (`sync/atomic`)**: O incremento dos IDs dentro da faixa é feito via instruções atômicas de CPU, permitindo gerar milhões de IDs por segundo sem os atrasos de locks tradicionais (Mutex).
2.  **Goroutines e Pre-fetch**: Enquanto o sistema entrega os IDs da faixa atual, uma **Goroutine** em segundo plano monitora o consumo. Ao atingir um limite (ex: 80% de uso), ela busca proativamente a próxima faixa no banco.
3.  **Bufferização Dupla (Double Buffering)**: Esta estratégia garante que o consumidor final (o sistema de pedidos) sempre tenha um ID disponível na memória, eliminando a latência de rede/banco no momento crítico da transação.
4.  **Concorrência Leve**: O baixo custo de memória das Goroutines permite que o microserviço atenda simultaneamente milhares de requisições de diferentes sistemas sem degradação de performance.

## 🏗️ Estrutura de Responsabilidade

Este é um microserviço de **responsabilidade única**. Ele não processa pedidos nem gerencia regras de negócio; sua única função é fornecer identificadores sequenciais e únicos de forma confiável e ultra veloz para todo o ecossistema de serviços da empresa.

---
*Este projeto é um estudo prático sobre sistemas distribuídos, concorrência e alta disponibilidade utilizando a linguagem Go.*
