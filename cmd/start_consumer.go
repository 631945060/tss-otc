package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"mpc-wallet-demo/config"
	"mpc-wallet-demo/consumers"
	redisclient "mpc-wallet-demo/tools/redis"
)

var startConsumerCmd = &cobra.Command{
	Use:   "startConsumers",
	Short: "start Redis Stream consumers",
	RunE:  func(command *cobra.Command, _ []string) error { return runConsumers(command.Context()) },
}

func init() { rootCmd.AddCommand(startConsumerCmd) }

func runConsumers(ctx context.Context) error {
	cfg := config.Load()
	if cfg.RedisAddr == "" {
		return fmt.Errorf("REDIS_ADDR is required for startConsumers")
	}
	rdb, err := redisclient.Open(ctx, redisclient.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	if err != nil {
		return err
	}
	defer rdb.Close()
	consumers.NewManager(rdb, consumerName()).Start(ctx)
	<-ctx.Done()
	return nil
}

func consumerName() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		return "tss-wallet-consumer"
	}
	return "tss-wallet-" + host
}
