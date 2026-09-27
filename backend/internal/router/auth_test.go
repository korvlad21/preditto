package router_test

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
	"preditto/internal/router"
)

func TestRegisterAuthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router.RegisterAuthRoutes(engine.Group("/api"), handler.NewAuthHandler(nil))

	var got []string
	for _, route := range engine.Routes() {
		got = append(got, route.Method+" "+route.Path)
	}
	sort.Strings(got)
	want := []string{
		"POST /api/auth/login",
		"POST /api/auth/logout",
		"POST /api/auth/refresh",
		"POST /api/auth/register",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("auth routes = %v, want %v", got, want)
	}

	for _, path := range []string{"login", "logout", "refresh", "register"} {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/"+path, strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest || w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("Pragma") != "no-cache" {
			t.Fatalf("%s: status=%d, headers=%v", path, w.Code, w.Header())
		}
	}
}
