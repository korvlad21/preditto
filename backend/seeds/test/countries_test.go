package test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"preditto/seeds/development"
)

func TestSeedCountriesRollsBackOnWriteError(t *testing.T) {
	conn := &teamsConn{failCountry: true}
	db := sql.OpenDB(teamsConnector{conn})
	t.Cleanup(func() { db.Close() })
	err := development.Run(context.Background(), db)
	if err == nil || !strings.Contains(err.Error(), "seed countries: upsert country") {
		t.Fatalf("expected country write error, got %v", err)
	}
	if conn.committed || !conn.rolledBack {
		t.Fatal("failed country seed must roll back without commit")
	}
}

func TestCountriesPostgres(t *testing.T) {
	db := seedPostgres(t)
	if _, err := db.Exec(`
		INSERT INTO countries (id, name, short_name, created_at)
		VALUES (81, 'Old name', 'ENG', '2000-01-01'), (91, 'Other', 'OTH', '2000-01-01');
		ALTER TABLE countries ALTER COLUMN id RESTART WITH 101;
	`); err != nil {
		t.Fatal(err)
	}
	type snapshot struct {
		id      int64
		created time.Time
	}
	previous := map[string]snapshot{}
	for run := 0; run < 2; run++ {
		if err := development.Run(context.Background(), db); err != nil {
			t.Fatal(err)
		}
		rows, err := db.Query("SELECT id, name, short_name, created_at FROM countries")
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for rows.Next() {
			var got snapshot
			var name, code string
			if err := rows.Scan(&got.id, &name, &code, &got.created); err != nil {
				t.Fatal(err)
			}
			if code == "ENG" && (got.id != 81 || name != "England" || got.created.Year() != 2000) {
				t.Fatalf("existing country was not preserved/restored: %v %s", got, name)
			}
			if code == "OTH" && (got.id != 91 || name != "Other") {
				t.Fatal("unrelated country changed")
			}
			if run == 1 && (got.id != previous[code].id || !got.created.Equal(previous[code].created)) {
				t.Fatalf("ID or created_at changed for %s", code)
			}
			if got.created.IsZero() {
				t.Fatalf("missing database timestamp for %s", code)
			}
			previous[code] = got
			count++
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		rows.Close()
		if count != 17 {
			t.Fatalf("countries=%d, want 17", count)
		}
	}
}

func TestCountriesRollBackWithLaterFailurePostgres(t *testing.T) {
	db := seedPostgres(t)
	if _, err := db.Exec("ALTER TABLE teams ALTER COLUMN id RESTART WITH 100"); err != nil {
		t.Fatal(err)
	}
	if err := development.Run(context.Background(), db); err == nil {
		t.Fatal("expected missing favorite team error")
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM countries").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rollback left %d countries", count)
	}
}
