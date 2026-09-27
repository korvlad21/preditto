package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"preditto/internal/auth"
	"preditto/internal/config"
	authdto "preditto/internal/dto/auth"
	"preditto/internal/handler"
	"preditto/internal/middleware"
	"preditto/internal/repository"
	"preditto/internal/router"
	"preditto/internal/service"
)

func postgres(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("PREDITTO_AUTH_TEST_DSN")
	if dsn == "" {
		t.Skip("set PREDITTO_AUTH_TEST_DSN to a PostgreSQL test database URL")
	}
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		t.Fatal("PREDITTO_AUTH_TEST_DSN must be a PostgreSQL URL")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Close() })
	schema := fmt.Sprintf("auth_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE"); err != nil {
			t.Error(err)
		}
	})
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(8)
	t.Cleanup(func() { db.Close() })
	paths, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil || len(paths) == 0 {
		t.Fatal("migrations not found")
	}
	sort.Strings(paths)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(string(data)); err != nil {
			t.Fatalf("migration %s: %v", path, err)
		}
	}
	execSQL(t, db, `
		ALTER TABLE users ALTER COLUMN id RESTART WITH 81;
		INSERT INTO roles (id, code, name) VALUES (91, 'participant', 'Participant');
		INSERT INTO countries (name, short_name) VALUES ('Test country', 'TST');
		INSERT INTO teams (id, name, slug, country) VALUES (5, 'Test team', 'test-team', 'TST');
	`)
	return db
}

func testConfig() config.AuthConfig {
	return config.AuthConfig{AccessSecret: strings.Repeat("a", 32), RefreshSecret: strings.Repeat("r", 32), AccessTTL: 15 * time.Minute, RefreshTTL: 720 * time.Hour}
}

func testAPI(t *testing.T, db *sql.DB) (*gin.Engine, *auth.TokenManager) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	tokens, err := auth.NewTokenManager(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.NewAuthService(repository.NewAuthRepository(db), tokens)
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	router.RegisterAuthRoutes(r.Group("/api"), handler.NewAuthHandler(svc))
	r.GET("/protected", middleware.RequireAuth(tokens), func(c *gin.Context) { c.JSON(200, gin.H{"user_id": c.GetInt64(middleware.UserIDKey)}) })
	return r, tokens
}

func registration() authdto.RegisterRequest {
	id := int64(5)
	return authdto.RegisterRequest{Username: "korvlad21", FirstName: "Vladislav", LastName: "Korobkin", Email: "example@example.com", Password: "password", PasswordConfirmation: "password", FavoriteTeamID: &id}
}

func call(t *testing.T, api http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/auth/"+path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	return w
}

func register(t *testing.T, api http.Handler) authdto.AuthResponse {
	t.Helper()
	w := call(t, api, "register", registration())
	if w.Code != 201 {
		t.Fatalf("register: %d %s", w.Code, w.Body.String())
	}
	var result authdto.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func execSQL(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatal(err)
	}
}

