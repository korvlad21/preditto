package auth

import (
	"strings"
	"testing"
)

func TestRegisterValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*RegisterRequest)
		valid  bool
	}{
		{"valid without team", func(*RegisterRequest) {}, true},
		{"valid with team", func(r *RegisterRequest) { id := int64(5); r.FavoriteTeamID = &id }, true},
		{"missing username", func(r *RegisterRequest) { r.Username = "" }, false},
		{"short username", func(r *RegisterRequest) { r.Username = "ab" }, false},
		{"long username", func(r *RegisterRequest) { r.Username = strings.Repeat("a", 51) }, false},
		{"email as username", func(r *RegisterRequest) { r.Username = "me@example.com" }, false},
		{"missing first name", func(r *RegisterRequest) { r.FirstName = "" }, false},
		{"blank last name", func(r *RegisterRequest) { r.LastName = "   " }, false},
		{"long name", func(r *RegisterRequest) { r.FirstName = strings.Repeat("a", 101) }, false},
		{"invalid email", func(r *RegisterRequest) { r.Email = "invalid" }, false},
		{"missing password", func(r *RegisterRequest) { r.Password = "" }, false},
		{"short password", func(r *RegisterRequest) { r.Password = "short"; r.PasswordConfirmation = r.Password }, false},
		{"mismatch", func(r *RegisterRequest) { r.PasswordConfirmation = "different" }, false},
		{"missing confirmation", func(r *RegisterRequest) { r.PasswordConfirmation = "" }, false},
		{"bcrypt byte limit", func(r *RegisterRequest) {
			r.Password = strings.Repeat("\u00e9", 37)
			r.PasswordConfirmation = r.Password
		}, false},
		{"zero team", func(r *RegisterRequest) { id := int64(0); r.FavoriteTeamID = &id }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := RegisterRequest{Username: "korvlad21", FirstName: "Vladislav", LastName: "Korobkin", Email: "example@example.com", Password: "password", PasswordConfirmation: "password"}
			tc.change(&r)
			if got := r.Validate() == nil; got != tc.valid {
				t.Fatalf("valid=%v, want %v", got, tc.valid)
			}
		})
	}
}
