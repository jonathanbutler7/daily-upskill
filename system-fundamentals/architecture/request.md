# CareCredit Integrations Service HTTP Request Flow

Complete guide to how HTTP requests are routed from browser to `internal/server/server.go`.

## Mermaid Request Flow

```mermaid
flowchart TD
  C[Client Browser\nPOST /v1/care-credit/purchases] --> LB[GCP Load Balancer\nTLS termination + routing]
  LB --> I[Kubernetes Ingress\nHost/path routing]
  I --> S[Kubernetes Service ClusterIP\nLoad balancing across pods]
  S --> P[Pod Container\ndaemon.Main\nHTTP :8080 / gRPC :9090]

  P --> A[Gateway Auth\ngateway.AuthClient + wauth.KeyClient]
  A -->|Valid JWT| R[Gateway Route Matching\nprotobuf google.api.http annotations]
  A -->|Invalid JWT| U[HTTP 401 Unauthorized]

  R --> X[Request Processing\nJSON parse + validation]
  X --> H[Handler\ninternal/server/server.go\nCareCreditIntegrationServer]
  H --> O[Response Payload]
  O --> B[Response Path Back\nGateway -> Pod -> Service -> Ingress -> LB -> Client]
```

## Visual Request Flow

```text
Client (Browser)
  POST https://api.getweave.com/v1/care-credit/purchases
  Authorization: Bearer <jwt>
  Content-Type: application/json
        |
        v
Layer 1: GCP Load Balancer
  - TLS termination
  - Geographic routing
  - DDoS protection
  - Health checks
        |
        v
Layer 2: Kubernetes Ingress
  - Host/path routing
  - Routes traffic to K8s Service
  - Config driven by .weave.yaml
        |
        v
Layer 3: Kubernetes Service (ClusterIP)
  - Service discovery
  - Load balancing across healthy pods
        |
        v
Layer 4: Pod (Container)
  - daemon.Main()
  - HTTP Gateway on :8080
  - gRPC server on :9090
        |
        v
Layer 5: HTTP Gateway (:8080)
  - Auth (gateway.AuthClient -> wauth.KeyClient)
  - Route matching from protobuf HTTP annotations
  - JSON parsing and request validation
  - Handler invocation
        |
        v
Layer 6: Handler (internal/server/server.go)
  - Business logic
  - Database operations
  - External CareCredit/Synchrony calls
  - Response construction
        |
        v
Response flows back through gateway/ingress/LB to client
```

## Layer Responsibilities

### Layer 1: GCP Load Balancer
- Component: Google Cloud Platform Load Balancer
- Managed by: GCP infrastructure
- Routes to cluster examples:
  - Production: `wsf-prod-1-gke1-west3` (`us-west3`)
  - Dev: `wsf-dev-0-gke1-west4` (`us-west4`)

### Layer 2: Kubernetes Ingress
- Component: GKE Ingress Controller
- Namespace: `care-credit-integrations`
- Source configuration (`.weave.yaml`):

```yaml
namespace: care-credit-integrations
schemas:
  - path: payments-platform/care-credit-integration
    public: true
```

### Layer 3: Kubernetes Service
- Component: Kubernetes Service (`ClusterIP`)
- Name: `care-credit-integrations`
- Provides pod selection and load balancing

### Layer 4: Pod / Container
- Container image: `care-credit-integrations:v1.2.3`
- Build style: Dockerfile with `scratch` base and compiled binary
- Process entry point: `/care-credit-integrations`
- Main process behavior (`daemon.Main()`):
  - App lifecycle initialization
  - Signal handling (`SIGTERM`, `SIGINT`)
  - Graceful shutdown
  - HTTP gateway startup on `:8080`
  - Internal gRPC server startup on `:9090`

### Layer 5: HTTP Gateway
- Component: auto-generated gateway
- Repo: `weavelab.xyz/schema-gen-go`
- Generated file: `schemas/payments-platform/care-credit-integration/gateway.pb.go`
- Instantiated by: `carecreditintegration.NewGateway(ctx, opts...)` in `main.go` (lines 91-97)

#### Step 1: Authentication
- Component: `gateway.AuthClient()`
- Type: `wauth.KeyClient`
- Repo: `weavelab.xyz/monorail/shared/wlib/wauth`