func assertError(t *testing.T, w *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	var body struct {
		Error struct{ Code, Message string }
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if w.Code != status || body.Error.Code != code || body.Error.Message == "" {
		t.Fatalf("response: %d %s", w.Code, w.Body.String())
	}
	for _, private := range []string{"pq:", "password_hash", "token_hash", "SQLSTATE"} {
		if strings.Contains(w.Body.String(), private) {
			t.Fatalf("private data leaked: %s", w.Body.String())
		}
	}
}

func assertEmptyUsers(t *testing.T, db *sql.DB) {
	t.Helper()
	var count int
	if err := db.QueryRow(`SELECT (SELECT count(*) FROM users) + (SELECT count(*) FROM user_info) + (SELECT count(*) FROM user_roles) + (SELECT count(*) FROM refresh_tokens)`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rollback left %d rows", count)
	}
}

func TestRegisterPostgres(t *testing.T) {
	for _, tc := range []struct {
		name      string
		change    func(*authdto.RegisterRequest)
		duplicate bool
		status    int
		code      string
	}{
		{"success", func(*authdto.RegisterRequest) {}, false, 201, ""},
		{"without favorite team", func(r *authdto.RegisterRequest) { r.FavoriteTeamID = nil }, false, 201, ""},
		{"duplicate username", func(r *authdto.RegisterRequest) { r.Email = "another@example.com" }, true, 409, "USERNAME_ALREADY_EXISTS"},
		{"duplicate email", func(r *authdto.RegisterRequest) { r.Username = "another" }, true, 409, "EMAIL_ALREADY_EXISTS"},
		{"confirmation mismatch", func(r *authdto.RegisterRequest) { r.PasswordConfirmation = "different" }, false, 400, "INVALID_REQUEST"},
		{"invalid email", func(r *authdto.RegisterRequest) { r.Email = "invalid" }, false, 400, "INVALID_REQUEST"},
		{"short password", func(r *authdto.RegisterRequest) { r.Password = "short"; r.PasswordConfirmation = "short" }, false, 400, "INVALID_REQUEST"},
		{"missing team", func(r *authdto.RegisterRequest) { id := int64(9999); r.FavoriteTeamID = &id }, false, 400, "FAVORITE_TEAM_NOT_FOUND"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := postgres(t)
			api, tokens := testAPI(t, db)
			if tc.duplicate {
				register(t, api)
			}
			req := registration()
			tc.change(&req)
			w := call(t, api, "register", req)
			if tc.status != 201 {
				assertError(t, w, tc.status, tc.code)
				if !tc.duplicate {
					assertEmptyUsers(t, db)
				}
				return
			}
			if w.Code != 201 {
				t.Fatalf("status=%d %s", w.Code, w.Body.String())
			}
			var result authdto.AuthResponse
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.User.ID != 81 || result.User.Username != req.Username || result.User.Email != req.Email || result.User.FirstName != req.FirstName || result.User.LastName != req.LastName || result.User.Status != "ACTIVE" {
				t.Fatalf("incorrect user: %+v", result.User)
			}
			if (result.User.FavoriteTeamID == nil) != (req.FavoriteTeamID == nil) {
				t.Fatal("favorite team nullability changed")
			}
			if result.User.FavoriteTeamID != nil && *result.User.FavoriteTeamID != 5 {
				t.Fatal("wrong favorite team")
			}
			for _, secret := range []string{"password", "password_confirmation", "password_hash", "token_hash"} {
				if strings.Contains(w.Body.String(), `"`+secret+`"`) {
					t.Fatalf("leaked %s", secret)
				}
			}
			var hash string
			var roleID int64
			if err := db.QueryRow(`SELECT password_hash, role_id FROM users JOIN user_roles ON users.id=user_roles.user_id WHERE users.id=$1`, result.User.ID).Scan(&hash, &roleID); err != nil {
				t.Fatal(err)
			}
			if roleID != 91 || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
				t.Fatal("role or password hash incorrect")
			}
			claims, err := tokens.ParseRefreshToken(result.RefreshToken)
			if err != nil {
				t.Fatal(err)
			}
			var stored []byte
			if err := db.QueryRow(`SELECT token_hash FROM refresh_tokens WHERE id=$1 AND user_id=$2`, claims.ID, result.User.ID).Scan(&stored); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(stored, auth.HashRefreshToken(result.RefreshToken)) {
				t.Fatal("refresh token not hashed")
			}
			if _, err := tokens.ParseAccessToken(result.AccessToken); err != nil {
				t.Fatal(err)
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("missing no-store")
			}
		})
	}
}

func TestRegisterRollbackPostgres(t *testing.T) {
	for _, tc := range []struct{ name, setup string }{
		{"profile failure", `ALTER TABLE user_info ADD CONSTRAINT reject_profile CHECK (user_id < 0)`},
		{"missing participant role", `DELETE FROM roles WHERE code='participant'`},
		{"role assignment failure", `ALTER TABLE user_roles ADD CONSTRAINT reject_assignment CHECK (user_id < 0)`},
		{"session failure", `ALTER TABLE refresh_tokens ADD CONSTRAINT reject_session CHECK (user_id < 0)`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := postgres(t)
			api, _ := testAPI(t, db)
			execSQL(t, db, tc.setup)
			assertError(t, call(t, api, "register", registration()), 500, "INTERNAL_ERROR")
			assertEmptyUsers(t, db)
		})
	}
}

