package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"preditto/internal/config"
)

type swaggerOperation struct {
	Security  []map[string][]string      `json:"security"`
	Responses map[string]json.RawMessage `json:"responses"`
}

func swaggerConfig(environment string) config.Config {
	return config.Config{Environment: environment, Auth: config.AuthConfig{
		AccessSecret: strings.Repeat("a", 32), RefreshSecret: strings.Repeat("r", 32),
		AccessTTL: 15 * time.Minute, ShortRefreshTTL: 12 * time.Hour, RefreshTTL: 720 * time.Hour,
	}}
}

// Exercise the actual application composition so new/removed routes cannot
// silently diverge from the checked-in Swagger document.
func TestSwaggerRouteCoverage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, err := newRouter(nil, swaggerConfig("development"))
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("specification: %d %s", w.Code, w.Body.String())
	}
	var document struct {
		Swagger             string                                     `json:"swagger"`
		Host                string                                     `json:"host"`
		BasePath            string                                     `json:"basePath"`
		Paths               map[string]map[string]swaggerOperation     `json:"paths"`
		SecurityDefinitions map[string]struct{ Type, In, Name string } `json:"securityDefinitions"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if document.Swagger != "2.0" || document.Host != "" || document.BasePath != "/" {
		t.Fatalf("expected Swagger 2.0 with same-origin URLs: %+v", document)
	}
	security := document.SecurityDefinitions["BearerAuth"]
	if security.Type != "apiKey" || security.In != "header" || security.Name != "Authorization" {
		t.Fatalf("invalid Bearer authorization definition: %+v", security)
	}
	protected := map[string]bool{
		"/api/user/get_user_info":              true,
		"/api/teams/get_favorite_team":         true,
		"/api/tournaments/get_all_tournaments": true,
		"/api/tournaments/select_tournament":   true,
	}
	for _, route := range r.Routes() {
		if route.Path == "/swagger/*any" {
			continue
		}
		method := strings.ToLower(route.Method)
		op, exists := document.Paths[route.Path][method]
		if !exists {
			t.Errorf("undocumented route: %s %s", route.Method, route.Path)
			continue
		}
		if len(op.Responses) == 0 {
			t.Errorf("missing response schemas: %s", route.Path)
		}
		secured := len(op.Security) == 1 && op.Security[0]["BearerAuth"] != nil
		if secured != protected[route.Path] {
			t.Errorf("incorrect security: %s", route.Path)
		}
		if protected[route.Path] {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(route.Method, route.Path, nil))
			if w.Code != 401 || w.Body.String() != `{"success":false,"error":"UNAUTHORIZED"}` {
				t.Errorf("protected route bypass: %s: %d %s", route.Path, w.Code, w.Body.String())
			}
		}
		delete(document.Paths[route.Path], method)
	}
	for path, operations := range document.Paths {
		for method := range operations {
			t.Errorf("documented route is not registered: %s %s", method, path)
		}
	}
}

func TestSwaggerEnvironmentGate(t *testing.T) {
	// Debug is Gin's default; it must not accidentally enable Swagger.
	gin.SetMode(gin.DebugMode)
	oldWriter := gin.DefaultWriter
	gin.DefaultWriter = io.Discard
	t.Cleanup(func() { gin.DefaultWriter = oldWriter; gin.SetMode(gin.TestMode) })
	for _, environment := range []string{"", "production", "staging", "test", "Development", " development", "development"} {
		t.Run(environment, func(t *testing.T) {
			r, err := newRouter(nil, swaggerConfig(environment))
			if err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{"/swagger/index.html", "/swagger/doc.json", "/swagger/swagger-ui-bundle.js", "/swagger/swagger-ui.css"} {
				w := httptest.NewRecorder()
				r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
				want := http.StatusNotFound
				if environment == "development" {
					want = http.StatusOK
				}
				if w.Code != want {
					t.Errorf("%s: %d, want %d", path, w.Code, want)
				}
				if environment == "development" && path == "/swagger/index.html" {
					for _, content := range []string{"Swagger UI", `url: "doc.json"`, "persistAuthorization: false"} {
						if !strings.Contains(strings.Join(strings.Fields(w.Body.String()), " "), content) {
							t.Errorf("UI missing %q", content)
						}
					}
				}
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
			if w.Code != 200 || w.Body.String() != `{"status":"ok"}` {
				t.Fatal("environment changed health contract")
			}
		})
	}
}