Flow:
1. Extract `Authorization: Bearer <token>` header.
2. Validate JWT signature.
3. Verify expiration, issuer, and audience claims.
4. Extract user claims.
5. If invalid, return `401 Unauthorized` and stop.
6. If valid, continue.

#### Step 2: Route Matching
- Routes are generated from protobuf `google.api.http` annotations.
- Example routes:
  - `POST /v1/care-credit/purchases` -> `CreatePurchase`
  - `POST /v1/care-credit/refunds` -> `CreateRefund`
  - `POST /v1/care-credit/prefill` -> `PrefillData...`
  - `POST /v1/care-credit/patient-data-purl` -> `GeneratePa...`
  - `POST /v1/care-credit/quickscreen/offer` -> `Quickscreen...`
  - `GET /v1/care-credit/quickscreen/offer/:id` -> `GetQuick...`

#### Step 3: Request Processing
1. Parse JSON body into Go structs/protobuf request types.
2. Validate required fields, field types, and constraints.
3. Convert to internal gRPC format when needed.
4. Invoke handler (`internal/server/server.go`).

Gateway configuration notes:
- Request timeout: 2 minutes (from `main.go`)
- Option: `wgateway.WithRequestTimeoutDuration(2*time.Minute)`

### Layer 6: Handler (`internal/server/server.go`)
- Component: `CareCreditIntegrationServer`
- Method example:

```go
func (c *CareCreditIntegrationServer) CreatePurchase(
    gctx context.Context,
    request *carecreditintegration.CreatePurchaseRequest,
) (*carecreditintegration.CreatePurchaseResponse, error)
```

- Responsibilities:
  - Business logic execution
  - Database operations (`c.queries`)
  - External API calls (Synchrony/CareCredit)
  - Error handling and response construction

Authentication has already been validated by the gateway before handler execution.

## Response Flow

```text
Handler -> Gateway -> Pod -> K8s Service -> Ingress -> Load Balancer -> Client
```

Example:
- Status: `200 OK`
- Content-Type: `application/json`
- Body: purchase payload with transaction details

## Key Components and Repositories

1. Infrastructure Layer
   - GCP Load Balancer
   - GKE (Kubernetes)
   - Deployment config in `.weave.yaml`
2. Application Framework
   - Repo: `weavelab.xyz/devx`
   - Package: `pkg/daemon`
   - Purpose: process lifecycle management
3. HTTP Gateway
   - Repo: `weavelab.xyz/schema-gen-go`
   - Package: `pkg/wgateway`
   - Purpose: auth, route mapping, HTTP-to-handler translation
4. Authentication
   - Repo: `weavelab.xyz/monorail`
   - Package: `shared/wlib/wauth`
   - Interface: `wauth.KeyClient`
5. Business Logic
   - Repo: `weavelab.xyz/care-credit-integrations`
   - File: `internal/server/server.go`

## Authentication Details

### How Authentication Works

Client request example:

```http
POST /v1/care-credit/purchases HTTP/1.1
Host: api.getweave.com
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

Token validation via `wauth.KeyClient`:
- Verify signature using public keys
- Check expiration
- Validate issuer and audience
- Extract user claims

If invalid: return `401 Unauthorized`.

If valid: continue to route matching and handler execution.

### Where Auth Happens
- Layer: HTTP Gateway (Layer 5)
- Component: `gateway.AuthClient()`
- Code location: generated gateway code (`gateway.pb.go`)

### What Handlers See
- Requests reaching `internal/server/server.go` are pre-authenticated.
- Handler code can focus on business logic.

## Schema-Driven Development

### Route Definition via Protobuf

```proto
syntax = "proto3";

package payments_platform.care_credit_integration;

import "google/api/annotations.proto";

