# CareCredit Integrations Service HTTP Request Flow

Complete guide to how HTTP requests are routed from browser to `internal/server/server.go`.

## Visual Request Flow

```text
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT (Browser)                       │
│                                                                 │
│  POST https://api.getweave.com/v1/care-credit/purchases        │
│  Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...  │
│  Content-Type: application/json                                 │
│  Body: {                                                        │
│    "payment_id": "550e8400-e29b-41d4-a716-446655440000",       │
│    "merchant_number": "1234567890",                            │
│    "account_id": "9876543210",                                 │
│    ...                                                          │
│  }                                                              │
└────────────────────────┬────────────────────────────────────────┘
                         │ HTTPS
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    LAYER 1: GCP LOAD BALANCER                  │
│                                                                 │
│  Component: Google Cloud Platform Load Balancer                │
│  Managed by: GCP Infrastructure                                │
│                                                                 │
│  Responsibilities:                                              │
│  • TLS/SSL termination (HTTPS -> HTTP)                         │
│  • Geographic routing                                           │
│  • DDoS protection                                              │
│  • Health checks                                                │
│  • Routes to appropriate GKE cluster:                          │
│    - Production: wsf-prod-1-gke1-west3 (us-west3)              │
│    - Dev: wsf-dev-0-gke1-west4 (us-west4)                      │
└────────────────────────┬────────────────────────────────────────┘
                         │ HTTP (internal)
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                  LAYER 2: KUBERNETES INGRESS                   │
│                                                                 │
│  Component: GKE Ingress Controller                             │
│  Namespace: care-credit-integrations                           │
│                                                                 │
│  Responsibilities:                                              │
│  • Path-based routing                                           │
│  • Host-based routing                                           │
│  • Routes to Kubernetes Service                                 │
│                                                                 │
│  Configuration Source: .weave.yaml                              │
│    namespace: care-credit-integrations                          │
│    schemas:                                                     │
│      - path: payments-platform/care-credit-integration          │
│        public: true                                             │
└────────────────────────┬────────────────────────────────────────┘
                         │ HTTP
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                 LAYER 3: KUBERNETES SERVICE                    │
│                                                                 │
│  Component: K8s Service (ClusterIP)                            │
│  Name: care-credit-integrations                                │
│                                                                 │
│  Responsibilities:                                              │
│  • Load balancing across pods                                   │
│  • Service discovery                                            │
│  • Health check enforcement                                     │
│  • Selects one healthy pod (e.g., Pod 2 of 3)                  │
│                                                                 │
│  Load Balancing: Round-robin or least-connections              │
└────────────────────────┬────────────────────────────────────────┘
                         │ HTTP
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                      LAYER 4: POD (Container)                  │
│                                                                 │
│  Container Image: care-credit-integrations:v1.2.3              │
│  Built from: Dockerfile (scratch + binary)                     │
│  Entry Point: /care-credit-integrations                        │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ Process: daemon.Main()                                   │  │
│  │ Repo: weavelab.xyz/devx/pkg/daemon                       │  │
│  │                                                           │  │
│  │ • Initializes application lifecycle                      │  │
│  │ • Sets up signal handling (SIGTERM, SIGINT)             │  │
│  │ • Manages graceful shutdown                              │  │
│  │ • Starts HTTP Gateway on :8080                           │  │
│  │ • Starts gRPC Server on :9090 (internal only)            │  │
│  └───────────────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────────┘
                         │ HTTP to :8080
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                 LAYER 5: HTTP GATEWAY (:8080)                 │
│                                                                 │
│  Component: Auto-generated Gateway                             │
│  Repo: weavelab.xyz/schema-gen-go                              │
│  File: schemas/payments-platform/care-credit-integration/      │
│        gateway.pb.go (generated)                               │
│                                                                 │
│  Created by: carecreditintegration.NewGateway(ctx, opts...)    │
│  Source: main.go lines 91-97                                   │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ STEP 1: AUTHENTICATION                                   │  │
│  │                                                           │  │
│  │ Component: gateway.AuthClient()                          │  │
│  │ Type: wauth.KeyClient                                    │  │
│  │ Repo: weavelab.xyz/monorail/shared/wlib/wauth            │  │
│  │                                                           │  │
│  │ Process:                                                  │  │
│  │ 1. Extract "Authorization: Bearer <token>" header        │  │
│  │ 2. Validate JWT signature using public keys              │  │
│  │ 3. Verify token expiration (exp claim)                   │  │
│  │ 4. Verify issuer (iss claim)                             │  │
│  │ 5. Verify audience (aud claim)                           │  │
│  │ 6. Extract user claims (user_id, location_id, etc.)      │  │
│  │                                                           │  │
│  │ If INVALID:                                               │  │
│  │   -> Return HTTP 401 Unauthorized                         │  │
│  │   -> STOP (request never reaches handler)                 │  │
│  │                                                           │  │
│  │ If VALID:                                                 │  │
│  │   -> Continue to Step 2                                   │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ STEP 2: ROUTE MATCHING                                   │  │
│  │                                                           │  │
│  │ HTTP Router (generated from protobuf annotations)        │  │
│  │                                                           │  │
│  │ Route Table:                                              │  │
│  │ POST /v1/care-credit/purchases            -> CreatePurchase│ │
│  │ POST /v1/care-credit/refunds              -> CreateRefund │ │
│  │ POST /v1/care-credit/prefill              -> PrefillData...│ │
│  │ POST /v1/care-credit/patient-data-purl    -> GeneratePa...│ │
│  │ POST /v1/care-credit/quickscreen/offer    -> Quickscreen...││
│  │ GET  /v1/care-credit/quickscreen/offer/:id -> GetQuick... │ │
│  │ ... (more routes)                                         │  │
│  │                                                           │  │
│  │ Defined in: .proto file with google.api.http annotations │  │
│  │ Example:                                                  │  │
│  │   rpc CreatePurchase(...) returns (...) {                │  │
│  │     option (google.api.http) = {                         │  │
│  │       post: "/v1/care-credit/purchases"                  │  │
│  │       body: "*"                                           │  │
│  │     };                                                    │  │
│  │   }                                                       │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ STEP 3: REQUEST PROCESSING                               │  │
│  │                                                           │  │
│  │ 1. Parse JSON body into Go struct                        │  │
│  │    JSON -> CreatePurchaseRequest (protobuf message)      │  │
│  │                                                           │  │
│  │ 2. Validate request schema                               │  │
│  │    - Required fields present                             │  │
│  │    - Field types correct                                 │  │
│  │    - Constraints satisfied                               │  │
│  │                                                           │  │
│  │ 3. Convert to internal gRPC format (if needed)           │  │
│  │                                                           │  │
│  │ 4. Call handler method                                   │  │
│  │    -> internal/server/server.go                          │  │
│  │    -> CareCreditIntegrationServer.CreatePurchase()       │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                 │
│  Configuration:                                                 │
│  • Request timeout: 2 minutes (from main.go)                   │
│  • wgateway.WithRequestTimeoutDuration(2*time.Minute)          │
└────────────────────────┬────────────────────────────────────────┘
                         │ Function call
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│             LAYER 6: HANDLER (internal/server/server.go)       │
│                                                                 │
│  Component: CareCreditIntegrationServer                        │
│  Repo: weavelab.xyz/care-credit-integrations (THIS REPO)       │
│  File: internal/server/server.go                               │
│                                                                 │
│  Method Signature:                                              │
│  func (c *CareCreditIntegrationServer) CreatePurchase(         │
│      gctx context.Context,                                     │
│      request *carecreditintegration.CreatePurchaseRequest,     │
│  ) (*carecreditintegration.CreatePurchaseResponse, error)      │
│                                                                 │
│  Responsibilities:                                              │
│  • Business logic execution                                     │
│  • Database operations (via c.queries)                         │
│  • External API calls (to Synchrony/CareCredit)                │
│  • Error handling                                               │
│  • Response construction                                        │
│                                                                 │
│  Note: Authentication already validated by gateway             │
│        No auth checks needed in handler                         │
└────────────────────────┬────────────────────────────────────────┘
                         │ Return response
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                      RESPONSE FLOW (Reverse)                    │
│                                                                 │
│  Handler -> Gateway -> Pod -> K8s Service -> Ingress -> LB -> Client │
│                                                                 │
│  HTTP 200 OK                                                    │
│  Content-Type: application/json                                 │
│  Body: {                                                        │
│    "data": {                                                    │
│      "purchase": {                                              │
│        "status": "approved",                                    │
│        "transaction_info": { ... },                            │
│        ...                                                      │
│      }                                                          │
│    }                                                            │
│  }                                                              │
└─────────────────────────────────────────────────────────────────┘
```

