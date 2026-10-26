# Architecture — rate-limiter-service
> Last updated: 2026-08-29 | Maturity: Full Prototype
> _API gateway rate limiting service using Redis Lua scripting._

## System Diagram
The following Mermaid.js sequence diagram maps the core workflow and interactions:

```mermaid
flowchart TD
    Client(["Client"])
    Gateway["Gateway Server (Go)"]
    Redis[("Redis")]
    Upstream["Upstream API (Simulated)"]

    Client -->|"HTTP Request"| Gateway
    Gateway -->|"EVAL Lua Script (Token Bucket)"| Redis
    Redis -->|"Allowed / Denied"| Gateway
    Gateway -.->|"If Allowed"| Upstream
    Upstream -.->|"Response"| Gateway
    Gateway -->|"429 Too Many Requests (If Denied)"| Client
```

## Component Table

| Component | File | Responsibility | Tech |
|---|---|---|---|
| API Server | `main.go` | Express server intercepting requests | Go |
| Rate Limiter | `middleware/limiter.go` | Middleware enforcing limits | Go |
| Lua Scripts | `middleware/redis.go` | Atomic token bucket logic | Lua/Redis |

## Port Assignments

| Service | Port | Notes |
|---|---|---|
| Gateway API | `8080` | Main entrypoint |
| Redis | `6379` | Backing store |

## Dependency Honesty Table

| Dependency | Status | Notes |
|---|---|---|
| Redis | **Real** | Required for atomic state tracking. |
| Upstream Service | **Simulated** | Returns a 200 OK immediately for demonstration purposes. |

## Progressive Credential Loading & Production Architecture (v0.2.0)

This service implements the **Progressive Credential Loading (PCL)** pattern, decoupling development speed from external infrastructure provisioning:

```mermaid
graph LR
    Env[Environment Credentials] -->|Detect Keys & URLs| Detector[Credential Detector]
    Detector -->|No Credentials / Missing| Mock[Mock Mode Factory]
    Detector -->|Valid Credentials Present| Live[Live Provider Factory]
    Mock --> Runtime[Application Core & /health]
    Live --> Runtime
```

### Core Architecture Components
1. **Mode Detection Singleton:** Inspects environment configuration at boot time without hard crashes on missing secrets.
2. **Dual-Mode Factory Wrappers:**
   - **Storage / Cache:** In-process maps or SQLite `:memory:` transparently promote to managed Redis or PostgreSQL when connection strings are supplied.
   - **AI Providers:** Deterministic heuristic generators and local embeddings transparently promote to live OpenAI, Anthropic, or Gemini API clients.
3. **Hot-Reload & Dynamic TLS:** File watchers monitor mounted volume paths (Kubernetes ConfigMaps and Secrets) to update certificates and limit configs atomically without process restarts.
4. **Health & Observability:** The `/health` endpoint exposes real-time runtime mode metadata (`mock` vs `live`) across all backing adapters.

## Declarative GitOps & Multi-Cluster Deployment Architecture (v0.2.1)

This service participates in the portfolio-wide **ArgoCD Hub-and-Spoke GitOps continuous delivery catalog**:

```mermaid
graph TD
    Hub[GitOps Hub: multi-cluster-gitops-promotion] -->|AppSet Matrix: dev| DevCluster[Dev Cluster: kubernetes.default]
    Hub -->|AppSet Matrix: staging| StagingCluster[Staging Cluster: staging-k8s:6443]
    Hub -->|AppSet Matrix: prod| ProdCluster[Prod Cluster: prod-k8s:6443]
    
    subgraph WorkloadRepo["Workload Repo: rate-limiter-service"]
        Base[deploy/k8s/base] --> DevOverlay[deploy/k8s/overlays/dev]
        Base --> StagingOverlay[deploy/k8s/overlays/staging]
        Base --> ProdOverlay[deploy/k8s/overlays/prod]
    end
    
    DevOverlay -.->|Continuous Reconciliation| DevCluster
    StagingOverlay -.->|Continuous Reconciliation| StagingCluster
    ProdOverlay -.->|Gated PR Promotion| ProdCluster
```

### GitOps Topology & Separation of Concerns
1. **Base Manifests (`deploy/k8s/base`):** Declares non-root PodSecurity standards (`runAsUser: 10001`, `drop: [ALL]`, `readOnlyRootFilesystem: true`), readiness/liveness HTTP probes, and CPU/memory requests and limits without hardcoding namespace.
2. **Environment Overlays (`deploy/k8s/overlays/`):**
   - `dev/`: 1 replica, `300m/256Mi` limit, automated drift self-healing.
   - `staging/`: 2 replicas, `500m/512Mi` limit, automated drift self-healing.
   - `prod/`: 3 replicas, `1000m/1Gi` limit, gated pull-request review.
3. **ArgoCD Declarations (`deploy/argocd/`):** Houses standalone `Application` CRDs (`application.yaml`, `application-dev.yaml`, `application-prod.yaml`) and values overrides for OCI Helm and Kustomize reconciliation.
