package test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"preditto/seeds/development"
)

// Each test uses a disposable schema, never the database's existing tables.
func seedPostgres(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("PREDITTO_SEED_TEST_DSN")
	if dsn == "" {
		t.Skip("set PREDITTO_SEED_TEST_DSN to run PostgreSQL seed integration tests")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	schema := fmt.Sprintf("seed_test_%d", time.Now().UnixNano())
	if _, err := db.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Error(err)
		}
	})
	if _, err := db.Exec("SET search_path TO " + schema); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"000001_create_users.up.sql", "000002_create_teams.up.sql", "000003_create_user_info.up.sql", "000004_create_roles.up.sql", "000006_create_user_roles.up.sql", "000008_create_countries.up.sql"} {
		migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(string(migration)); err != nil {
			t.Fatalf("apply %s: %v", name, err)
		}
	}
	return db
}

func TestUsersAndUserInfoPostgres(t *testing.T) {
	db := seedPostgres(t)
	// Exercise username-only and email-only matches, with IDs unlike 1, 2, 3.
	_, err := db.Exec(`
		INSERT INTO users (id, username, email, password_hash, status) VALUES
			(81, 'korvlad21', 'previous@example.test', 'old hash', 'BLOCKED'),
			(91, 'previous_hanna', 'hannastoma9@gmail.com', 'old hash', 'BLOCKED');
		ALTER TABLE users ALTER COLUMN id RESTART WITH 101;
	`)
	if err != nil {
		t.Fatal(err)
	}
	expected := []struct {
		id                      int64
		username, email, status string
		firstName, lastName     string
		favorite                sql.NullInt64
	}{
		{81, "korvlad21", "vladek2000@inbox.ru", "ACTIVE", "Vladislav", "Korobkin", sql.NullInt64{Int64: 5, Valid: true}},
		{91, "hanna_stoma", "hannastoma9@gmail.com", "ACTIVE", "Hanna", "Stoma", sql.NullInt64{}},
		{101, "boroda", "borodin_nikita@yandex.ru", "BLOCKED", "Nikita", "Borodin", sql.NullInt64{Int64: 19, Valid: true}},
	}
	profileIDs := make(map[string]int64)
	for run := 0; run < 2; run++ {
		if err := development.Run(context.Background(), db); err != nil {
			t.Fatal(err)
		}
		for _, table := range []string{"users", "user_info", "teams"} {
			count := 0
			if err := db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil {
				t.Fatal(err)
			}
			want := 3
			if table == "teams" {
				want = 36
			}
			if count != want {
				t.Fatalf("run %d: %s count = %d, want %d", run, table, count, want)
			}
		}
		for _, want := range expected {
			var id, profileID int64
			var email, hash, status, firstName, lastName string
			var avatar sql.NullString
			var favorite sql.NullInt64
			var recent bool
			err := db.QueryRow(`
				SELECT u.id, u.email, u.password_hash, u.status, i.id,
					i.first_name, i.last_name, i.avatar_url, i.favorite_team_id,
					i.updated_at BETWEEN CURRENT_TIMESTAMP - INTERVAL '10 seconds' AND CURRENT_TIMESTAMP
				FROM users u JOIN user_info i ON i.user_id = u.id WHERE u.username = $1
			`, want.username).Scan(&id, &email, &hash, &status, &profileID, &firstName, &lastName, &avatar, &favorite, &recent)
			if err != nil {
				t.Fatal(err)
			}
			if id != want.id || email != want.email || status != want.status || firstName != want.firstName || lastName != want.lastName || avatar.Valid || favorite != want.favorite || !recent {
				t.Fatalf("incorrect user/profile for %s: id=%d email=%s status=%s name=%s %s avatar=%v favorite=%v recent=%v", want.username, id, email, status, firstName, lastName, avatar, favorite, recent)
			}
			if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("12345678")); err != nil {
				t.Fatalf("invalid password hash for %s: %v", want.username, err)
			}
			if run == 1 && profileIDs[want.username] != profileID {
				t.Fatalf("profile ID changed for %s", want.username)
			}
			profileIDs[want.username] = profileID
		}
		if run == 0 {
			if _, err := db.Exec(`
				UPDATE users SET status = 'MODIFIED';
				UPDATE user_info SET first_name = 'Modified', avatar_url = 'old avatar',
					favorite_team_id = NULL, updated_at = '2000-01-01';
			`); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestUsersUniqueConflictRollsBackPostgres(t *testing.T) {
	db := seedPostgres(t)
	if _, err := db.Exec(`
		INSERT INTO users (username, email, password_hash, status) VALUES
			('korvlad21', 'other@example.test', 'old hash', 'BLOCKED'),
			('another_user', 'vladek2000@inbox.ru', 'old hash', 'ACTIVE');
	`); err != nil {
		t.Fatal(err)
	}
	err := development.Run(context.Background(), db)
	if err == nil || !strings.Contains(err.Error(), "seed users: upsert user") {
		t.Fatalf("expected user identity conflict, got %v", err)
	}
	var teams, users, profiles int
	if err := db.QueryRow("SELECT (SELECT count(*) FROM teams), (SELECT count(*) FROM users), (SELECT count(*) FROM user_info)").Scan(&teams, &users, &profiles); err != nil {
		t.Fatal(err)
	}
	if teams != 0 || users != 2 || profiles != 0 {
		t.Fatalf("rollback failed: teams=%d users=%d profiles=%d", teams, users, profiles)
	}
}

func TestUserInfoMissingFavoriteTeamRollsBackPostgres(t *testing.T) {
	db := seedPostgres(t)
	if _, err := db.Exec("ALTER TABLE teams ALTER COLUMN id RESTART WITH 100"); err != nil {
		t.Fatal(err)
	}
	err := development.Run(context.Background(), db)
	if err == nil || !strings.Contains(err.Error(), "seed user_info: upsert user_info") {
		t.Fatalf("expected missing favorite team error, got %v", err)
	}
	var count int
	if err := db.QueryRow("SELECT (SELECT count(*) FROM teams) + (SELECT count(*) FROM users) + (SELECT count(*) FROM user_info)").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rollback failed: %d rows remain", count)
	}
}