## Key Components and Repositories

### 1. Infrastructure Layer
- GCP Load Balancer: Managed by Google Cloud Platform
- GKE (Kubernetes): Google Kubernetes Engine
- Configuration: `.weave.yaml` (Weave deployment config)

### 2. Application Framework
- Repo: `weavelab.xyz/devx`
- Package: `pkg/daemon`
- Purpose: Application lifecycle management
- Code: `main.go` line 24

### 3. HTTP Gateway
- Repo: `weavelab.xyz/schema-gen-go`
- Package: `pkg/wgateway`
- Generated code: `schemas/payments-platform/care-credit-integration/gateway.pb.go`
- Purpose: HTTP to handler translation, authentication, routing
- Code: `main.go` lines 91-97

### 4. Authentication
- Repo: `weavelab.xyz/monorail`
- Package: `shared/wlib/wauth`
- Interface: `wauth.KeyClient`
- Purpose: JWT token validation
- When: Before any business logic executes

### 5. Business Logic
- Repo: `weavelab.xyz/care-credit-integrations` (this repo)
- File: `internal/server/server.go`
- Type: `CareCreditIntegrationServer`
- Purpose: Core service implementation

## Authentication Details

### How Authentication Works
Client sends request with JWT token:

```http
POST /v1/care-credit/purchases HTTP/1.1
Host: api.getweave.com
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

Gateway extracts token from `Authorization` header.

`wauth.KeyClient` validates token:
- Verifies JWT signature using public keys
- Checks expiration timestamp
- Validates issuer and audience
- Extracts user claims

If invalid: returns `401 Unauthorized` immediately.

If valid: continues to route matching and handler execution.

### Where Auth Happens
- Layer: HTTP Gateway (Layer 5)
- Component: `gateway.AuthClient()`
- Code: auto-generated in `gateway.pb.go`
- Repo: `weavelab.xyz/monorail/wauth`

### What Handlers See
Handlers in `internal/server/server.go` receive already-authenticated requests.

- No need to validate auth in business logic
- Can trust that user is authenticated and authorized

## Schema-Driven Development

### How Routes Are Defined
Routes are defined in protobuf files with HTTP annotations:

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

message CreatePurchaseRequest {
  string payment_id = 1;
  string merchant_number = 2;
  string account_id = 3;
  // ... more fields
}

message CreatePurchaseResponse {
  oneof result {
    Data data = 1;
    ErrorResponse error_response = 2;
  }
}
```

