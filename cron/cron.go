package cron

import (
	"context"
	"log"
	"time"
)

// Run holds the explicit cron process role. It is intentionally separated from
// HTTP and consumer processes so operational workloads can be scaled safely.
func Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	log.Printf("tss wallet cron started")
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			log.Printf("cron sweep: no stale in-memory sessions to reconcile")
		}
	}
}
