# MPC Wallet System

This repository is the runnable delivery for the TSS wallet thesis project. It provides a Gin REST API, an administration console, MySQL schema migrations, Redis Stream consumers, and 2-of-3 threshold-signature parameter and protocol tests.

## Run locally

The service starts without MySQL or Redis in demonstration mode. It retains wallet, transaction, signing-session, approval, participant, heartbeat, reshare-request, audit, and metrics APIs in memory.

```bash
go test ./... -count=1
go run ./apps/tss-wallet-service/cmd httpServer
```

Open `http://127.0.0.1:8080` to use the administration console. The frontend is a static application in `web/admin-console`; no Node.js installation is required.

To enable persistent schema storage and asynchronous session events, configure `MYSQL_DSN` and `REDIS_ADDR` in `.env.example` and run the corresponding processes:

```bash
go run ./apps/tss-wallet-service/cmd httpServer
go run ./apps/tss-wallet-service/cmd startConsumers
go run ./apps/tss-wallet-service/cmd cron
```

With Docker available, start the full dependency set with:

```bash
docker compose -f deployments/docker/docker-compose.yml up --build
```

## Directory layout

```text
apps/tss-wallet-service/
  cmd/                         executable entry point and Cobra commands
  internal/
    application/               wallet and signing-session use cases
    domain/                    domain data models
    transport/httpapi/         Gin controllers, routes, request/response DTOs
    transport/                 HTTP server lifecycle
    worker/consumers/          Redis Stream event consumer
    worker/cron/               scheduled operational task runner
    infrastructure/redis/      service-specific Redis event publisher
    config/                    environment configuration
internal/database/
  mysql/                       MySQL connection pool
  redis/                       shared Redis client setup
migrations/wallet/             versioned MySQL schema and migration runner
signer/tss-common/             threshold-signature configuration and protocol tests
api/openapi/                   API specification
web/admin-console/             static administration console
deployments/docker/            Dockerfile and Compose environment
docs/test-reports/             test cases and historical run logs
```

## API overview

| Route | Method | Purpose |
|---|---|---|
| `/api/v1/health` | GET | Service health and threshold parameters |
| `/api/v1/wallets` | GET/POST | List or create wallet metadata |
| `/api/v1/wallets/:id/addresses` | GET | Get wallet addresses |
| `/api/v1/wallets/:id/balance` | GET | Get demonstration balance |
| `/api/v1/transactions` | GET/POST | List or create transfers |
| `/api/v1/transactions/:id` | GET | Get transfer details |
| `/api/v1/sign-sessions` | GET/POST | List or create signing sessions |
| `/api/v1/sign-sessions/:id/approve` | POST | Approve a session as a participant |
| `/api/v1/sign-sessions/:id/cancel` | POST | Cancel a signing session |
| `/api/v1/participants` | GET | List participant nodes |
| `/api/v1/participants/:id/refresh` | POST | Create a reshare request |
| `/api/v1/key-epochs` | GET | List key epochs |
| `/api/v1/key-epochs/:id` | GET | Get key epoch detail |
| `/api/v1/nodes/heartbeat` | POST | Report participant heartbeat |
| `/api/v1/audit-logs` | GET | Read audit entries |
| `/api/v1/system/metrics` | GET | Read operational metrics |

## Threshold-signature boundary

`go.mod` pins `github.com/bnb-chain/tss-lib v1.5.0`. `signer/tss-common/parameters` verifies that three participants with library parameter `threshold=1` implement a 2-of-3 signing policy: the library threshold is the polynomial degree, so `threshold + 1` participants are required.

The protocol tests use the library's `ecdsa/keygen` and `ecdsa/signing` message state machines. The API service demonstrates session orchestration and approval rules; it does not persist private key shares. MySQL stores only share references and fingerprints.
