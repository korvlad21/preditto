package test

import (
	"context"
	"strings"
	"testing"

	"preditto/seeds/development"
)

func TestTournamentsAndPrizeRulesPostgres(t *testing.T) {
	for _, scenario := range []string{"new", "existing"} {
		t.Run(scenario, func(t *testing.T) {
			db := seedPostgres(t)
			const slug = "uefa-champions-league-2026-2027"
			if _, err := db.Exec(`
				ALTER TABLE tournaments ALTER COLUMN id RESTART WITH 101;
				INSERT INTO tournaments (id, name, slug) VALUES (1, 'Other', 'other');
				INSERT INTO tournament_prize_rules VALUES (1, 1, 100.00);
			`); err != nil {
				t.Fatal(err)
			}
			wantID := int64(101)
			if scenario == "existing" {
				wantID = 81
				if _, err := db.Exec(`
					INSERT INTO tournaments (id, name, slug, status, starts_at, ends_at,
						type_of_tournament, cost_amount, amount_currency, created_at, updated_at)
					VALUES (81, 'Old name', $1, 'planned', '2000-01-01', '2000-02-01',
						'custom', 123.45, 'EUR', '2000-01-01', '2000-01-01');
				`, slug); err != nil {
					t.Fatal(err)
				}
			}
			var createdAt string
			for run := 0; run < 2; run++ {
				if err := development.Run(context.Background(), db); err != nil {
					t.Fatal(err)
				}
				var id int64
				var name, status, starts, ends, created string
				var fieldsValid, timestampsValid bool
				if err := db.QueryRow(`
					SELECT id, name, status, starts_at::text, ends_at::text, created_at::text,
						cost_amount = 10.00
							AND type_of_tournament = 'Champions league' AND amount_currency = 'USD',
						created_at IS NOT NULL AND updated_at >= created_at AND updated_at > TIMESTAMP '2000-01-01'
					FROM tournaments WHERE slug = $1
				`, slug).Scan(&id, &name, &status, &starts, &ends, &created, &fieldsValid, &timestampsValid); err != nil {
					t.Fatal(err)
				}
				if id != wantID || name != "Лига чемпионов УЕФА 2026/2027" || status != "active" ||
					starts != "2026-09-08 00:00:00" || ends != "2027-06-05 00:00:00" || !fieldsValid || !timestampsValid {
					t.Fatalf("invalid tournament: id=%d name=%q status=%s dates=%s/%s fields=%v timestamps=%v",
						id, name, status, starts, ends, fieldsValid, timestampsValid)
				}
				if (run == 1 && created != createdAt) || (scenario == "existing" && created != "2000-01-01 00:00:00") {
					t.Fatal("created_at changed")
				}
				createdAt = created
				var tournaments, rules int
				var percentages, total string
				if err := db.QueryRow(`
					SELECT (SELECT count(*) FROM tournaments), count(*),
						string_agg(place::text || ':' || percentage::text, ',' ORDER BY place), sum(percentage)::text
					FROM tournament_prize_rules WHERE tournament_id = $1
				`, id).Scan(&tournaments, &rules, &percentages, &total); err != nil {
					t.Fatal(err)
				}
				if tournaments != 2 || rules != 3 || percentages != "1:50.00,2:30.00,3:15.00" || total != "95.00" {
					t.Fatalf("unexpected tournament/prize rows: %d %d %s %s", tournaments, rules, percentages, total)
				}
				var otherUnchanged bool
				if err := db.QueryRow(`SELECT t.name = 'Other' AND p.percentage = 100.00
					FROM tournaments t JOIN tournament_prize_rules p ON p.tournament_id = t.id
					WHERE t.id = 1 AND p.place = 1`).Scan(&otherUnchanged); err != nil || !otherUnchanged {
					t.Fatalf("unrelated tournament changed: %v", err)
				}
				if run == 0 {
					if _, err := db.Exec(`UPDATE tournament_prize_rules SET percentage = 1 WHERE tournament_id = $1`, id); err != nil {
						t.Fatal(err)
					}
					if _, err := db.Exec(`DELETE FROM tournament_prize_rules WHERE tournament_id = $1 AND place = 3`, id); err != nil {
						t.Fatal(err)
					}
				}
			}
		})
	}
}

func TestTournamentPrizeFailureRollsBackPostgres(t *testing.T) {
	db := seedPostgres(t)
	if _, err := db.Exec(`ALTER TABLE tournament_prize_rules ADD CONSTRAINT reject_second_prize CHECK (place <> 2)`); err != nil {
		t.Fatal(err)
	}
	err := development.Run(context.Background(), db)
	if err == nil || !strings.Contains(err.Error(), "seed tournament_prize_rules:") {
		t.Fatalf("expected prize rule failure, got %v", err)
	}
	var count int
	if err := db.QueryRow(`SELECT (SELECT count(*) FROM countries) + (SELECT count(*) FROM teams)
		+ (SELECT count(*) FROM users) + (SELECT count(*) FROM user_info) + (SELECT count(*) FROM roles)
		+ (SELECT count(*) FROM user_roles) + (SELECT count(*) FROM tournaments)
		+ (SELECT count(*) FROM tournament_participants)
		+ (SELECT count(*) FROM tournament_prize_rules)`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rollback left %d seeded rows", count)
	}
}
