package httpapi

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

// TestTwentyFunctionalChecks executes the twenty numbered functional checks
// described in the thesis test-case table. Each subtest maps to T01..T20 and
// exercises the Gin route entry point so the response envelope, validation and
// approval state machine are covered end to end.
func TestTwentyFunctionalChecks(t *testing.T) {
	h := NewRouter("../../../../../web/admin-console")

	check := func(name, method, path, body string, wantCode int) response {
		t.Helper()
		got := request(t, h, method, path, body)
		if got.Code != wantCode {
			t.Fatalf("%s: expected code %d, got %d (%s)", name, wantCode, got.Code, got.Message)
		}
		return got
	}

	t.Run("T01 health check", func(t *testing.T) {
		got := check("T01", http.MethodGet, "/api/v1/health", "", 0)
		var data struct {
			Threshold string `json:"threshold"`
		}
		if err := json.Unmarshal(got.Data, &data); err != nil {
			t.Fatal(err)
		}
		if data.Threshold != "2-of-3 (library threshold=1)" {
			t.Fatalf("unexpected threshold: %s", data.Threshold)
		}
	})

	t.Run("T02 wallet list", func(t *testing.T) {
		check("T02", http.MethodGet, "/api/v1/wallets", "", 0)
	})

	t.Run("T03 wallet detail", func(t *testing.T) {
		check("T03", http.MethodGet, "/api/v1/wallets/wallet-demo-001", "", 0)
	})

	t.Run("T04 wallet addresses", func(t *testing.T) {
		check("T04", http.MethodGet, "/api/v1/wallets/wallet-demo-001/addresses", "", 0)
	})

	t.Run("T05 wallet balance", func(t *testing.T) {
		check("T05", http.MethodGet, "/api/v1/wallets/wallet-demo-001/balance", "", 0)
	})

	t.Run("T06 participant list", func(t *testing.T) {
		check("T06", http.MethodGet, "/api/v1/participants", "", 0)
	})

	t.Run("T07 key epoch list", func(t *testing.T) {
		check("T07", http.MethodGet, "/api/v1/key-epochs", "", 0)
	})

	t.Run("T08 key epoch detail", func(t *testing.T) {
		check("T08", http.MethodGet, "/api/v1/key-epochs/1", "", 0)
	})

	t.Run("T09 audit log list", func(t *testing.T) {
		check("T09", http.MethodGet, "/api/v1/audit-logs", "", 0)
	})

	t.Run("T10 metrics", func(t *testing.T) {
		check("T10", http.MethodGet, "/api/v1/system/metrics", "", 0)
	})

	t.Run("T11 node heartbeat", func(t *testing.T) {
		check("T11", http.MethodPost, "/api/v1/nodes/heartbeat", `{"node_id":"node-3","version":"tss-agent/1.5.0"}`, 0)
	})

	var txID string
	t.Run("T12 create transaction", func(t *testing.T) {
		got := check("T12", http.MethodPost, "/api/v1/transactions", `{"wallet_id":"wallet-demo-001","to_address":"test-recipient","amount":"1.25","digest":"abc123"}`, 0)
		txID = dataID(t, got.Data)
	})

	t.Run("T13 transaction detail", func(t *testing.T) {
		if txID == "" {
			t.Fatal("transaction id was not captured")
		}
		check("T13", http.MethodGet, "/api/v1/transactions/"+txID, "", 0)
	})

	var sessionID string
	t.Run("T14 create signing session", func(t *testing.T) {
		got := check("T14", http.MethodPost, "/api/v1/sign-sessions", `{"wallet_id":"wallet-demo-001","transaction_id":"`+txID+`","digest":"abc123"}`, 0)
		sessionID = dataID(t, got.Data)
	})

	t.Run("T15 first approval", func(t *testing.T) {
		got := check("T15", http.MethodPost, "/api/v1/sign-sessions/"+sessionID+"/approve", `{"node_id":"node-1"}`, 0)
		var data struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(got.Data, &data); err != nil {
			t.Fatal(err)
		}
		if data.Status != "signing" {
			t.Fatalf("unexpected status after first approval: %s", data.Status)
		}
	})

	t.Run("T16 second approval completes signing", func(t *testing.T) {
		got := check("T16", http.MethodPost, "/api/v1/sign-sessions/"+sessionID+"/approve", `{"node_id":"node-2"}`, 0)
		var data struct {
			Status            string `json:"status"`
			SignatureVerified bool   `json:"signature_verified"`
		}
		if err := json.Unmarshal(got.Data, &data); err != nil {
			t.Fatal(err)
		}
		if data.Status != "success" || !data.SignatureVerified {
			t.Fatalf("unexpected signing result: status=%s verified=%v", data.Status, data.SignatureVerified)
		}
	})

	t.Run("T17 duplicate approval rejected", func(t *testing.T) {
		check("T17", http.MethodPost, "/api/v1/sign-sessions/"+sessionID+"/approve", `{"node_id":"node-2"}`, 1009)
	})

	t.Run("T18 missing wallet", func(t *testing.T) {
		check("T18", http.MethodGet, "/api/v1/wallets/missing", "", 1004)
	})

	t.Run("T19 invalid session payload", func(t *testing.T) {
		check("T19", http.MethodPost, "/api/v1/sign-sessions", `{}`, 1001)
	})

	t.Run("T20 session detail", func(t *testing.T) {
		check("T20", http.MethodGet, "/api/v1/sign-sessions/"+sessionID, "", 0)
	})
}
