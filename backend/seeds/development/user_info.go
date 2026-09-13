package development

import (
	"context"
	"database/sql"
	"fmt"
)

type userInfo struct {
	username       string
	firstName      string
	lastName       string
	favoriteTeamID sql.NullInt64
}

var userInfos = []userInfo{
	{"korvlad21", "Vladislav", "Korobkin", sql.NullInt64{Int64: 5, Valid: true}},
	{"hanna_stoma", "Hanna", "Stoma", sql.NullInt64{}},
	{"boroda", "Nikita", "Borodin", sql.NullInt64{Int64: 19, Valid: true}},
}

func seedUserInfo(ctx context.Context, tx *sql.Tx) error {
	for _, item := range userInfos {
		var userID int64
		if err := tx.QueryRowContext(ctx, `
			SELECT id FROM users WHERE username = $1
		`, item.username).Scan(&userID); err != nil {
			return fmt.Errorf("find user for user_info %q: %w", item.username, err)
		}
		_, err := tx.ExecContext(ctx, `а
			INSERT INTO user_info (user_id, first_name, last_name, avatar_url, favorite_team_id, updated_at)
			VALUES ($1, $2, $3, NULL, $4, CURRENT_TIMESTAMP)
			ON CONFLICT (user_id) DO UPDATE
			SET first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name,
				avatar_url = NULL, favorite_team_id = EXCLUDED.favorite_team_id,
				updated_at = CURRENT_TIMESTAMP
		`, userID, item.firstName, item.lastName, item.favoriteTeamID)
		if err != nil {
			return fmt.Errorf("upsert user_info %q: %w", item.username, err)
		}
	}
	return nil
}
