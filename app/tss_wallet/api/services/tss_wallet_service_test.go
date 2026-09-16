package services

import (
	"context"
	"testing"
	"time"

	"mpc-wallet-demo/app/tss_wallet/api/requests"
)

type testPublisher struct{ events chan SessionEvent }

func (p *testPublisher) PublishSessionEvent(_ context.Context, event SessionEvent) error {
	p.events <- event
	return nil
}

func TestSessionLifecyclePublishesNonSensitiveEvent(t *testing.T) {
	publisher := &testPublisher{events: make(chan SessionEvent, 3)}
	service := NewTSSWalletService(publisher)
	session, err := service.CreateSession(requests.CreateSignSessionReq{WalletID: "wallet-demo-001", Digest: "test-digest"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case event := <-publisher.events:
		if event.SessionID != session.ID || event.Type != "created" || event.Status != "pending" {
			t.Fatalf("unexpected event: %+v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("session creation did not publish an event")
	}
}
