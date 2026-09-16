package consumers

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"mpc-wallet-demo/apps/tss-wallet-service/internal/application"
)

const consumerGroup = "tss-wallet-audit"

// Manager owns background consumers. More consumers can be registered here
// without coupling HTTP handlers to asynchronous operational work.
type Manager struct {
	client       *goredis.Client
	consumerName string
}

func NewManager(client *goredis.Client, consumerName string) *Manager {
	return &Manager{client: client, consumerName: consumerName}
}

func (m *Manager) Start(ctx context.Context) {
	if m == nil || m.client == nil {
		return
	}
	go m.consumeSessionEvents(ctx)
}

func (m *Manager) consumeSessionEvents(ctx context.Context) {
	if err := m.client.XGroupCreateMkStream(ctx, application.SessionEventStream, consumerGroup, "0").Err(); err != nil && !strings.HasPrefix(err.Error(), "BUSYGROUP") {
		log.Printf("create Redis consumer group: %v", err)
		return
	}
	for ctx.Err() == nil {
		streams, err := m.client.XReadGroup(ctx, &goredis.XReadGroupArgs{Group: consumerGroup, Consumer: m.consumerName, Streams: []string{application.SessionEventStream, ">"}, Count: 10, Block: time.Second}).Result()
		if err != nil {
			if errors.Is(err, goredis.Nil) || ctx.Err() != nil {
				continue
			}
			log.Printf("read signing-session event: %v", err)
			continue
		}
		for _, stream := range streams {
			for _, message := range stream.Messages {
				log.Printf("consumed signing-session event id=%s session=%v type=%v status=%v", message.ID, message.Values["session_id"], message.Values["type"], message.Values["status"])
				if err := m.client.XAck(ctx, application.SessionEventStream, consumerGroup, message.ID).Err(); err != nil {
					log.Printf("ack signing-session event: %v", err)
				}
			}
		}
	}
}
