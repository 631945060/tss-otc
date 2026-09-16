package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"mpc-wallet-demo/apps/tss-wallet-service/internal/domain"
)

var (
	ErrInvalidState = errors.New("operation is not allowed in the current state")
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("duplicate or conflicting operation")
)

// TSSWalletService is the business layer. In-memory state is intentional for
// the thesis demonstration; databases/migrations defines the persistence boundary.
type TSSWalletService struct {
	mu           sync.RWMutex
	sequence     uint64
	wallets      map[string]domain.Wallet
	participants map[string]domain.Participant
	transactions map[string]domain.Transaction
	sessions     map[string]domain.SignSession
	epochs       []domain.KeyEpoch
	audits       []domain.AuditLog
	publisher    SessionEventPublisher
}

func NewTSSWalletService(publishers ...SessionEventPublisher) *TSSWalletService {
	now := time.Now().UTC()
	wallet := domain.Wallet{ID: "wallet-demo-001", PublicKey: "tss-demo-public-key", Address: "demo-address", Network: "testnet", Status: "active", CreatedAt: now, UpdatedAt: now}
	participants := map[string]domain.Participant{}
	for index := 1; index <= 3; index++ {
		id := fmt.Sprintf("node-%d", index)
		participants[id] = domain.Participant{ID: id, Endpoint: fmt.Sprintf("https://%s.internal:9443", id), CertificateFingerprint: "demo-cert-" + id, Status: "online", KeyEpoch: 1, Version: "tss-agent/1.5.0", LastHeartbeat: now}
	}
	service := &TSSWalletService{
		wallets:      map[string]domain.Wallet{wallet.ID: wallet},
		participants: participants,
		transactions: make(map[string]domain.Transaction),
		sessions:     make(map[string]domain.SignSession),
		epochs:       []domain.KeyEpoch{{WalletID: wallet.ID, Epoch: 1, Operation: "dkg", PublicKeyUnchanged: false, Status: "active", CreatedAt: now}},
	}
	if len(publishers) > 0 {
		service.publisher = publishers[0]
	}
	return service
}

func (s *TSSWalletService) ListWallets() []domain.Wallet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.Wallet, 0, len(s.wallets))
	for _, wallet := range s.wallets {
		items = append(items, wallet)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items
}

func (s *TSSWalletService) CreateWallet(input CreateWalletInput) domain.Wallet {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	network := defaultValue(input.Network, "testnet")
	id := s.nextID("wallet", now)
	fingerprint := sha256.Sum256([]byte(id + network))
	wallet := domain.Wallet{ID: id, PublicKey: "dkg-public-" + hex.EncodeToString(fingerprint[:8]), Address: "tss-" + hex.EncodeToString(fingerprint[8:14]), Network: network, Status: "active", CreatedAt: now, UpdatedAt: now}
	s.wallets[id] = wallet
	s.epochs = append(s.epochs, domain.KeyEpoch{WalletID: id, Epoch: 1, Operation: "dkg", PublicKeyUnchanged: false, Status: "active", CreatedAt: now})
	s.appendAudit("operator", "create_wallet", "success", "Created wallet metadata and DKG key epoch", "", now)
	return wallet
}

func (s *TSSWalletService) Wallet(id string) (domain.Wallet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	wallet, ok := s.wallets[id]
	if !ok {
		return domain.Wallet{}, ErrNotFound
	}
	return wallet, nil
}

func (s *TSSWalletService) WalletAddresses(id string) (map[string]any, error) {
	wallet, err := s.Wallet(id)
	if err != nil {
		return nil, err
	}
	return map[string]any{"wallet_id": wallet.ID, "items": []map[string]string{{"address": wallet.Address, "network": wallet.Network, "type": "tss"}}}, nil
}

func (s *TSSWalletService) WalletBalance(id string) (map[string]any, error) {
	wallet, err := s.Wallet(id)
	if err != nil {
		return nil, err
	}
	return map[string]any{"wallet_id": wallet.ID, "network": wallet.Network, "available": "100.00000000", "locked": "0.00000000", "unit": "TEST"}, nil
}

