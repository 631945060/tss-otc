package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"mpc-wallet-demo/app/tss_wallet/api/services"
	"mpc-wallet-demo/config"
	"mpc-wallet-demo/core/httpserver"
	sqlcore "mpc-wallet-demo/core/sql"
	"mpc-wallet-demo/databases/migrations"
	"mpc-wallet-demo/routes"
	redisclient "mpc-wallet-demo/tools/redis"
)

var httpServerCmd = &cobra.Command{
	Use:   "httpServer",
	Short: "start the HTTP API and administration console",
	RunE:  func(command *cobra.Command, _ []string) error { return runHTTPServer(command.Context()) },
}

func init() { rootCmd.AddCommand(httpServerCmd) }

func runHTTPServer(ctx context.Context) error {
	cfg := config.Load()
	db, err := sqlcore.Open(ctx, sqlcore.Options{DSN: cfg.MySQLDSN, MaxOpenConns: cfg.MySQLMaxOpenConns, MaxIdleConns: cfg.MySQLMaxIdleConns})
	if err != nil {
		return err
	}
	if db != nil {
		defer db.Close()
		if err := migrations.Apply(ctx, db); err != nil {
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

	var publisher services.SessionEventPublisher
	if rdb != nil {
		publisher = redisclient.NewSessionEventPublisher(rdb)
	}
	server := httpserver.New(cfg.HTTPAddr, routes.NewServer(cfg.StaticDir, services.NewTSSWalletService(publisher)))
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
