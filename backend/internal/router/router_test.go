package router_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
	"preditto/internal/response"
	"preditto/internal/router"
)

func TestAPIFallbacks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var diagnostics bytes.Buffer
	oldWriter, oldErrorWriter := gin.DefaultWriter, gin.DefaultErrorWriter
	gin.DefaultWriter, gin.DefaultErrorWriter = &diagnostics, &diagnostics
	t.Cleanup(func() { gin.DefaultWriter, gin.DefaultErrorWriter = oldWriter, oldErrorWriter })
	r := router.New()
	r.POST("/api/known", func(c *gin.Context) { response.Success(c, http.StatusOK, nil) })
	r.GET("/api/panic", func(*gin.Context) { panic("private panic diagnostic") })
	r.GET("/api/error", func(c *gin.Context) { _ = c.Error(errors.New("private error diagnostic")) })
	r.GET("/api/status", func(c *gin.Context) { c.Status(http.StatusInternalServerError) })
	r.GET("/api/written-error", func(c *gin.Context) {
		response.Success(c, http.StatusAccepted, nil)
		_ = c.Error(errors.New("late diagnostic"))
	})
	r.GET("/api/written-panic", func(c *gin.Context) {
		response.Success(c, http.StatusAccepted, nil)
		panic("late panic diagnostic")
	})
	r.GET("/api/handled-error", func(c *gin.Context) {
		_ = c.Error(errors.New("handled diagnostic"))
		response.AbortError(c, http.StatusConflict, response.UsernameAlreadyExists)
	})
	hello := handler.NewHelloHandler()
	r.GET("/", gin.WrapF(hello.Index))
	r.GET("/hello", gin.WrapF(hello.Hello))
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/panic", func(*gin.Context) { panic("outside API") })
	r.StaticFS("/static", http.FS(fstest.MapFS{"example.txt": {Data: []byte("static content")}}))

	for _, tc := range []struct {
		method, path      string
		status            int
		body, contentType string
	}{
		{"POST", "/api/known", 200, `{"success":true,"data":null}`, "application/json; charset=utf-8"},
		{"GET", "/api/missing", 404, `{"success":false,"error":"NOT_FOUND"}`, "application/json; charset=utf-8"},
		{"POST", "/api/missing", 404, `{"success":false,"error":"NOT_FOUND"}`, "application/json; charset=utf-8"},
		{"GET", "/api/known", 404, `{"success":false,"error":"NOT_FOUND"}`, "application/json; charset=utf-8"},
		{"OPTIONS", "/api/known", 404, `{"success":false,"error":"NOT_FOUND"}`, "application/json; charset=utf-8"},
		{"PUT", "/api/known", 404, `{"success":false,"error":"NOT_FOUND"}`, "application/json; charset=utf-8"},
		{"GET", "/api/panic", 500, `{"success":false,"error":"INTERNAL_ERROR"}`, "application/json; charset=utf-8"},
		{"GET", "/api/error", 500, `{"success":false,"error":"INTERNAL_ERROR"}`, "application/json; charset=utf-8"},
		{"GET", "/api/status", 500, `{"success":false,"error":"INTERNAL_ERROR"}`, "application/json; charset=utf-8"},
		{"GET", "/api/written-error", 202, `{"success":true,"data":null}`, "application/json; charset=utf-8"},
		{"GET", "/api/written-panic", 202, `{"success":true,"data":null}`, "application/json; charset=utf-8"},
		{"GET", "/api/handled-error", 409, `{"success":false,"error":"USERNAME_ALREADY_EXISTS"}`, "application/json; charset=utf-8"},
		{"GET", "/", 200, "<div>Hello World</div>", "text/html; charset=utf-8"},
		{"GET", "/hello", 200, "{\"message\":\"Hello from Go!\"}\n", "application/json"},
		{"GET", "/health", 200, `{"status":"ok"}`, "application/json; charset=utf-8"},
		{"GET", "/static/example.txt", 200, "static content", "text/plain; charset=utf-8"},
		{"GET", "/missing", 404, "404 page not found", "text/plain"},
		{"GET", "/api", 404, "404 page not found", "text/plain"},
		{"GET", "/api-other/missing", 404, "404 page not found", "text/plain"},
		{"POST", "/health", 404, "404 page not found", "text/plain"},
		{"GET", "/panic", 500, "", ""},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != tc.status || w.Body.String() != tc.body || w.Header().Get("Content-Type") != tc.contentType {
				t.Fatalf("status=%d body=%s headers=%v", w.Code, w.Body.String(), w.Header())
			}
		})
	}
	for _, message := range []string{"private panic diagnostic", "private error diagnostic", "late diagnostic", "late panic diagnostic"} {
		if !strings.Contains(diagnostics.String(), message) {
			t.Errorf("missing server diagnostic: %s", message)
		}
	}

	// A real HTTP server suppresses HEAD bodies; status and JSON headers remain.
	server := httptest.NewServer(r)
	defer server.Close()
	res, err := server.Client().Head(server.URL + "/api/known")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body bytes.Buffer
	if _, err := body.ReadFrom(res.Body); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 404 || res.Header.Get("Content-Type") != "application/json; charset=utf-8" || body.Len() != 0 {
		t.Fatalf("HEAD response: %d %v %s", res.StatusCode, res.Header, body.String())
	}
}
