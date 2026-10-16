package test

import (
	"context"
	"strings"
	"testing"

	"preditto/seeds/development"
)

func TestTournamentParticipantsPostgres(t *testing.T) {
	db := seedPostgres(t)
	if _, err := db.Exec("ALTER TABLE users ALTER COLUMN id RESTART WITH 101"); err != nil {
		t.Fatal(err)
	}
	var originalSnapshot string
	for run := 0; run < 2; run++ {
		if err := development.Run(context.Background(), db); err != nil {
			t.Fatal(err)
		}
		var total, active, notPaid, blocked, invalid int
		var snapshot string
		if err := db.QueryRow(`
			SELECT count(*),
				count(*) FILTER (WHERE p.status = 'active'),
				count(*) FILTER (WHERE p.status = 'not_paid'),
				count(*) FILTER (WHERE p.status = 'blocked'),
				count(*) FILTER (WHERE p.tournament_id <> 1 OR p.user_id < 101
					OR u.username IN ('korvlad21', 'hanna_stoma', 'boroda')
					OR p.joined_at IS NULL OR p.joined_at > CURRENT_TIMESTAMP),
				string_agg(p.id::text || ':' || u.username || ':' || p.status || ':' || p.joined_at::text,
					',' ORDER BY u.username)
			FROM tournament_participants p JOIN users u ON u.id = p.user_id
		`).Scan(&total, &active, &notPaid, &blocked, &invalid, &snapshot); err != nil {
			t.Fatal(err)
		}
		if total != 10 || active != 8 || notPaid != 1 || blocked != 1 || invalid != 0 {
			t.Fatalf("run %d: total=%d active=%d not_paid=%d blocked=%d invalid=%d",
				run, total, active, notPaid, blocked, invalid)
		}
		if run == 1 && snapshot != originalSnapshot {
			t.Fatal("rerun changed participant identities/joined_at or failed to restore statuses")
		}
		originalSnapshot = snapshot
		if run == 0 {
			var timestampsDefaulted bool
			if err := db.QueryRow(`SELECT bool_and(p.joined_at = t.created_at)
				FROM tournament_participants p JOIN tournaments t ON t.id = p.tournament_id`).Scan(&timestampsDefaulted); err != nil || !timestampsDefaulted {
				t.Fatalf("joined_at did not use the transaction timestamp: %v", err)
			}
			if _, err := db.Exec("UPDATE tournament_participants SET status = 'blocked'"); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestTournamentParticipantsMissingTournamentRollsBackPostgres(t *testing.T) {
	db := seedPostgres(t)
	if _, err := db.Exec("ALTER TABLE tournaments ALTER COLUMN id RESTART WITH 101"); err != nil {
		t.Fatal(err)
	}
	err := development.Run(context.Background(), db)
	if err == nil || !strings.Contains(err.Error(), "seed tournament_participants: upsert tournament 1 participant") {
		t.Fatalf("expected missing tournament 1 error, got %v", err)
	}
	var count int
	if err := db.QueryRow(`SELECT (SELECT count(*) FROM users) + (SELECT count(*) FROM tournaments)
		+ (SELECT count(*) FROM tournament_participants)`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rollback left %d rows", count)
	}
}
