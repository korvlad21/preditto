package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"preditto/internal/config"
	"preditto/internal/database"
	"preditto/internal/seed"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := database.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := seed.Run(ctx, db); err != nil {
		return err
	}
	log.Print("Seeds applied successfully")
	return nil
}