func TestLoginPostgres(t *testing.T) {
	db := postgres(t)
	api, _ := testAPI(t, db)
	user := register(t, api)
	for _, login := range []string{user.User.Username, user.User.Email} {
		w := call(t, api, "login", authdto.LoginRequest{Login: login, Password: "password"})
		if w.Code != 200 {
			t.Fatalf("login %s: %d %s", login, w.Code, w.Body.String())
		}
		var result authdto.AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.User.ID != user.User.ID || result.User.PasswordHash != "" || result.AccessToken == "" || result.RefreshToken == "" {
			t.Fatal("invalid login response")
		}
	}
	wrong := call(t, api, "login", authdto.LoginRequest{Login: user.User.Username, Password: "wrong-password"})
	unknown := call(t, api, "login", authdto.LoginRequest{Login: "unknown", Password: "password"})
	assertError(t, wrong, 401, "INVALID_CREDENTIALS")
	assertError(t, unknown, 401, "INVALID_CREDENTIALS")
	if wrong.Body.String() != unknown.Body.String() || !strings.Contains(wrong.Body.String(), "Invalid login or password") {
		t.Fatal("credential errors differ")
	}
	execSQL(t, db, `UPDATE users SET status='BLOCKED' WHERE id=$1`, user.User.ID)
	assertError(t, call(t, api, "login", authdto.LoginRequest{Login: user.User.Username, Password: "password"}), 403, "USER_BLOCKED")
	assertError(t, call(t, api, "login", authdto.LoginRequest{Login: user.User.Username, Password: "wrong-password"}), 401, "INVALID_CREDENTIALS")
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM refresh_tokens`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("failed login issued sessions: %d", count)
	}
}

func TestRefreshAndLogoutPostgres(t *testing.T) {
	db := postgres(t)
	api, tokens := testAPI(t, db)
	user := register(t, api)
	w := call(t, api, "refresh", authdto.RefreshRequest{RefreshToken: user.RefreshToken})
	if w.Code != 200 {
		t.Fatalf("refresh: %d %s", w.Code, w.Body.String())
	}
	var pair authdto.TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &pair); err != nil {
		t.Fatal(err)
	}
	if pair.RefreshToken == user.RefreshToken || pair.AccessToken == user.AccessToken || pair.ExpiresIn != 900 {
		t.Fatal("tokens were not rotated")
	}
	if _, err := tokens.ParseAccessToken(pair.AccessToken); err != nil {
		t.Fatal(err)
	}
	assertError(t, call(t, api, "refresh", authdto.RefreshRequest{RefreshToken: user.RefreshToken}), 401, "INVALID_TOKEN")
	for i := 0; i < 2; i++ {
		if w := call(t, api, "logout", authdto.RefreshRequest{RefreshToken: pair.RefreshToken}); w.Code != 204 || w.Body.Len() != 0 {
			t.Fatalf("logout: %d %s", w.Code, w.Body.String())
		}
	}
	assertError(t, call(t, api, "refresh", authdto.RefreshRequest{RefreshToken: pair.RefreshToken}), 401, "INVALID_TOKEN")
	// Logout revokes only its refresh session, leaving access valid until expiry.
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w = httptest.NewRecorder()
	api.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal("logout unexpectedly invalidated access token")
	}
}

func TestRefreshRejectionsPostgres(t *testing.T) {
	for _, name := range []string{"expired JWT", "invalid JWT", "access as refresh", "wrong type correct key", "missing session", "revoked session", "wrong hash", "wrong user", "expired session", "blocked user"} {
		t.Run(name, func(t *testing.T) {
			db := postgres(t)
			api, tokens := testAPI(t, db)
			user := register(t, api)
			raw := user.RefreshToken
			claims, err := tokens.ParseRefreshToken(raw)
			if err != nil {
				t.Fatal(err)
			}
			status, code := 401, "INVALID_TOKEN"
			switch name {
			case "expired JWT":
				claims.IssuedAt = jwt.NewNumericDate(time.Now().Add(-2 * time.Hour))
				claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
				raw, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testConfig().RefreshSecret))
			case "invalid JWT":
				raw = "not.a.jwt"
			case "access as refresh":
				raw = user.AccessToken
			case "wrong type correct key":
				claims.TokenType = "access"
				raw, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testConfig().RefreshSecret))
			case "missing session":
				execSQL(t, db, `DELETE FROM refresh_tokens`)
			case "revoked session":
				execSQL(t, db, `UPDATE refresh_tokens SET revoked_at=CURRENT_TIMESTAMP`)
			case "wrong hash":
				execSQL(t, db, `UPDATE refresh_tokens SET token_hash=$1`, make([]byte, 32))
			case "wrong user":
				execSQL(t, db, `INSERT INTO users (id,username,email,password_hash,status) VALUES (99,'other','other@example.com','unused','ACTIVE'); UPDATE refresh_tokens SET user_id=99`)
			case "expired session":
				execSQL(t, db, `UPDATE refresh_tokens SET created_at=CURRENT_TIMESTAMP-INTERVAL '2 hours',expires_at=CURRENT_TIMESTAMP-INTERVAL '1 hour'`)
			case "blocked user":
				execSQL(t, db, `UPDATE users SET status='BLOCKED'`)
				status, code = 403, "USER_BLOCKED"
			}
			if err != nil {
				t.Fatal(err)
			}
			assertError(t, call(t, api, "refresh", authdto.RefreshRequest{RefreshToken: raw}), status, code)
			var count int
			if err := db.QueryRow(`SELECT count(*) FROM refresh_tokens`).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count > 1 {
				t.Fatal("rejected refresh issued a new session")
			}
		})
	}
}

func TestRotationRollbackPostgres(t *testing.T) {
	db := postgres(t)
	api, tokens := testAPI(t, db)
	user := register(t, api)
	claims, err := tokens.ParseRefreshToken(user.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	execSQL(t, db, `ALTER TABLE refresh_tokens ADD CONSTRAINT reject_new_session CHECK (id = '`+claims.ID+`')`)
	assertError(t, call(t, api, "refresh", authdto.RefreshRequest{RefreshToken: user.RefreshToken}), 500, "INTERNAL_ERROR")
	var revoked bool
	if err := db.QueryRow(`SELECT revoked_at IS NOT NULL FROM refresh_tokens WHERE id=$1`, claims.ID).Scan(&revoked); err != nil {
		t.Fatal(err)
	}
	if revoked {
		t.Fatal("failed rotation consumed old token")
	}
	execSQL(t, db, `ALTER TABLE refresh_tokens DROP CONSTRAINT reject_new_session`)
	if w := call(t, api, "refresh", authdto.RefreshRequest{RefreshToken: user.RefreshToken}); w.Code != 200 {
		t.Fatalf("old token unusable after rollback: %d", w.Code)
	}
}

func TestConcurrentRotationPostgres(t *testing.T) {
	db := postgres(t)
	api, _ := testAPI(t, db)
	user := register(t, api)
	var wg sync.WaitGroup
	start := make(chan struct{})
	codes := make(chan int, 8)
	body, err := json.Marshal(authdto.RefreshRequest{RefreshToken: user.RefreshToken})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			api.ServeHTTP(w, req)
			codes <- w.Code
		}()
	}
	close(start)
	wg.Wait()
	close(codes)
	success, rejected := 0, 0
	for code := range codes {
		switch code {
		case 200:
			success++
		case 401:
			rejected++
		default:
			t.Errorf("unexpected status: %d", code)
		}
	}
	if success != 1 || rejected != 7 {
		t.Fatalf("successful=%d rejected=%d", success, rejected)
	}
	var active, total int
	if err := db.QueryRow(`SELECT count(*) FILTER (WHERE revoked_at IS NULL),count(*) FROM refresh_tokens`).Scan(&active, &total); err != nil {
		t.Fatal(err)
	}
	if active != 1 || total != 2 {
		t.Fatalf("active=%d total=%d", active, total)
	}
}

func TestInvalidRequestsPostgres(t *testing.T) {
	db := postgres(t)
	api, _ := testAPI(t, db)
	for _, path := range []string{"register", "login", "refresh", "logout"} {
		for _, body := range []string{"{", `{}`, `null`, strings.Repeat(" ", 17<<10) + `{}`} {
			req := httptest.NewRequest(http.MethodPost, "/api/auth/"+path, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			api.ServeHTTP(w, req)
			assertError(t, w, 400, "INVALID_REQUEST")
		}
	}
	assertEmptyUsers(t, db)
	assertError(t, call(t, api, "logout", authdto.RefreshRequest{RefreshToken: "invalid"}), 401, "INVALID_TOKEN")
}

func TestRefreshMigrationDownUpPostgres(t *testing.T) {
	db := postgres(t)
	for _, direction := range []string{"down", "up"} {
		data, err := os.ReadFile("../../migrations/000009_create_refresh_tokens." + direction + ".sql")
		if err != nil {
			t.Fatal(err)
		}
		execSQL(t, db, string(data))
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM refresh_tokens`).Scan(&count); err != nil {
		t.Fatal(err)
	}
}
