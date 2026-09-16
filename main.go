package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"mpc-wallet-demo/cmd"
)

// main mirrors the reference service: it only owns process lifecycle and
// delegates executable roles to commands registered in cmd/.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}
