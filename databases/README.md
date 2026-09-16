# Database migrations

`migrations/` is the schema source of truth. On `httpServer` startup, `cmd/http_server.go`
opens MySQL when `MYSQL_DSN` is configured and applies each `*.up.sql` file once,
recording its file name in `schema_migrations`.

The current migration creates wallet metadata, share references, participants,
key epochs, transactions, signing sessions, approvals, audit logs and node
heartbeats. No private key share is stored in these tables; only a share
reference and fingerprint are persisted.
