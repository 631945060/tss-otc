package main

import (
	"log"
	"os"

	"mpc-wallet-demo/routes"
)

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	server := routes.NewServer("./web")
	log.Printf("tss wallet service listening on %s", addr)
	log.Fatal(server.Run(addr))
}