func (s *TSSWalletService) CreateTransaction(input CreateTransactionInput) (domain.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	wallet, ok := s.wallets[input.WalletID]
	if !ok {
		return domain.Transaction{}, ErrNotFound
	}
	now := time.Now().UTC()
	tx := domain.Transaction{ID: s.nextID("tx", now), WalletID: wallet.ID, FromAddress: wallet.Address, ToAddress: input.ToAddress, Amount: input.Amount, Fee: defaultValue(input.Fee, "0.00010000"), Digest: input.Digest, Status: "awaiting_approvals", CreatedAt: now, UpdatedAt: now}
	s.transactions[tx.ID] = tx
	s.appendAudit("operator", "create_transaction", "success", "Created transaction awaiting threshold approval", "", now)
	return tx, nil
}

func (s *TSSWalletService) Transaction(id string) (domain.Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tx, ok := s.transactions[id]
	if !ok {
		return domain.Transaction{}, ErrNotFound
	}
	return tx, nil
}

func (s *TSSWalletService) ListTransactions() []domain.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.Transaction, 0, len(s.transactions))
	for _, tx := range s.transactions {
		items = append(items, tx)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items
}

func (s *TSSWalletService) CreateSession(input CreateSignSessionInput) (domain.SignSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.wallets[input.WalletID]; !ok {
		return domain.SignSession{}, ErrNotFound
	}
	if input.TransactionID != "" {
		tx, ok := s.transactions[input.TransactionID]
		if !ok {
			return domain.SignSession{}, ErrNotFound
		}
		if tx.WalletID != input.WalletID || tx.Digest != input.Digest || tx.SessionID != "" {
			return domain.SignSession{}, ErrConflict
		}
	}
	now := time.Now().UTC()
	session := domain.SignSession{ID: s.nextID("sign", now), WalletID: input.WalletID, TransactionID: input.TransactionID, Digest: input.Digest, Threshold: 2, Status: "pending", Participants: []string{"node-1", "node-2", "node-3"}, Approvers: []string{}, CreatedAt: now}
	s.sessions[session.ID] = session
	if input.TransactionID != "" {
		tx := s.transactions[input.TransactionID]
		tx.SessionID, tx.UpdatedAt = session.ID, now
		s.transactions[tx.ID] = tx
	}
	s.appendAudit("operator", "create_sign_session", "success", "Created 2-of-3 signing session", session.ID, now)
	s.publishSessionEvent(session, "created")
	return session, nil
}

func (s *TSSWalletService) Session(id string) (domain.SignSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return domain.SignSession{}, ErrNotFound
	}
	return session, nil
}

func (s *TSSWalletService) ListSessions() []domain.SignSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.SignSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		items = append(items, session)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items
}

func (s *TSSWalletService) ApproveSession(id, nodeID string) (domain.SignSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return domain.SignSession{}, ErrNotFound
	}
	if session.Status != "pending" && session.Status != "signing" {
		return domain.SignSession{}, ErrInvalidState
	}
	if _, ok := s.participants[nodeID]; !ok {
		return domain.SignSession{}, ErrNotFound
	}
	for _, approver := range session.Approvers {
		if approver == nodeID {
			return domain.SignSession{}, ErrConflict
		}
	}
	now := time.Now().UTC()
	session.Approvers = append(session.Approvers, nodeID)
	session.Status = "signing"
	s.appendAudit(nodeID, "approve_sign_session", "success", "Approved signing session", session.ID, now)
	if len(session.Approvers) >= session.Threshold {
		session.Status, session.FinishedAt = "success", &now
		s.appendAudit("coordinator", "complete_sign_session", "success", "Approval threshold reached; protocol output accepted", session.ID, now)
		if session.TransactionID != "" {
			tx := s.transactions[session.TransactionID]
			tx.Status, tx.TxHash, tx.UpdatedAt = "signed", demoHash(session.ID+session.Digest), now
			s.transactions[tx.ID] = tx
		}
	}
	s.sessions[id] = session
	s.publishSessionEvent(session, "approval_recorded")
	return session, nil
}

