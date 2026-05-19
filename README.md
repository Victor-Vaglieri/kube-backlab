# Kube-Backlab (Dev-as-Code Laboratory)

Este repositório contém a infraestrutura e o código-fonte de um laboratório de engenharia focado em transformar o Kubernetes local (k3d) em uma plataforma de utilitários para desenvolvimento e simulação de ambientes produtivos.

## 1. Visão Geral

O **Kube-Backlab** atua como um ecossistema de desenvolvimento onde a infraestrutura é tratada como código (IaC), permitindo ciclos de feedback e observabilidade.

### Objetivos do Projeto
1.  **Automação Total:** Ciclo de Build/Deploy contínuo via Skaffold sem intervenção manual.
2.  **Persistência Resiliente:** Gestão de banco de dados PostgreSQL com volumes persistentes (PVC) via Helm.
3.  **Observabilidade Nativa:** Stack Prometheus + Grafana pré-configurada para monitoramento de saúde e performance.
4.  **Arquitetura Separada:** Separação lógica por Namespaces, isolamento de projetos e Health Checks.

## 2. Resultados e Evidências

Abaixo estão documentadas as evidências que comprovam a implementação dos objetivos do laboratório.

### 2.1 Automação e Isolamento (Namespaces)
O sistema realiza o deploy automatizado garantindo o isolamento lógico entre a infraestrutura de dados e a camada de aplicação através de Namespaces dedicadas.

![IMG 1](https://via.placeholder.com/800x400?text=IMG+1+-+Isolamento+de+Namespaces)
*Demonstração da estrutura de Namespaces e Pods em execução.*

### 2.2 Persistência de Dados (PostgreSQL + PVC)
A resiliência de dados é garantida pelo uso de Persistent Volume Claims. O contador de hits mantém seu estado mesmo após a deleção ou reinicialização dos Pods de aplicação.

![IMG 2](https://via.placeholder.com/800x400?text=IMG+2+-+Persistência+de+Dados)
*Evidência de persistência: o estado é mantido entre ciclos de vida dos Pods.*

### 2.3 Descoberta de Serviços (DNS Interno)
A comunicação entre microsserviços utiliza a resolução de nomes nativa do Kubernetes (CoreDNS), permitindo que a aplicação principal alcance serviços secundários via endereços internos.

![IMG 3](https://via.placeholder.com/800x400?text=IMG+3+-+Resolução+DNS+Interno)
*Comunicação service-to-service via FQDN interno.*

### 2.4 Observabilidade e Monitoramento
A saúde do sistema é monitorada via Liveness/Readiness probes, com visualização centralizada de métricas de performance no Grafana.

![IMG 4](https://via.placeholder.com/800x400?text=IMG+4+-+Configuração+de+Health+Checks)
*Detalhamento das sondas de saúde no manifesto do Pod.*

![IMG 5](https://via.placeholder.com/800x400?text=IMG+5+-+Dashboard+Grafana)
*Monitoramento de recursos computacionais em tempo real.*

## 3. Ciclo de Desenvolvimento

1.  **Orquestração:** O **k3d** cria o cluster local simulando múltiplos nós via Docker.
2.  **Ingestão:** O **Skaffold** monitora o código em `/src`, builda imagens e as injeta em namespaces isolados.
3.  **Persistência:** O **PostgreSQL** é gerenciado via Helm na namespace `infra`, garantindo que os dados sobrevivam a reinicializações.
4.  **Isolamento:** Cada projeto roda em seu próprio **Namespace**, com Ingress configurado dinamicamente para hosts `<projeto>.dev.local`.
5.  **Monitoramento:** A **Kube-Prometheus-Stack** coleta métricas de todos os pods e as projeta no Grafana.

## 4. Tecnologias e Ferramentas (Stack)

*   **Orquestrador:** k3d (Kubernetes em Docker).
*   **Backend:** Node.js 20 (Slim) com API CRUD, teste de PVC e rota de Worker.
*   **Banco de Dados:** PostgreSQL (Bitnami Helm Chart).
*   **Observabilidade:** Prometheus & Grafana (Dashboard em `grafana.dev.local`).
*   **Ingress:** Nginx Ingress Controller com suporte a TLS simulado.
*   **Automação:** Skaffold & PowerShell (Gerenciamento de Ciclo de Vida).

## 5. Estrutura do Projeto

```text
kube-backlab/
├── k8s/                  # Manifestos Kubernetes Parametrizados
│   ├── infra/            # Configurações de infraestrutura (Postgres, PVC, Values)
│   ├── worker.yaml       # Serviço secundário para testes de rede interna
│   └── ...               # ConfigMaps, Secrets, Ingress e Deployments
├── src/                  # Código-fonte da aplicação (Node.js)
│   ├── test/             # Scripts de teste customizados (.ps1)
│   └── app.js            # Core Backend com lógica de API e Banco
├── check-lab.ps1         # Validação de infra e testes de integração
├── debug-logs.ps1        # Logs unificados e diagnóstico de falhas
├── list-projects.ps1     # Lista todos os projetos ativos no cluster
├── simulate-failure.ps1  # Simula falha matando um pod aleatório
├── start-lab.ps1         # Setup mestre com suporte a múltiplos projetos
├── stop-lab.ps1          # Shutdown seletivo por projeto ou total (-Full)
└── skaffold.yaml         # Automação de build e deploy
```

## 6. Execução

### Passo a Passo

1.  **Configuração de Rede:**
    Adicionar entrada ao arquivo `hosts` (C:\Windows\System32\drivers\etc\hosts):
    ```text
    127.0.0.1  hello.dev.local
    127.0.0.1  meu-app.dev.local
    ```

2. **Iniciar um Projeto:** 
    ```powershell
    go run start-lab.go -Project "Meu-App"
    ```

3.  **Validação:**
    ```powershell
    go run check-lab.go -Project "Meu-App"
    ```

## 7. Desligamento

```powershell
go run stop-lab.go -Project "meu-app"
go run stop-lab.go -Full
```
