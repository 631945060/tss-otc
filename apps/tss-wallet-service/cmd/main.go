package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"mpc-wallet-demo/apps/tss-wallet-service/cmd/command"
)

// main mirrors the reference service: it only owns process lifecycle and
// delegates executable roles to commands registered in cmd/.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := command.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}
