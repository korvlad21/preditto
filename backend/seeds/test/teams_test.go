package test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"

	"preditto/seeds/development"
)

// The connector keeps the test driver local without global registration.
type teamsConnector struct{ conn *teamsConn }

func (c teamsConnector) Connect(context.Context) (driver.Conn, error) { return c.conn, nil }
func (c teamsConnector) Driver() driver.Driver                        { return teamsDriver{} }

type teamsDriver struct{}

func (teamsDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("use the test connector")
}

type teamsConn struct {
	rows                        [][]driver.NamedValue
	committed, rolledBack, fail bool
	users                       map[string]int64
	roles                       map[string]int64
	assignments                 [][2]int64
	failAssignment              bool
	failCountry                 bool
}

func (*teamsConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepared statements are not supported")
}
func (*teamsConn) Close() error                { return nil }
func (c *teamsConn) Begin() (driver.Tx, error) { return c, nil }
func (c *teamsConn) Commit() error             { c.committed = true; return nil }
func (c *teamsConn) Rollback() error           { c.rolledBack = true; return nil }
func (c *teamsConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if strings.Contains(query, "INSERT INTO countries") {
		if c.failCountry {
			return nil, errors.New("injected country failure")
		}
		return driver.RowsAffected(1), nil
	}
	if strings.Contains(query, "INSERT INTO roles") {
		if c.roles == nil {
			c.roles = make(map[string]int64)
		}
		c.roles[args[0].Value.(string)] = int64(70 + len(c.roles))
		return driver.RowsAffected(1), nil
	}
	if strings.Contains(query, "INSERT INTO user_roles") {
		if c.failAssignment {
			return nil, errors.New("injected assignment failure")
		}
		c.assignments = append(c.assignments, [2]int64{args[0].Value.(int64), args[1].Value.(int64)})
		return driver.RowsAffected(1), nil
	}
	if strings.Contains(query, "INSERT INTO users") {
		if c.users == nil {
			c.users = make(map[string]int64)
		}
		c.users[args[0].Value.(string)] = int64(40 + len(c.users))
		return driver.RowsAffected(1), nil
	}
	if strings.Contains(query, "INSERT INTO user_info") {
		return driver.RowsAffected(1), nil
	}
	if !strings.Contains(query, "INSERT INTO teams") || !strings.Contains(query, "ON CONFLICT (slug) DO UPDATE") {
		return nil, errors.New("unexpected seed query")
	}
	if !strings.Contains(query, "logo_url = EXCLUDED.logo_url") {
		return nil, errors.New("team upsert must update the local logo path")
	}
	if c.fail {
		return nil, errors.New("injected write failure")
	}
	c.rows = append(c.rows, append([]driver.NamedValue(nil), args...))
	return driver.RowsAffected(1), nil
}

func (c *teamsConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(query, "SELECT id FROM roles") {
		id, found := c.roles[args[0].Value.(string)]
		return &userIDRows{id: id, found: found}, nil
	}
	if !strings.Contains(query, "SELECT id FROM users") {
		return nil, errors.New("unexpected seed query")
	}
	id, found := c.users[args[0].Value.(string)]
	return &userIDRows{id: id, found: found}, nil
}

type userIDRows struct {
	id    int64
	found bool
}

func (*userIDRows) Columns() []string { return []string{"id"} }
func (*userIDRows) Close() error      { return nil }
func (r *userIDRows) Next(values []driver.Value) error {
	if !r.found {
		return io.EOF
	}
	values[0] = r.id
	r.found = false
	return nil
}

func TestSeedTeams(t *testing.T) {
	conn := &teamsConn{}
	db := sql.OpenDB(teamsConnector{conn})
	t.Cleanup(func() { db.Close() })
	if err := development.Run(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	if !conn.committed || conn.rolledBack {
		t.Fatal("successful seed must commit without rollback")
	}
	if len(conn.rows) != 36 {
		t.Fatalf("expected 36 seed teams, got %d", len(conn.rows))
	}
	seen := make(map[string]bool)
	for _, row := range conn.rows {
		if len(row) != 5 {
			t.Fatalf("expected name, short_name, slug, country and logo_url, got %v", row)
		}
		name, nameOK := row[0].Value.(string)
		shortName, shortOK := row[1].Value.(string)
		slug, slugOK := row[2].Value.(string)
		country, countryOK := row[3].Value.(string)
		if logo, ok := row[4].Value.(string); !ok || logo != "/logos/teams/"+slug+".svg" {
			t.Fatalf("invalid local logo_url for %q: %v", slug, row[4].Value)
		}
		if !nameOK || !shortOK || !slugOK || !countryOK || name == "" || slug == "" || len(shortName) != 3 || len(country) != 3 {
			t.Fatalf("invalid seed team: %v", row)
		}
		for _, letter := range shortName {
			if letter < 'A' || letter > 'Z' {
				t.Fatalf("invalid short_name: %q", shortName)
			}
		}
		if seen[slug] {
			t.Fatalf("duplicate seed slug: %q", slug)
		}
		seen[slug] = true
	}
}

func TestSeedTeamsRollsBackOnWriteError(t *testing.T) {
	conn := &teamsConn{fail: true}
	db := sql.OpenDB(teamsConnector{conn})
	t.Cleanup(func() { db.Close() })
	err := development.Run(context.Background(), db)
	if err == nil || !strings.Contains(err.Error(), "seed teams: upsert team") {
		t.Fatalf("expected wrapped teams write error, got %v", err)
	}
	if conn.committed || !conn.rolledBack {
		t.Fatal("failed seed must roll back without commit")
	}
}

func TestSeedUserRolesUsesResolvedIDs(t *testing.T) {
	conn := &teamsConn{}
	db := sql.OpenDB(teamsConnector{conn})
	t.Cleanup(func() { db.Close() })
	if err := development.Run(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	want := [][2]int64{{40, 70}, {41, 70}, {42, 71}}
	if len(conn.assignments) != len(want) {
		t.Fatalf("assignments=%v, want %v", conn.assignments, want)
	}
	for i := range want {
		if conn.assignments[i] != want[i] {
			t.Fatalf("assignment=%v, want %v", conn.assignments[i], want[i])
		}
	}
}

func TestSeedUserRolesRollsBackOnWriteError(t *testing.T) {
	conn := &teamsConn{failAssignment: true}
	db := sql.OpenDB(teamsConnector{conn})
	t.Cleanup(func() { db.Close() })
	err := development.Run(context.Background(), db)
	if err == nil || !strings.Contains(err.Error(), "seed user_roles: assign role") {
		t.Fatalf("expected assignment error, got %v", err)
	}
	if conn.committed || !conn.rolledBack {
		t.Fatal("assignment failure must roll back the entire seed transaction")
	}
}
