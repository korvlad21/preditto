package database

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"preditto/internal/config"
)

func testConfig() config.PostgresConfig {
	return config.PostgresConfig{
		Host: "127.0.0.1", Port: "5432", User: "test",
		Password: "secret", Database: "test", SSLMode: "disable",
	}
}

func TestPostgresDSNPreservesSpecialCharacters(t *testing.T) {
	cfg := testConfig()
	cfg.Host = "::1"
	cfg.User = "user:@/ name"
	cfg.Password = " p@ss:'\\/?#%&= é"
	cfg.Database = "db/name?#&= é"
	parsed, err := pq.NewConfig(postgresDSN(cfg))
	if err != nil {
		t.Fatal("driver could not parse the generated DSN")
	}
	if parsed.Host != cfg.Host || parsed.Port != 5432 || parsed.User != cfg.User ||
		parsed.Password != cfg.Password || parsed.Database != cfg.Database || string(parsed.SSLMode) != cfg.SSLMode {
		t.Fatal("driver parsed connection settings differently from the configuration")
	}
}

func TestNewPostgresReturnsOpenError(t *testing.T) {
	cfg := testConfig()
	cfg.SSLMode = "invalid"
	db, err := NewPostgres(context.Background(), cfg)
	if db != nil || err == nil || !strings.HasPrefix(err.Error(), "open PostgreSQL:") || errors.Unwrap(err) == nil {
		t.Fatalf("expected a wrapped open error and nil database, got %v, %v", db, err)
	}
}

func TestNewPostgresReturnsCanceledPing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	db, err := NewPostgres(ctx, testConfig())
	if db != nil || !errors.Is(err, context.Canceled) || !strings.HasPrefix(err.Error(), "ping PostgreSQL:") {
		t.Fatalf("expected a wrapped cancellation and nil database, got %v, %v", db, err)
	}
}

func TestNewPostgresReturnsConnectionFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan struct{})
	go func() {
		defer close(accepted)
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
		}
	}()
	cfg := testConfig()
	cfg.Host, cfg.Port, err = net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, err := NewPostgres(ctx, cfg)
	if db != nil || err == nil || !strings.HasPrefix(err.Error(), "ping PostgreSQL:") || errors.Unwrap(err) == nil {
		t.Fatalf("expected a wrapped connection error and nil database, got %v, %v", db, err)
	}
	select {
	case <-accepted:
	case <-ctx.Done():
		t.Fatal("Ping did not attempt a connection")
	}
}

func TestNewPostgresTimesOutDuringHandshake(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan struct{})
	defer close(done)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			defer conn.Close()
			<-done
		}
	}()
	cfg := testConfig()
	cfg.Host, cfg.Port, err = net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	result := make(chan error, 1)
	go func() {
		db, err := NewPostgres(ctx, cfg)
		if db != nil {
			db.Close()
		}
		result <- err
	}()
	select {
	case err := <-result:
		var timeout net.Error
		if !errors.As(err, &timeout) || !timeout.Timeout() {
			t.Fatalf("expected a handshake timeout, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("connection handshake exceeded the context deadline")
	}
}
