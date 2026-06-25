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

<img width="758" height="350" alt="image" src="https://github.com/user-attachments/assets/a0e8ab23-2ef6-4ae4-a5f2-a6ec325db8a5" />

*Demonstração da estrutura de Namespaces e Pods em execução.*

### 2.2 Persistência de Dados (PostgreSQL + PVC)
A resiliência de dados é garantida pelo uso de Persistent Volume Claims. O contador de hits mantém seu estado mesmo após a deleção ou reinicialização dos Pods de aplicação.

<img width="886" height="150" alt="image" src="https://github.com/user-attachments/assets/2db9959d-7a1b-4d0e-8969-6631f2a46949" />

*Evidência de persistência: o estado é mantido entre ciclos de vida dos Pods, a qual o numero de hits se manteve*

### 2.3 Descoberta de Serviços (DNS Interno)
A comunicação entre microsserviços utiliza a resolução de nomes nativa do Kubernetes (CoreDNS), permitindo que a aplicação principal alcance serviços secundários via endereços internos.

<img width="886" height="89" alt="image" src="https://github.com/user-attachments/assets/28795145-32dc-4151-a0e8-98712cd76b9a" />

*Comunicação service-to-service via FQDN interno.*

### 2.4 Observabilidade e Monitoramento
A saúde do sistema é monitorada via Liveness/Readiness probes, com visualização centralizada de métricas de performance no Grafana.

<img width="886" height="106" alt="image" src="https://github.com/user-attachments/assets/222011d9-65f3-438d-9db3-c6d2e02daf96" />

*Detalhamento das sondas de saúde no manifesto do Pod.*

<img width="886" height="423" alt="image" src="https://github.com/user-attachments/assets/d229f5ca-4b2c-4ee0-b062-7b1e36c54820" />

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

## 5. Motivação e Escolhas Arquiteturais (Trade-offs)

*   **k3d vs Minikube/Kind:** A escolha do **k3d** foi baseada da sua leveza e rapidez. Por rodar clusters Kubernetes dentro de containers Docker, ele consome menos recursos do Host e permite uma melhor recriação de ambientes locais se comparado a soluções baseadas em máquinas virtuais.
*   **Skaffold para Inner-loop:** Para evitar o processo manual de `docker build` + `kubectl apply`, foi adotado o **Skaffold**. Ele monitora alterações no código-fonte (`/src`), automatiza builds via multi-stage Dockerfile e injeta as mudanças no cluster.
*   **Ingress Local vs Port-Forwards:** A configuração de um Nginx Ingress Controller com DNS local simulado (ex: `hello.dev.local`) traz a experiência de acesso melhor ao ambiente, permitindo testar regras de roteamento e TLS.

## 6. Desafios Enfrentados e Soluções

*   **Persistência de Dados em Clusters Efêmeros:** 
    *   *Desafio:* Manter o estado do banco de dados intocável mesmo em meio a deleções constantes de pods durante o desenvolvimento local.
    *   *Solução:* Adoção de Helm Charts (Bitnami) com Persistent Volume Claims (PVC) mapeados, garantindo a retenção da integridade de dados e estado da aplicação entre as reinicializações.
*   **Resolução de DNS Interno (Service-to-Service):**
    *   *Desafio:* Fazer com que a aplicação principal e os workers secundários se comunicassem de forma confiável.
    *   *Solução:* Utilização do CoreDNS do Kubernetes para resolver comunicação entre serviços via FQDN (Fully Qualified Domain Name) interno, dispensando IPs estáticos e copiando uma topologia usada no mercado de microsserviços em produção.

## 7. Estrutura do Projeto

```text
kube-backlab/
├── cmd/                  # Ferramentas em Go para gestão do laboratório
│   ├── check-lab.go      # Validação de infra e testes de integração
│   ├── debug-logs.go     # Logs unificados e diagnóstico de falhas
│   ├── list-projects.go  # Lista todos os projetos ativos no cluster
│   ├── simulate-failure.go # Simula falha matando um pod aleatório
│   ├── start-lab.go      # Setup mestre com suporte a múltiplos projetos
│   └── stop-lab.go       # Shutdown seletivo por projeto ou total (-Full)
├── k8s/                  # Manifestos Kubernetes Parametrizados
│   ├── infra/            # Configurações de infraestrutura (Postgres, PVC, Values)
│   ├── worker.yaml       # Serviço secundário para testes de rede interna
│   └── ...               # ConfigMaps, Secrets, Ingress e Deployments
├── scripts/              # Scripts em PowerShell para gestão
│   ├── check-lab.ps1
│   ├── debug-logs.ps1
│   ├── list-projects.ps1
│   ├── simulate-failure.ps1
│   ├── start-lab.ps1
│   └── stop-lab.ps1
├── src/                  # Código-fonte da aplicação (Node.js)
│   ├── test/             # Scripts de teste customizados (.ps1)
│   └── app.js            # Core Backend com lógica de API e Banco
└── skaffold.yaml         # Automação de build e deploy
```

## 8. Execução

### Passo a Passo

1.  **Configuração de Rede:**
    Adicionar entrada ao arquivo `hosts` (C:\Windows\System32\drivers\etc\hosts):
    ```text
    127.0.0.1  hello.dev.local
    127.0.0.1  meu-app.dev.local
    ```

2. **Iniciar um Projeto:** 
    ```powershell
    # Opção A (Go - Recomendado para multiplataforma)
    go run ./cmd/start-lab.go -Project "Meu-App"

    # Opção B (PowerShell)
    ./scripts/start-lab.ps1 -Project "Meu-App"
    ```

3.  **Validação:**
    ```powershell
    # Opção A (Go)
    go run ./cmd/check-lab.go -Project "Meu-App"

    # Opção B (PowerShell)
    ./scripts/check-lab.ps1 -Project "Meu-App"
    ```
4. **Testes de Resiliência (Caos)**
   ```powershell
   # Opção A (Go)
    go run ./cmd/simulate-failure.go -Project "Meu-App"

    # Opção B (PowerShell)
    ./scripts/simulate-failure.ps1 -Project "Meu-App"
   ```

## 9. Desligamento

```powershell
# Opção A (Go)
go run ./cmd/stop-lab.go -Project "Meu-App"
go run ./cmd/stop-lab.go -Full

# Opção B (PowerShell)
./scripts/stop-lab.ps1 -Project "Meu-App"
./scripts/stop-lab.ps1 -Full
```
