package seed

import (
	"context"
	"database/sql"
	"fmt"
)

// Run applies all seeds in order and commits them together.
func Run(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin seeds: %w", err)
	}
	defer tx.Rollback()

	for _, step := range []struct {
		name string
		run  func(context.Context, *sql.Tx) error
	}{
		{"teams", seedTeams},
	} {
		if err := step.run(ctx, tx); err != nil {
			return fmt.Errorf("seed %s: %w", step.name, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seeds: %w", err)
	}
	return nil
}
