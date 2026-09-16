package development

import (
	"context"
	"database/sql"
	"fmt"

	"preditto/internal/repository"
)

type userRole struct {
	username string
	roleCode string
}

var userRoles = []userRole{
	{"korvlad21", "admin"},
	{"hanna_stoma", "admin"},
	{"boroda", "participant"},
}

func seedUserRoles(ctx context.Context, tx *sql.Tx) error {
	for _, item := range userRoles {
		var userID, roleID int64
		if err := tx.QueryRowContext(ctx, `
			SELECT id FROM users WHERE username = $1
		`, item.username).Scan(&userID); err != nil {
			return fmt.Errorf("find user for user_roles %q: %w", item.username, err)
		}
		if err := tx.QueryRowContext(ctx, `
			SELECT id FROM roles WHERE code = $1
		`, item.roleCode).Scan(&roleID); err != nil {
			return fmt.Errorf("find role for user_roles %q: %w", item.roleCode, err)
		}
		if err := repository.AssignUserRole(ctx, tx, userID, roleID); err != nil {
			return fmt.Errorf("assign role %q to user %q: %w", item.roleCode, item.username, err)
		}
	}
	return nil
}