### Code Generation Pipeline

```text
┌─────────────────────────────────────────────────────────────────┐
│ 1. Write .proto file                                            │
│    Repo: weavelab.xyz/schemas (or similar)                     │
│    File: payments-platform/care-credit-integration/service.proto │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Run protoc compiler with plugins                            │
│    protoc --go_out=. --go-grpc_out=. --grpc-gateway_out=. ...  │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. Generated Go code                                            │
│    Repo: weavelab.xyz/schema-gen-go                            │
│    Files:                                                       │
│    • service.pb.go (message types)                             │
│    • service_grpc.pb.go (gRPC interfaces)                      │
│    • service.pb.gw.go (HTTP gateway mappings)                  │
│    • gateway.go (NewGateway constructor)                       │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. Import in service                                            │
│    Repo: weavelab.xyz/care-credit-integrations                 │
│    File: main.go                                                │
│    Code: import carecreditintegration "weavelab.xyz/..."       │
└─────────────────────────────────────────────────────────────────┘
```

### Request and Response Bodies
Defined by: protobuf message definitions.

Serialization:
- HTTP: JSON (via `encoding/json`)
- Internal: protobuf binary (for gRPC, if used)

Example:

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
Source: `.weave.yaml`

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
┌─────────────────────────────────────────────────────────────────┐
│ Deployment: care-credit-integrations                           │
│ Replicas: 3 (example)                                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │ Pod 1        │    │ Pod 2        │    │ Pod 3        │      │
│  │              │    │              │    │              │      │
│  │ Container    │    │ Container    │    │ Container    │      │
│  │ :8080 (HTTP) │    │ :8080 (HTTP) │    │ :8080 (HTTP) │      │
│  │ :9090 (gRPC) │    │ :9090 (gRPC) │    │ :9090 (gRPC) │      │
│  └──────────────┘    └──────────────┘    └──────────────┘      │
│         ▲                   ▲                   ▲               │
└─────────┼───────────────────┼───────────────────┼───────────────┘
          │                   │                   │
          └───────────────────┴───────────────────┘
                              │
                ┌─────────────▼──────────────┐
                │ Kubernetes Service         │
                │ Type: ClusterIP            │
                │ Load balances across pods  │
                └────────────────────────────┘
```

### Shared State
OAuth tokens (for Synchrony API):
- Each pod maintains its own in-memory token cache
- Background refresher proactively refreshes tokens
- Slight staleness across pods is acceptable (tokens valid ~1 hour)

Database:
- All pods share the same CloudSQL PostgreSQL instance
- Connection pooling per pod
- Transactions ensure consistency

## Summary

### Complete HTTP Request Journey
1. Client sends HTTPS request to `api.getweave.com`.
2. GCP Load Balancer terminates TLS and routes to GKE cluster.
3. Kubernetes Ingress routes to Service based on path.
4. Kubernetes Service load balances to one of N pods.
5. HTTP Gateway (`:8080`):
   - Authenticates request via `wauth.KeyClient`
   - Matches route to handler
   - Parses JSON request body
   - Calls handler method
6. Handler (`internal/server/server.go`):
   - Executes business logic
   - Returns response
7. Response flows back through layers to client.

### Key Repositories

| Repo | Purpose |
|---|---|
| `weavelab.xyz/devx` | Application framework (`daemon`) |
| `weavelab.xyz/schema-gen-go` | Generated code from protobuf |
| `weavelab.xyz/monorail` | Shared libraries (`wauth`, `wgrpcserver`) |
| `weavelab.xyz/care-credit-integrations` | This service (business logic) |

### Authentication
- Where: HTTP Gateway layer
- Component: `wauth.KeyClient`
- Method: JWT token validation
- When: Before any business logic
- Result: Handlers receive pre-authenticated requests
