package mysql

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Options struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
}

// Open owns the MySQL pool used by the process composition root.
func Open(ctx context.Context, options Options) (*stdsql.DB, error) {
	if options.DSN == "" {
		return nil, nil
	}
	db, err := stdsql.Open("mysql", options.DSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	db.SetMaxOpenConns(options.MaxOpenConns)
	db.SetMaxIdleConns(options.MaxIdleConns)
	db.SetConnMaxLifetime(30 * time.Minute)
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}
