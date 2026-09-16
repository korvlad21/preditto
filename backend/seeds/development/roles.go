package development

import (
	"context"
	"database/sql"
	"fmt"
)

type role struct {
	code string
	name string
}

var roles = []role{
	{"admin", "Администратор"},
	{"participant", "Участник"},
}

func seedRoles(ctx context.Context, tx *sql.Tx) error {
	for _, item := range roles {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO roles (code, name) VALUES ($1, $2)
			ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
		`, item.code, item.name)
		if err != nil {
			return fmt.Errorf("upsert role %q: %w", item.code, err)
		}
	}
	return nil
}
