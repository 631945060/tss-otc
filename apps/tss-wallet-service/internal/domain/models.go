package domain

import "time"

type Wallet struct {
	ID        string    `json:"id"`
	PublicKey string    `json:"public_key"`
	Address   string    `json:"address"`
	Network   string    `json:"network"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Participant struct {
	ID                     string    `json:"id"`
	Endpoint               string    `json:"endpoint"`
	CertificateFingerprint string    `json:"certificate_fingerprint"`
	Status                 string    `json:"status"`
	KeyEpoch               int       `json:"key_epoch"`
	Version                string    `json:"version"`
	LastHeartbeat          time.Time `json:"last_heartbeat"`
}

type Transaction struct {
	ID          string    `json:"id"`
	WalletID    string    `json:"wallet_id"`
	FromAddress string    `json:"from_address"`
	ToAddress   string    `json:"to_address"`
	Amount      string    `json:"amount"`
	Fee         string    `json:"fee"`
	Digest      string    `json:"digest"`
	SessionID   string    `json:"session_id,omitempty"`
	TxHash      string    `json:"tx_hash,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SignSession struct {
	ID            string     `json:"id"`
	WalletID      string     `json:"wallet_id"`
	TransactionID string     `json:"transaction_id,omitempty"`
	Digest        string     `json:"digest"`
	Threshold     int        `json:"threshold"`
	Status        string     `json:"status"`
	Participants  []string   `json:"participants"`
	Approvers     []string   `json:"approvers"`
	CreatedAt     time.Time  `json:"created_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
}

type KeyEpoch struct {
	WalletID           string    `json:"wallet_id"`
	Epoch              int       `json:"epoch"`
	Operation          string    `json:"operation"`
	PublicKeyUnchanged bool      `json:"public_key_unchanged"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
}

type AuditLog struct {
	ID        string    `json:"id"`
	Actor     string    `json:"actor"`
	Action    string    `json:"action"`
	Result    string    `json:"result"`
	Detail    string    `json:"detail"`
	SessionID string    `json:"session_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