func (s *TSSWalletService) CancelSession(id string) (domain.SignSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return domain.SignSession{}, ErrNotFound
	}
	if session.Status != "pending" && session.Status != "signing" {
		return domain.SignSession{}, ErrInvalidState
	}
	now := time.Now().UTC()
	session.Status, session.FinishedAt = "aborted", &now
	s.sessions[id] = session
	if session.TransactionID != "" {
		tx := s.transactions[session.TransactionID]
		tx.Status, tx.UpdatedAt = "cancelled", now
		s.transactions[tx.ID] = tx
	}
	s.appendAudit("operator", "cancel_sign_session", "success", "Cancelled signing session", session.ID, now)
	s.publishSessionEvent(session, "cancelled")
	return session, nil
}

func (s *TSSWalletService) ListParticipants() []domain.Participant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.Participant, 0, len(s.participants))
	for _, participant := range s.participants {
		items = append(items, participant)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}

func (s *TSSWalletService) RefreshParticipant(id string) (domain.KeyEpoch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	participant, ok := s.participants[id]
	if !ok {
		return domain.KeyEpoch{}, ErrNotFound
	}
	now := time.Now().UTC()
	participant.KeyEpoch++
	s.participants[id] = participant
	epoch := domain.KeyEpoch{WalletID: "wallet-demo-001", Epoch: participant.KeyEpoch, Operation: "resharing_requested", PublicKeyUnchanged: true, Status: "pending", CreatedAt: now}
	s.epochs = append(s.epochs, epoch)
	s.appendAudit("operator", "refresh_key_share", "pending", "Requested resharing; demo does not persist key material", "", now)
	return epoch, nil
}

func (s *TSSWalletService) Heartbeat(input HeartbeatInput) (domain.Participant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	participant, ok := s.participants[input.NodeID]
	if !ok {
		return domain.Participant{}, ErrNotFound
	}
	participant.Version, participant.Status, participant.LastHeartbeat = input.Version, "online", time.Now().UTC()
	s.participants[participant.ID] = participant
	return participant, nil
}

func (s *TSSWalletService) ListAuditLogs() []domain.AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := append([]domain.AuditLog(nil), s.audits...)
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items
}

func (s *TSSWalletService) Metrics() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	online, completed := 0, 0
	for _, participant := range s.participants {
		if participant.Status == "online" {
			online++
		}
	}
	for _, session := range s.sessions {
		if session.Status == "success" {
			completed++
		}
	}
	return map[string]any{"wallet_count": len(s.wallets), "online_nodes": online, "active_sessions": len(s.sessions), "completed_sessions": completed, "transactions": len(s.transactions), "threshold": "2-of-3"}
}

func (s *TSSWalletService) appendAudit(actor, action, result, detail, sessionID string, now time.Time) {
	s.sequence++
	s.audits = append(s.audits, domain.AuditLog{ID: fmt.Sprintf("audit-%06d", s.sequence), Actor: actor, Action: action, Result: result, Detail: detail, SessionID: sessionID, CreatedAt: now})
}

func (s *TSSWalletService) publishSessionEvent(session domain.SignSession, eventType string) {
	if s.publisher == nil {
		return
	}
	event := SessionEvent{SessionID: session.ID, Type: eventType, Status: session.Status, CreatedAt: time.Now().UTC()}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := s.publisher.PublishSessionEvent(ctx, event); err != nil {
			return
		}
	}()
}

func (s *TSSWalletService) nextID(prefix string, now time.Time) string {
	s.sequence++
	return fmt.Sprintf("%s-%s-%06d", prefix, now.Format("20060102150405"), s.sequence)
}
func defaultValue(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
func demoHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
