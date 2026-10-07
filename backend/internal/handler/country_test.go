package handler_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
	"preditto/internal/repository"
	"preditto/internal/router"
	"preditto/internal/service"
)

func countryAPI(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	api := gin.New()
	countries := handler.NewCountryHandler(service.NewCountryService(repository.NewCountryRepository(db)))
	router.RegisterTeamRoutes(api.Group("/api"), handler.NewTeamHandler(nil), countries)
	return api
}

func countryRequest(api http.Handler, query, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/teams/get_all_countries"+query, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	return w
}

func TestGetAllCountriesDatabaseError(t *testing.T) {
	db, err := sql.Open("postgres", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	api := countryAPI(db)
	w := countryRequest(api, "", "")
	assertError(t, w, http.StatusInternalServerError, "INTERNAL_ERROR")
	if strings.Contains(w.Body.String(), "database is closed") {
		t.Fatalf("database error leaked: %s", w.Body.String())
	}

	w = httptest.NewRecorder()
	api.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/teams/get_all_countries", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("GET status = %d, want 404", w.Code)
	}
}

func TestGetAllCountriesPostgres(t *testing.T) {
	db := postgres(t)
	execSQL(t, db, `
		UPDATE countries SET created_at = '2026-10-07 12:34:56.123456';
		INSERT INTO countries (id, name, short_name, created_at) VALUES
			(11, 'United Kingdom', 'GBR', '2026-10-07 01:02:03.654321'),
			(12, 'Another country', 'ATC', '2026-10-07 00:00:00');
	`)
	api := countryAPI(db)
	want := []map[string]any{
		{"id": float64(12), "name": "Another country", "short_name": "ATC", "created_at": "2026-10-07T00:00:00Z"},
		{"id": float64(1), "name": "Test country", "short_name": "TST", "created_at": "2026-10-07T12:34:56.123456Z"},
		{"id": float64(11), "name": "United Kingdom", "short_name": "GBR", "created_at": "2026-10-07T01:02:03.654321Z"},
	}
	for _, tc := range []struct{ name, query, body string }{
		{"no body", "", ""},
		{"empty object", "", `{}`},
		{"no filtering", "?country=GBR", `{"country":"GBR"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := countryRequest(api, tc.query, tc.body)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d: %s", w.Code, w.Body.String())
			}
			var got []map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("countries = %v, want %v", got, want)
			}
		})
	}

	execSQL(t, db, `DELETE FROM teams; DELETE FROM countries`)
	if w := countryRequest(api, "", ""); w.Code != http.StatusOK || strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatalf("empty table: %d %s", w.Code, w.Body.String())
	}
	execSQL(t, db, `DROP TABLE countries CASCADE`)
	assertError(t, countryRequest(api, "", ""), http.StatusInternalServerError, "INTERNAL_ERROR")
}
