package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"preditto/internal/model"
	"preditto/internal/repository"
)

type participantLoader func(context.Context, int64, int64) (model.ParticipantInfo, error)

func (f participantLoader) GetParticipantInfo(ctx context.Context, userID, tournamentID int64) (model.ParticipantInfo, error) {
	return f(ctx, userID, tournamentID)
}

func TestLoadParticipant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tournamentID := int64(201)
	user := &model.UserInfo{ID: 81, SelectedTournamentID: &tournamentID}
	participant := model.ParticipantInfo{ID: 7, JoinedAt: time.Date(2026, 1, 2, 3, 4, 5, 123456000, time.UTC)}
	for _, tc := range []struct {
		name  string
		user  *model.UserInfo
		err   error
		calls int
		code  int
		body  string
	}{
		{"missing user", nil, nil, 0, 401, `{"success":false,"error":"UNAUTHORIZED"}`},
		{"no selection", &model.UserInfo{ID: 81}, nil, 0, 404, `{"success":false,"error":"TOURNAMENT_NOT_SELECTED"}`},
		{"missing participant", user, repository.ErrNotFound, 1, 404, `{"success":false,"error":"TOURNAMENT_PARTICIPANT_NOT_FOUND"}`},
		{"database failure", user, errors.New("private database error"), 1, 500, `{"success":false,"error":"INTERNAL_ERROR"}`},
		{"success", user, nil, 1, 200, `{"id":7,"joined_at":"2026-01-02T03:04:05.123456Z"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
			calls := 0
			load := participantLoader(func(ctx context.Context, userID, selectedID int64) (model.ParticipantInfo, error) {
				calls++
				if ctx != req.Context() || userID != user.ID || selectedID != tournamentID {
					t.Fatalf("unexpected participant lookup for user %d, tournament %d", userID, selectedID)
				}
				return participant, tc.err
			})
			r := gin.New()
			r.GET("/", func(c *gin.Context) {
				if tc.user != nil {
					c.Set(userInfoKey, tc.user)
				}
			}, LoadParticipant(load), func(c *gin.Context) {
				got, ok := CurrentParticipant(c)
				if tc.code != http.StatusOK || !ok || *got != participant {
					t.Fatalf("unexpected participant in downstream handler: %v", got)
				}
				c.JSON(http.StatusOK, got)
			})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.code || w.Body.String() != tc.body || calls != tc.calls {
				t.Fatalf("status=%d body=%s calls=%d", w.Code, w.Body.String(), calls)
			}
		})
	}
}
