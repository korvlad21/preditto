package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// executor allows callers to use either a database or their existing transaction.
type executor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

// AssignUserRole adds a role without changing an existing assignment.
func AssignUserRole(ctx context.Context, db executor, userID, roleID int64) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, userID, roleID)
	if err != nil {
		return fmt.Errorf("assign role %d to user %d: %w", roleID, userID, err)
	}
	return nil
}

// RemoveUserRole succeeds even when the assignment does not exist.
func RemoveUserRole(ctx context.Context, db executor, userID, roleID int64) error {
	_, err := db.ExecContext(ctx, `
		DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2
	`, userID, roleID)
	if err != nil {
		return fmt.Errorf("remove role %d from user %d: %w", roleID, userID, err)
	}
	return nil
}
