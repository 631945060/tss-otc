package walletmigration

import (
	"strings"
	"testing"
)

func TestEmbeddedMigrationContainsTSSCoreTables(t *testing.T) {
	script, err := files.ReadFile("000001_create_tss_wallet_schema.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"wallets", "participants", "sign_sessions", "transactions", "approvals", "audit_logs", "node_heartbeats"} {
		if !strings.Contains(string(script), "CREATE TABLE IF NOT EXISTS "+table) {
			t.Fatalf("migration does not create %s", table)
		}
	}
}
