package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

type session struct {
	ID        string    `json:"id"`
	Digest    string    `json:"digest"`
	Approvals int       `json:"approvals"`
	Approvers []string  `json:"approvers"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type auditLog struct {
	ID        string    `json:"id"`
	Action    string    `json:"action"`
	SessionID string    `json:"sessionId"`
	Actor     string    `json:"actor"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"createdAt"`
}

type api struct {
	mu       sync.RWMutex
	sessions map[string]*session
	audits   []auditLog
	seq      uint64
}

func main() {
	a := newAPI()
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("mpc wallet demo listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, a.routes()))
}

func newAPI() *api { return &api{sessions: make(map[string]*session)} }

func (a *api) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", a.health)
	mux.HandleFunc("/api/v1/wallets", a.wallets)
	mux.HandleFunc("/api/v1/wallets/", a.walletHandler)
	mux.HandleFunc("/api/v1/nodes", a.nodes)
	mux.HandleFunc("/api/v1/nodes/", a.nodeHandler)
	mux.HandleFunc("/api/v1/sessions", a.sessionsHandler)
	mux.HandleFunc("/api/v1/sessions/", a.sessionHandler)
	mux.HandleFunc("/api/v1/transactions", a.transactions)
	mux.HandleFunc("/api/v1/transactions/", a.transactionHandler)
	mux.HandleFunc("/api/v1/audit-logs", a.auditLogs)
	mux.HandleFunc("/api/v1/metrics", a.metrics)
	mux.HandleFunc("/api/v1/key-epochs", a.keyEpochs)
	mux.HandleFunc("/api/v1/key-epochs/", a.keyEpochHandler)
	// The frontend is deliberately dependency-free so the demo starts with one Go command.
	mux.Handle("/", http.FileServer(http.Dir("./web")))
	return allowCORS(mux)
}

func allowCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (a *api) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "protocol": "tss-lib v1.5.0", "threshold": "2-of-3 (library threshold=1)"})
}

func (a *api) wallets(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, 200, map[string]any{"items": []map[string]string{{"id": "wallet-demo-001", "address": "demo-address", "network": "testnet"}}})
		return
	}
	writeJSON(w, 405, map[string]string{"error": "method not allowed"})
}
func (a *api) walletHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, 200, map[string]string{"id": r.URL.Path[len("/api/v1/wallets/"):], "address": "demo-address", "network": "testnet"})
		return
	}
	writeJSON(w, 405, map[string]string{"error": "method not allowed"})
}

func (a *api) nodes(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, 200, map[string]any{"items": []map[string]any{{"id": "node-1", "status": "online"}, {"id": "node-2", "status": "online"}, {"id": "node-3", "status": "online"}}})
		return
	}
	writeJSON(w, 405, map[string]string{"error": "method not allowed"})
}
func (a *api) nodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, 200, map[string]string{"id": r.URL.Path[len("/api/v1/nodes/"):], "status": "online"})
		return
	}
	writeJSON(w, 405, map[string]string{"error": "method not allowed"})
}

func (a *api) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.mu.RLock()
		items := make([]*session, 0, len(a.sessions))
		for _, s := range a.sessions {
			items = append(items, s)
		}
		a.mu.RUnlock()
		writeJSON(w, 200, map[string]any{"items": items})
	case http.MethodPost:
		var input struct {
			Digest string `json:"digest"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Digest == "" {
			writeJSON(w, 400, map[string]string{"error": "digest is required"})
			return
		}
		a.mu.Lock()
		a.seq++
		id := "sign-" + time.Now().UTC().Format("20060102150405") + "-" + hex.EncodeToString([]byte{byte(a.seq)})
		now := time.Now().UTC()
		s := &session{ID: id, Digest: input.Digest, Status: "pending", CreatedAt: now}
		a.sessions[id] = s
		a.appendAudit("create_session", id, "operator", "Created signing session", now)
		a.mu.Unlock()
		writeJSON(w, 201, s)
	default:
		writeJSON(w, 405, map[string]string{"error": "method not allowed"})
	}
}

