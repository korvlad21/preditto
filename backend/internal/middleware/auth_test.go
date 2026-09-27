package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"preditto/internal/auth"
	"preditto/internal/config"
)

func TestRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.AuthConfig{AccessSecret: strings.Repeat("a", 32), RefreshSecret: strings.Repeat("r", 32), AccessTTL: 15 * time.Minute, RefreshTTL: 720 * time.Hour}
	m, err := auth.NewTokenManager(cfg)
	if err != nil {
		t.Fatal(err)
	}
	pair, _, err := m.GeneratePair(81)
	if err != nil {
		t.Fatal(err)
	}
	signed := func(kind, key string, expires time.Time) string {
		c := auth.Claims{TokenType: kind, RegisteredClaims: jwt.RegisteredClaims{Subject: "81", ID: "test", IssuedAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), ExpiresAt: jwt.NewNumericDate(expires)}}
		raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(key))
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	for _, tc := range []struct {
		name, header string
		want         int
	}{
		{"valid", "Bearer " + pair.AccessToken, 204},
		{"case insensitive scheme", "bearer " + pair.AccessToken, 204},
		{"missing", "", 401}, {"empty bearer", "Bearer", 401},
		{"basic", "Basic " + pair.AccessToken, 401}, {"extra field", "Bearer " + pair.AccessToken + " extra", 401},
		{"malformed jwt", "Bearer invalid", 401},
		{"expired", "Bearer " + signed("access", cfg.AccessSecret, time.Now().Add(-time.Minute)), 401},
		{"refresh as access", "Bearer " + pair.RefreshToken, 401},
		{"wrong type correct secret", "Bearer " + signed("refresh", cfg.AccessSecret, time.Now().Add(time.Minute)), 401},
		{"invalid signature", "Bearer " + signed("access", strings.Repeat("z", 32), time.Now().Add(time.Minute)), 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			called := false
			r.GET("/protected", RequireAuth(m), func(c *gin.Context) {
				called = true
				if c.GetInt64(UserIDKey) != 81 {
					t.Error("user ID not set")
				}
				c.Status(204)
			})
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.want || called != (tc.want == 204) {
				t.Fatalf("status=%d called=%v", w.Code, called)
			}
			if tc.want == 401 && w.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Fatal("missing challenge")
			}
		})
	}
}
