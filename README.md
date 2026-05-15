# Kube-Backlab (Dev-as-Code Laboratory)

Este repositório contém a infraestrutura e o código-fonte de um laboratório de engenharia focado em transformar o Kubernetes local (k3d) em uma plataforma de utilitários para desenvolvimento e simulação de ambientes produtivos.

## 1. Visão Geral

O **Kube-Backlab** atua como um ecossistema de desenvolvimento onde a infraestrutura é tratada como código (IaC), permitindo ciclos de feedback e observabilidade.

### Objetivos do Projeto
1.  **Automação Total:** Ciclo de Build/Deploy contínuo via Skaffold sem intervenção manual.
2.  **Persistência Resiliente:** Gestão de banco de dados PostgreSQL com volumes persistentes (PVC) via Helm.
3.  **Observabilidade Nativa:** Stack Prometheus + Grafana pré-configurada para monitoramento de saúde e performance.
4.  **Arquitetura Separada:** Separação lógica por Namespaces, isolamento de projetos e Health Checks.

## 2. Ciclo de Desenvolvimento

1.  **Orquestração:** O **k3d** cria o cluster local simulando múltiplos nós via Docker.
2.  **Ingestão:** O **Skaffold** monitora o código em `/src`, builda imagens e as injeta em namespaces isolados.
3.  **Persistência:** O **PostgreSQL** é gerenciado via Helm na namespace `infra`, garantindo que os dados sobrevivam a reinicializações.
4.  **Isolamento:** Cada projeto roda em seu próprio **Namespace**, com Ingress configurado dinamicamente para hosts `<projeto>.dev.local`.
5.  **Monitoramento:** A **Kube-Prometheus-Stack** coleta métricas de todos os pods e as projeta no Grafana.

## 3. Tecnologias e Ferramentas (Stack)

*   **Orquestrador:** k3d (Kubernetes em Docker).
*   **Backend:** Node.js 20 (Slim) com API CRUD, teste de PVC e rota de Worker.
*   **Banco de Dados:** PostgreSQL (Bitnami Helm Chart).
*   **Observabilidade:** Prometheus & Grafana (Dashboard em `grafana.dev.local`).
*   **Ingress:** Nginx Ingress Controller com suporte a TLS simulado.
*   **Automação:** Skaffold & PowerShell (Gerenciamento de Ciclo de Vida).

## 4. Estrutura do Projeto

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

## 5. Execução

### Pré-requisitos
*   **Docker Desktop**
*   **k3d, kubectl, helm e skaffold** instalados.
*   **PowerShell** (para execução dos scripts).

### Passo a Passo

1.  **Configuração de Rede:**
    Adicione ao arquivo `hosts` os projetos que deseja usar (ex: `hello.dev.local`, `projeto-a.dev.local`).

2. **Iniciar um Projeto:** 
    O laboratório suporta isolamento por nome de projeto:
    ```powershell
    # Sobe o projeto padrão no namespace 'dev'
    ./start-lab.ps1

    # Sobe um ambiente isolado em um namespace customizado
    ./start-lab.ps1 -Project "meu-app"
    ```

3.  **Gerenciar Projetos:**
    ```powershell
    # Ver o que está rodando e quais as URLs
    ./list-projects.ps1
    ```

4.  **Validação:**
    ```powershell
    ./check-lab.ps1 -HostHeader "meu-app.dev.local"
    ```

## 6. Funcionalidades de Rede e Observabilidade

### Endpoints (Porta 8080)
*   **App Principal:** `http://<projeto>.dev.local:8080`
*   **Teste de PVC:** `/pvc-test` (Contador que sobrevive ao caos)
*   **DNS Interno:** `/worker` (Comunicação entre serviços via DNS do K8s)
*   **Grafana:** [http://grafana.dev.local:8080](http://grafana.dev.local:8080)

## 7. Testes de Resiliência (Fase 4)

1. Acesse `/pvc-test` e anote o valor do contador.
2. Execute `./simulate-failure.ps1` (ele matará um pod do projeto).
3. Aguarde o Kubernetes recriar o pod e valide que o contador continua de onde parou.

## 8. Desligamento

```powershell
# Para apenas um projeto específico
./stop-lab.ps1 -Project "meu-app"

# Para desligar o cluster inteiro e todos os projetos
./stop-lab.ps1 -Full
```
