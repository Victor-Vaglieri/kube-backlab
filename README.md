# Kube-Backlab (Dev-as-Code Laboratory)

Este repositório contém a infraestrutura e o código-fonte de um laboratório de engenharia focado em transformar o Kubernetes local (k3d) em uma plataforma de utilitários para desenvolvimento e simulação de ambientes produtivos.

## 1. Visão Geral

O **Kube-Backlab** atua como um ecossistema de desenvolvimento onde a infraestrutura é tratada como código (IaC), permitindo ciclos de feedback e observabilidade.

### Objetivos do Projeto
1.  **Automação Total:** Ciclo de Build/Deploy contínuo via Skaffold sem intervenção manual.
2.  **Persistência Resiliente:** Gestão de banco de dados PostgreSQL com volumes persistentes (PVC) via Helm.
3.  **Observabilidade Nativa:** Stack Prometheus + Grafana pré-configurada para monitoramento de saúde e performance.
4.  **Arquitetura Profissional:** Separação lógica por Namespaces, uso de Secrets/ConfigMaps e Health Checks (Liveness/Readiness).

## 2. Ciclo de Desenvolvimento

1.  **Orquestração:** O **k3d** cria o cluster local simulando múltiplos nós via Docker.
2.  **Ingestão:** O **Skaffold** monitora o código em `/src`, builda imagens multi-stage e as injeta no cluster.
3.  **Persistência:** O **PostgreSQL** é gerenciado via Helm na namespace `infra`, garantindo que os dados sobrevivam a reinicializações.
4.  **Roteamento:** O **Nginx Ingress Controller** gerencia o tráfego externo através de hosts virtuais (`.dev.local`).
5.  **Monitoramento:** A **Kube-Prometheus-Stack** coleta métricas de todos os pods e as projeta no Grafana.

## 3. Tecnologias e Ferramentas (Stack)

*   **Orquestrador:** k3d (Kubernetes em Docker).
*   **Backend:** Node.js 20 (Slim) com API CRUD e Driver `pg`.
*   **Banco de Dados:** PostgreSQL (Bitnami Helm Chart).
*   **Observabilidade:** Prometheus & Grafana (Dashboard em `grafana.dev.local`).
*   **Ingress:** Nginx Ingress Controller.
*   **Automação:** Skaffold & PowerShell (Orquestrador de Testes).

## 4. Estrutura do Projeto

```text
kube-backlab/
├── k8s/                  # Manifestos Kubernetes (Deployments, Services, Ingress)
│   ├── infra/            # Configurações de infraestrutura (Postgres, PVC, Values)
│   └── ...               # ConfigMaps e Secrets da aplicação
├── src/                  # Código-fonte da aplicação (Node.js)
│   ├── test/             # Scripts de teste customizados (.ps1)
│   └── app.js            # Core Backend com lógica de API e Banco
├── check-lab.ps1         # Orquestrador de validação de infra e testes
└── skaffold.yaml         # Configuração mestre de automação e deploy
```

## 5. Execução

### Pré-requisitos
*   **Docker Desktop**
*   **k3d, kubectl, helm e skaffold** instalados.
*   **PowerShell** (para execução dos scripts de validação).

### Passo a Passo

1.  **Configuração de Rede:**
    Adicione ao seu arquivo `hosts` (`C:\Windows\System32\drivers\etc\hosts`):
    ```text
    127.0.0.1 hello.dev.local
    127.0.0.1 grafana.dev.local
    ```

2. **Build & Up:** 
    Inicie o ambiente completo com um único comando. O laboratório aceita o caminho do projeto como argumento opcional:
    ```powershell
    # Para rodar o projeto padrão (pasta /src)
    ./start-lab.ps1

    # Para rodar um projeto externo (ex: seu projeto real)
    ./start-lab.ps1 "C:\Caminho\Para\Seu\Projeto"
    ```

3.  **Configuração do Projeto Externo:**
    Se estiver testando um projeto externo, copie o arquivo `k8s-project-template.yaml` para a raiz do seu projeto como `k8s-config.yaml` para definir variáveis de ambiente e segredos personalizados.

4.  **Validação Profissional:**
    Em um novo terminal, execute o orquestrador de testes:
    ```powershell
    ./check-lab.ps1
    ```

## 6. Funcionalidades de Observabilidade
Acesse o **Grafana** para monitorar o sistema:
- **URL:** [http://grafana.dev.local:8080](http://grafana.dev.local:8080)
- **Credenciais:** `admin` / `admin123`
- **Dashboards:** Navegue em `Dashboards > Browse > Default` para visualizar métricas de CPU, Memória e Rede dos Pods.
