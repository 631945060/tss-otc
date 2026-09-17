package command

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"mpc-wallet-demo/apps/tss-wallet-service/internal/application"
	"mpc-wallet-demo/apps/tss-wallet-service/internal/config"
	eventredis "mpc-wallet-demo/apps/tss-wallet-service/internal/infrastructure/redis"
	"mpc-wallet-demo/apps/tss-wallet-service/internal/transport"
	"mpc-wallet-demo/apps/tss-wallet-service/internal/transport/httpapi"
	mysql "mpc-wallet-demo/internal/database/mysql"
	redisclient "mpc-wallet-demo/internal/database/redis"
	walletmigration "mpc-wallet-demo/migrations/wallet"
)

var httpServerCmd = &cobra.Command{
	Use:   "httpServer",
	Short: "start the HTTP API and administration console",
	RunE:  func(command *cobra.Command, _ []string) error { return runHTTPServer(command.Context()) },
}

func init() { rootCmd.AddCommand(httpServerCmd) }

func runHTTPServer(ctx context.Context) error {
	cfg := config.Load()
	if cfg.SignerMode == "production" {
		return fmt.Errorf("SIGNER_MODE=production requires an external distributed signer; local key storage is disabled")
	}
	db, err := mysql.Open(ctx, mysql.Options{DSN: cfg.MySQLDSN, MaxOpenConns: cfg.MySQLMaxOpenConns, MaxIdleConns: cfg.MySQLMaxIdleConns})
	if err != nil {
		return err
	}
	if db != nil {
		defer db.Close()
		if err := walletmigration.Apply(ctx, db); err != nil {
			return err
		}
		log.Printf("mysql connected; migrations applied")
	} else {
		log.Printf("MYSQL_DSN is empty; using in-memory demonstration state")
	}

	rdb, err := redisclient.Open(ctx, redisclient.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	if err != nil {
		return err
	}
	if rdb != nil {
		defer rdb.Close()
		log.Printf("redis connected")
	} else {
		log.Printf("REDIS_ADDR is empty; session events are not queued")
	}

	var publisher application.SessionEventPublisher
	if rdb != nil {
		publisher = eventredis.NewSessionEventPublisher(rdb)
	}
	server := transport.NewHTTPServer(cfg.HTTPAddr, httpapi.NewRouter(cfg.StaticDir, application.NewTSSWalletService(publisher)))
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("http shutdown: %v", err)
		}
	}()
	log.Printf("tss wallet HTTP server listening on %s", cfg.HTTPAddr)
	return server.Start()
}
