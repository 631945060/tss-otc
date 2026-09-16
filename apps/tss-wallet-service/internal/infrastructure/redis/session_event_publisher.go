package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"mpc-wallet-demo/apps/tss-wallet-service/internal/application"
)

type SessionEventPublisher struct{ client *goredis.Client }

func NewSessionEventPublisher(client *goredis.Client) *SessionEventPublisher {
	return &SessionEventPublisher{client: client}
}

func (p *SessionEventPublisher) PublishSessionEvent(ctx context.Context, event application.SessionEvent) error {
	if p == nil || p.client == nil {
		return nil
	}
	return p.client.XAdd(ctx, &goredis.XAddArgs{Stream: application.SessionEventStream, Values: map[string]any{"session_id": event.SessionID, "type": event.Type, "status": event.Status, "created_at": event.CreatedAt.UTC().Format(time.RFC3339Nano)}}).Err()
}