service CareCreditIntegration {
  rpc CreatePurchase(CreatePurchaseRequest) returns (CreatePurchaseResponse) {
    option (google.api.http) = {
      post: "/v1/care-credit/purchases"
      body: "*"
    };
  }

  rpc CreateRefund(CreateRefundRequest) returns (CreateRefundResponse) {
    option (google.api.http) = {
      post: "/v1/care-credit/refunds"
      body: "*"
    };
  }
}
```

### Code Generation Pipeline

1. Author `.proto` definitions.
2. Run `protoc` with Go, gRPC, and gateway plugins.
3. Generate code in `weavelab.xyz/schema-gen-go` (`*.pb.go`, `*.pb.gw.go`, etc.).
4. Import generated packages in this service (`main.go`).

### Request and Response Bodies
- Source of truth: protobuf messages
- Serialization:
  - HTTP: JSON
  - Internal: protobuf binary (for gRPC use cases)

Request example:

```json
{
  "payment_id": "550e8400-e29b-41d4-a716-446655440000",
  "merchant_number": "1234567890",
  "account_id": "9876543210",
  "card_info": {
    "cardholder_name": "John Doe",
    "security_code": "123",
    "expiry_month": 12,
    "expiry_year": 25
  },
  "amount": 10000,
  "promo_code": "PROMO123"
}
```

Response example:

```json
{
  "data": {
    "purchase": {
      "status": "approved",
      "lender": "Synchrony",
      "merchant_fee_percentage": "3.5",
      "transaction_info": {
        "payment_transaction_id": "TXN123456",
        "authorization_code": "AUTH789"
      }
    }
  }
}
```

## Multi-Container Deployment

### Build Process

```bash
# 1. Compile Go binary
go build -o care-credit-integrations main.go

# 2. Build Docker image
docker build -t care-credit-integrations:v1.2.3 .

# 3. Push to Google Container Registry
docker tag care-credit-integrations:v1.2.3 \
  gcr.io/weave-project/care-credit-integrations:v1.2.3
docker push gcr.io/weave-project/care-credit-integrations:v1.2.3
```

### Dockerfile

```dockerfile
FROM scratch
ADD care-credit-integrations /
COPY .weave.yaml /
CMD ["/care-credit-integrations"]
```

### Deployment Configuration

```yaml
name: care-credit-integrations
namespace: care-credit-integrations

deploy:
  # Production cluster
  wsf-prod-1-gke1-west3:
    env:
      - name: ENVIRONMENT
        value: prod
      - name: CLOUDSQL_HOST
        value: wsf-prod-1:us-west3:pgsql-west3-payments-3a
      # ... more env vars

  # Dev cluster
  wsf-dev-0-gke1-west4:
    env:
      - name: ENVIRONMENT
        value: dev
      - name: CLOUDSQL_HOST
        value: wsf-dev-0:us-west4:pgsql-west4-payments-2a
      # ... more env vars
```

### Kubernetes Architecture

```text
Deployment: care-credit-integrations
Replicas: 3 (example)
Pods:
  - Pod 1: :8080 (HTTP), :9090 (gRPC)
  - Pod 2: :8080 (HTTP), :9090 (gRPC)
  - Pod 3: :8080 (HTTP), :9090 (gRPC)

Kubernetes Service (ClusterIP) load balances across pods.
```

### Shared State

OAuth tokens (Synchrony API):
- Each pod has its own in-memory token cache
- Background refreshers keep tokens fresh
- Minor cross-pod staleness is acceptable (about 1-hour validity)

Database:
- Shared CloudSQL PostgreSQL instance
- Per-pod connection pooling
- Transactional consistency

## Summary

### Complete HTTP Request Journey
1. Client sends HTTPS request to `api.getweave.com`.
2. GCP Load Balancer terminates TLS and routes to a GKE cluster.
3. Kubernetes Ingress routes by host/path to the service.
4. Kubernetes Service load balances to a healthy pod.
5. Gateway (`:8080`) authenticates, matches route, parses request, and calls handler.
6. Handler executes business logic and returns response.
7. Response travels back through gateway -> ingress -> load balancer -> client.

### Key Repositories

| Repo | Purpose |
|---|---|
| `weavelab.xyz/devx` | Application framework (`daemon`) |
| `weavelab.xyz/schema-gen-go` | Generated code from protobuf |
| `weavelab.xyz/monorail` | Shared libraries (`wauth`, `wgrpcserver`) |
| `weavelab.xyz/care-credit-integrations` | This service (business logic) |

### Authentication Quick Facts
- Where: HTTP gateway layer
- Component: `wauth.KeyClient`
- Method: JWT token validation
- Timing: before business logic
- Outcome: handlers receive pre-authenticated requests

