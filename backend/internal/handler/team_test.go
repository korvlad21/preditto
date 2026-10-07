package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
	"preditto/internal/repository"
	"preditto/internal/router"
	"preditto/internal/service"
)

func teamRequest(api http.Handler, query, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/teams/get_all_teams"+query, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	return w
}

func TestGetAllTeamsInvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := gin.New()
	router.RegisterTeamRoutes(api.Group("/api"), handler.NewTeamHandler(nil), handler.NewCountryHandler(nil))
	for _, body := range []string{`{`, `[]`, `{"country":123}`, `{"country":true}`, `{"country":{}}`} {
		t.Run(body, func(t *testing.T) {
			assertError(t, teamRequest(api, "", body), http.StatusBadRequest, "INVALID_REQUEST")
		})
	}
	w := httptest.NewRecorder()
	api.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/teams/get_all_teams", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("GET status = %d, want 404", w.Code)
	}
}

func TestGetAllTeamsPostgres(t *testing.T) {
	db := postgres(t)
	execSQL(t, db, `
		INSERT INTO countries (name, short_name) VALUES ('United Kingdom', 'GBR');
		INSERT INTO teams (id, name, short_name, slug, country, logo_url) VALUES
			(11, 'British team', 'BRI', 'british-team', 'GBR', '/logos/teams/british-team.svg'),
			(12, 'Another test team', 'ATT', 'another-test-team', 'TST', '/logos/teams/another-test-team.svg');
	`)
	gin.SetMode(gin.TestMode)
	api := gin.New()
	router.RegisterTeamRoutes(api.Group("/api"), handler.NewTeamHandler(service.NewTeamService(repository.NewTeamRepository(db))), handler.NewCountryHandler(nil))

	all := []map[string]any{
		{"id": float64(12), "name": "Another test team", "short_name": "ATT", "slug": "another-test-team", "country": "TST", "logo_url": "/logos/teams/another-test-team.svg"},
		{"id": float64(11), "name": "British team", "short_name": "BRI", "slug": "british-team", "country": "GBR", "logo_url": "/logos/teams/british-team.svg"},
		{"id": float64(5), "name": "Test team", "short_name": nil, "slug": "test-team", "country": "TST", "logo_url": nil},
	}
	empty := []map[string]any{}
	for _, tc := range []struct {
		name, query, body string
		want              []map[string]any
	}{
		{"no body", "", "", all},
		{"omitted country", "", `{}`, all},
		{"empty country", "", `{"country":""}`, all},
		{"all countries", "", `{"country":"ALL"}`, all},
		{"matching country", "", `{"country":"TST"}`, []map[string]any{all[0], all[2]}},
		{"another country", "", `{"country":"GBR"}`, []map[string]any{all[1]}},
		{"unknown country", "", `{"country":"XYZ"}`, empty},
		{"case sensitive", "", `{"country":"tst"}`, empty},
		{"lowercase all", "", `{"country":"all"}`, empty},
		{"whitespace preserved", "", `{"country":" TST "}`, empty},
		{"SQL injection", "", `{"country":"TST' OR 1=1 --"}`, empty},
		{"query country", "?country=GBR", "", []map[string]any{all[1]}},
		{"empty query country", "?country=", "", all},
		{"query all", "?country=ALL", "", all},
		{"query unknown", "?country=XYZ", "", empty},
		{"query SQL injection", "?country=" + url.QueryEscape("TST' OR 1=1 --"), "", empty},
		{"JSON overrides query", "?country=GBR", `{"country":"TST"}`, []map[string]any{all[0], all[2]}},
		{"empty JSON overrides query", "?country=GBR", `{"country":""}`, all},
		{"query with empty object", "?country=GBR", `{}`, []map[string]any{all[1]}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := teamRequest(api, tc.query, tc.body)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d: %s", w.Code, w.Body.String())
			}
			var got []map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("teams = %v, want %v", got, tc.want)
			}
		})
	}

	execSQL(t, db, `DELETE FROM teams`)
	if w := teamRequest(api, "", `{}`); w.Code != http.StatusOK || strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatalf("empty table: %d %s", w.Code, w.Body.String())
	}
	execSQL(t, db, `DROP TABLE teams CASCADE`)
	assertError(t, teamRequest(api, "", `{}`), http.StatusInternalServerError, "INTERNAL_ERROR")
}