func (a *api) sessionHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/sessions/"):]
	a.mu.Lock()
	defer a.mu.Unlock()
	s, ok := a.sessions[id]
	if !ok {
		writeJSON(w, 404, map[string]string{"error": "session not found"})
		return
	}
	if r.Method == http.MethodGet {
		writeJSON(w, 200, s)
		return
	}
	if r.Method == http.MethodPost && r.URL.Query().Get("action") == "approve" {
		if s.Status != "pending" {
			writeJSON(w, 409, map[string]string{"error": "session is not pending"})
			return
		}
		nodeID := r.URL.Query().Get("node")
		if nodeID == "" {
			nodeID = "node-" + string(rune('1'+s.Approvals))
		}
		for _, approver := range s.Approvers {
			if approver == nodeID {
				writeJSON(w, 409, map[string]string{"error": "node already approved this session"})
				return
			}
		}
		s.Approvers = append(s.Approvers, nodeID)
		s.Approvals = len(s.Approvers)
		now := time.Now().UTC()
		a.appendAudit("approve_session", id, nodeID, "Approved signing session", now)
		if s.Approvals >= 2 {
			s.Status = "signed"
			a.appendAudit("sign_complete", id, "coordinator", "2-of-3 threshold reached; signature released", now)
		}
		writeJSON(w, 200, s)
		return
	}
	writeJSON(w, 405, map[string]string{"error": "method not allowed"})
}

func (a *api) transactions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		a.mu.RLock()
		items := make([]map[string]any, 0)
		for _, s := range a.sessions {
			if s.Status == "signed" {
				items = append(items, map[string]any{
					"id":        "tx-" + s.ID,
					"sessionId": s.ID,
					"digest":    s.Digest,
					"network":   "testnet",
					"status":    "signed",
					"createdAt": s.CreatedAt,
				})
			}
		}
		a.mu.RUnlock()
		sort.Slice(items, func(i, j int) bool { return items[i]["createdAt"].(time.Time).After(items[j]["createdAt"].(time.Time)) })
		writeJSON(w, 200, map[string]any{"items": items})
		return
	}
	writeJSON(w, 405, map[string]string{"error": "method not allowed"})
}
func (a *api) transactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, 200, map[string]string{"id": r.URL.Path[len("/api/v1/transactions/"):], "status": "pending"})
		return
	}
	writeJSON(w, 405, map[string]string{"error": "method not allowed"})
}
func (a *api) auditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		a.mu.RLock()
		items := append([]auditLog(nil), a.audits...)
		a.mu.RUnlock()
		sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
		writeJSON(w, 200, map[string]any{"items": items})
		return
	}
	writeJSON(w, 405, map[string]string{"error": "method not allowed"})
}
func (a *api) metrics(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	n := len(a.sessions)
	signed := 0
	for _, s := range a.sessions {
		if s.Status == "signed" {
			signed++
		}
	}
	a.mu.RUnlock()
	h := sha256.Sum256([]byte(time.Now().UTC().Format(time.RFC3339)))
	writeJSON(w, 200, map[string]any{"activeSessions": n, "signedSessions": signed, "onlineNodes": 3, "sample": hex.EncodeToString(h[:4])})
}

func (a *api) appendAudit(action, sessionID, actor, detail string, createdAt time.Time) {
	a.audits = append(a.audits, auditLog{
		ID:        "audit-" + hex.EncodeToString([]byte{byte(len(a.audits) + 1)}),
		Action:    action,
		SessionID: sessionID,
		Actor:     actor,
		Detail:    detail,
		CreatedAt: createdAt,
	})
}
func (a *api) keyEpochs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, 200, map[string]any{"items": []map[string]any{{"epoch": 1, "operation": "resharing", "publicKeyUnchanged": true}}})
		return
	}
	writeJSON(w, 405, map[string]string{"error": "method not allowed"})
}
func (a *api) keyEpochHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, 200, map[string]string{"id": r.URL.Path[len("/api/v1/key-epochs/"):], "operation": "resharing"})
		return
	}
	writeJSON(w, 405, map[string]string{"error": "method not allowed"})
}
