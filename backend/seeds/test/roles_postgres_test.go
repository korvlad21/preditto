package test

import (
	"context"
	"testing"

	"preditto/internal/repository"
	"preditto/seeds/development"
)

func TestRolesAndAssignmentsPostgres(t *testing.T) {
	db := seedPostgres(t)
	ctx := context.Background()
	if _, err := db.Exec(`
		ALTER TABLE users ALTER COLUMN id RESTART WITH 101;
		INSERT INTO roles (id, code, name) VALUES (81, 'admin', 'Old name');
		ALTER TABLE roles ALTER COLUMN id RESTART WITH 201;
	`); err != nil {
		t.Fatal(err)
	}
	for run := 0; run < 2; run++ {
		if err := development.Run(ctx, db); err != nil {
			t.Fatal(err)
		}
		var roles, assignments int
		if err := db.QueryRow(`SELECT (SELECT count(*) FROM roles), (SELECT count(*) FROM user_roles)`).Scan(&roles, &assignments); err != nil {
			t.Fatal(err)
		}
		if roles != 2 || assignments != 3 {
			t.Fatalf("run %d: roles=%d assignments=%d", run, roles, assignments)
		}
		for _, want := range []struct {
			username, code, name string
			userID, roleID       int64
		}{
			{"korvlad21", "admin", "Администратор", 101, 81},
			{"hanna_stoma", "admin", "Администратор", 102, 81},
			{"boroda", "participant", "Участник", 103, 202},
		} {
			var userID, roleID int64
			var code, name string
			if err := db.QueryRow(`
				SELECT u.id, r.id, r.code, r.name FROM users u
				JOIN user_roles ur ON ur.user_id = u.id
				JOIN roles r ON r.id = ur.role_id WHERE u.username = $1
			`, want.username).Scan(&userID, &roleID, &code, &name); err != nil {
				t.Fatal(err)
			}
			if userID != want.userID || roleID != want.roleID || code != want.code || name != want.name {
				t.Fatalf("incorrect assignment for %s: %d %d %s %s", want.username, userID, roleID, code, name)
			}
		}
	}
	// Both operations accept a DB or a caller-owned transaction.
	for i := 0; i < 2; i++ {
		if err := repository.RemoveUserRole(ctx, db, 101, 81); err != nil {
			t.Fatal(err)
		}
	}
	assertCount := func(want int) {
		t.Helper()
		var got int
		if err := db.QueryRow("SELECT count(*) FROM user_roles").Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("assignments=%d, want %d", got, want)
		}
	}
	assertCount(2)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	for i := 0; i < 2; i++ {
		if err := repository.AssignUserRole(ctx, tx, 101, 81); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	assertCount(2)
	if err := development.Run(ctx, db); err != nil {
		t.Fatal(err)
	}
	assertCount(3)
	if err := repository.AssignUserRole(ctx, db, -1, 81); err == nil {
		t.Fatal("expected missing user foreign key error")
	}
	if err := repository.AssignUserRole(ctx, db, 101, -1); err == nil {
		t.Fatal("expected missing role foreign key error")
	}
}

func TestUserRoleFailureRollsBackPostgres(t *testing.T) {
	db := seedPostgres(t)
	if _, err := db.Exec("ALTER TABLE user_roles ADD CONSTRAINT reject_assignments CHECK (user_id < 0)"); err != nil {
		t.Fatal(err)
	}
	if err := development.Run(context.Background(), db); err == nil {
		t.Fatal("expected assignment error")
	}
	var count int
	if err := db.QueryRow(`SELECT (SELECT count(*) FROM teams) + (SELECT count(*) FROM users)
		+ (SELECT count(*) FROM user_info) + (SELECT count(*) FROM roles) + (SELECT count(*) FROM user_roles)`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rollback left %d rows", count)
	}
}
