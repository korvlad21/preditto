package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"time"

	"preditto/internal/config"

	"github.com/lib/pq"
)

func NewPostgres(ctx context.Context, cfg config.PostgresConfig) (*sql.DB, error) {
	driverConfig, err := pq.NewConfig(postgresDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL: %w", err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		driverConfig.ConnectTimeout = time.Until(deadline)
		if driverConfig.ConnectTimeout <= 0 {
			return nil, fmt.Errorf("ping PostgreSQL: %w", context.DeadlineExceeded)
		}
	}
	connector, err := pq.NewConnectorConfig(driverConfig)
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL: %w", err)
	}
	db := sql.OpenDB(connector)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}
	return db, nil
}

func postgresDSN(cfg config.PostgresConfig) string {
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, cfg.Port),
	}
	query := url.Values{}
	query.Set("dbname", cfg.Database)
	query.Set("sslmode", cfg.SSLMode)
	dsn.RawQuery = query.Encode()
	return dsn.String()
}
