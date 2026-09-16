package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type response struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func request(t *testing.T, h http.Handler, method, path, body string) response {
	t.Helper()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("%s %s returned HTTP %d", method, path, recorder.Code)
	}
	var result response
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func dataID(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var value struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	if value.ID == "" {
		t.Fatal("response data did not include id")
	}
	return value.ID
}

func TestWalletTransactionAndSigningWorkflow(t *testing.T) {
	h := NewServer("../web")
	if got := request(t, h, http.MethodGet, "/api/v1/health", ""); got.Code != 0 {
		t.Fatalf("health code = %d", got.Code)
	}
	if got := request(t, h, http.MethodGet, "/api/v1/wallets", ""); got.Code != 0 {
		t.Fatalf("wallet list code = %d", got.Code)
	}

	tx := request(t, h, http.MethodPost, "/api/v1/transactions", `{"wallet_id":"wallet-demo-001","to_address":"test-recipient","amount":"1.25","digest":"abc123"}`)
	if tx.Code != 0 {
		t.Fatalf("transaction code = %d: %s", tx.Code, tx.Message)
	}
	txID := dataID(t, tx.Data)
	session := request(t, h, http.MethodPost, "/api/v1/sign-sessions", `{"wallet_id":"wallet-demo-001","transaction_id":"`+txID+`","digest":"abc123"}`)
	if session.Code != 0 {
		t.Fatalf("session code = %d: %s", session.Code, session.Message)
	}
	sessionID := dataID(t, session.Data)
	if got := request(t, h, http.MethodPost, "/api/v1/sign-sessions/"+sessionID+"/approve", `{"node_id":"node-1"}`); got.Code != 0 {
		t.Fatalf("first approval code = %d", got.Code)
	}
	if got := request(t, h, http.MethodPost, "/api/v1/sign-sessions/"+sessionID+"/approve", `{"node_id":"node-2"}`); got.Code != 0 {
		t.Fatalf("second approval code = %d", got.Code)
	}
	if got := request(t, h, http.MethodPost, "/api/v1/sign-sessions/"+sessionID+"/approve", `{"node_id":"node-2"}`); got.Code != 1009 {
		t.Fatalf("duplicate approval code = %d", got.Code)
	}
	if got := request(t, h, http.MethodGet, "/api/v1/transactions/"+txID, ""); got.Code != 0 {
		t.Fatalf("transaction detail code = %d", got.Code)
	}
	if got := request(t, h, http.MethodGet, "/api/v1/audit-logs", ""); got.Code != 0 {
		t.Fatalf("audit code = %d", got.Code)
	}
}

func TestValidationAndOperationalInterfaces(t *testing.T) {
	h := NewServer("../web")
	if got := request(t, h, http.MethodPost, "/api/v1/sign-sessions", `{}`); got.Code != 1001 {
		t.Fatalf("invalid session code = %d", got.Code)
	}
	if got := request(t, h, http.MethodGet, "/api/v1/wallets/missing", ""); got.Code != 1004 {
		t.Fatalf("missing wallet code = %d", got.Code)
	}
	if got := request(t, h, http.MethodGet, "/api/v1/participants", ""); got.Code != 0 {
		t.Fatalf("participant list code = %d", got.Code)
	}
	if got := request(t, h, http.MethodPost, "/api/v1/nodes/heartbeat", `{"node_id":"node-3","version":"tss-agent/1.5.0"}`); got.Code != 0 {
		t.Fatalf("heartbeat code = %d", got.Code)
	}
	if got := request(t, h, http.MethodPost, "/api/v1/participants/node-3/refresh", `{}`); got.Code != 0 {
		t.Fatalf("refresh code = %d", got.Code)
	}
	if got := request(t, h, http.MethodGet, "/api/v1/system/metrics", ""); got.Code != 0 {
		t.Fatalf("metrics code = %d", got.Code)
	}
}
