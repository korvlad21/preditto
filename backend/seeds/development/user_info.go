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
	{"alex_morgan", "Alex", "Morgan", sql.NullInt64{Int64: 2, Valid: true}},
	{"sofia_rossi", "Sofia", "Rossi", sql.NullInt64{Int64: 14, Valid: true}},
	{"daniel_weber", "Daniel", "Weber", sql.NullInt64{Int64: 6, Valid: true}},
	{"emma_wilson", "Emma", "Wilson", sql.NullInt64{}},
	{"lucas_silva", "Lucas", "Silva", sql.NullInt64{Int64: 24, Valid: true}},
	{"clara_martin", "Clara", "Martin", sql.NullInt64{Int64: 23, Valid: true}},
	{"mateo_garcia", "Mateo", "Garcia", sql.NullInt64{Int64: 27, Valid: true}},
	{"olivia_brown", "Olivia", "Brown", sql.NullInt64{Int64: 19, Valid: true}},
	{"felix_schmidt", "Felix", "Schmidt", sql.NullInt64{Int64: 8, Valid: true}},
	{"mia_jensen", "Mia", "Jensen", sql.NullInt64{}},
	{"hugo_dubois", "Hugo", "Dubois", sql.NullInt64{Int64: 18, Valid: true}},
	{"ines_costa", "Ines", "Costa", sql.NullInt64{Int64: 33, Valid: true}},
	{"oscar_lind", "Oscar", "Lind", sql.NullInt64{Int64: 7, Valid: true}},
	{"elena_popov", "Elena", "Popov", sql.NullInt64{}},
	{"noah_clarke", "Noah", "Clarke", sql.NullInt64{Int64: 20, Valid: true}},
	{"zara_ahmed", "Zara", "Ahmed", sql.NullInt64{}},
	{"leo_moretti", "Leo", "Moretti", sql.NullInt64{Int64: 22, Valid: true}},
}

func seedUserInfo(ctx context.Context, tx *sql.Tx) error {
	for _, item := range userInfos {
		var userID int64
		if err := tx.QueryRowContext(ctx, `
			SELECT id FROM users WHERE username = $1
		`, item.username).Scan(&userID); err != nil {
			return fmt.Errorf("find user for user_info %q: %w", item.username, err)
		}
		_, err := tx.ExecContext(ctx, `
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
