package services

import (
	"context"
	"time"
)

const SessionEventStream = "tss:sign-session:events"

// SessionEvent is a non-sensitive operational event. It never includes key
// shares, protocol payloads or signature material.
type SessionEvent struct {
	SessionID string    `json:"session_id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionEventPublisher interface {
	PublishSessionEvent(context.Context, SessionEvent) error
}
