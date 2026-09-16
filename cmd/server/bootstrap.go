package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"mpc-wallet-demo/app/tss_wallet/api/services"
	"mpc-wallet-demo/config"
	"mpc-wallet-demo/consumers"
	"mpc-wallet-demo/core/httpserver"
	rediscore "mpc-wallet-demo/core/redis"
	sqlcore "mpc-wallet-demo/core/sql"
	"mpc-wallet-demo/databases/migrations"
	"mpc-wallet-demo/routes"
)

// run is the composition root: external infrastructure is initialized before
// the HTTP server begins accepting traffic and closed on process shutdown.
func run(ctx context.Context) error {
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

	redisClient, err := rediscore.Open(ctx, rediscore.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	if err != nil {
		return err
	}
	if redisClient != nil {
		defer redisClient.Close()
		log.Printf("redis connected")
	} else {
		log.Printf("REDIS_ADDR is empty; asynchronous consumers disabled")
	}

	var publisher services.SessionEventPublisher
	if redisClient != nil {
		publisher = rediscore.NewSessionEventPublisher(redisClient)
		if cfg.RunConsumers {
			consumers.NewManager(redisClient, consumerName()).Start(ctx)
		}
	}
	service := services.NewTSSWalletService(publisher)
	server := httpserver.New(cfg.HTTPAddr, routes.NewServer(cfg.StaticDir, service))
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("http shutdown: %v", err)
		}
	}()
	log.Printf("tss wallet service listening on %s", cfg.HTTPAddr)
	return server.Start()
}

func consumerName() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		return "tss-wallet-consumer"
	}
	return fmt.Sprintf("tss-wallet-%s", host)
}
