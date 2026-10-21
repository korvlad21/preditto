package response_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"preditto/internal/model"
	"preditto/internal/response"
)

func TestSuccessData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name string
		data any
		want string
	}{
		{"no data", nil, `{"success":true}`},
		{"false data", false, `{"success":true,"data":false}`},
		{"zero data", 0, `{"success":true,"data":0}`},
		{"empty string data", "", `{"success":true,"data":""}`},
		{"empty list", gin.H{"teams": []model.Team{}}, `{"success":true,"data":{"teams":[]}}`},
		{"nullable fields", gin.H{"avatar_url": (*string)(nil), "favorite_team_id": (*int64)(nil)}, `{"success":true,"data":{"avatar_url":null,"favorite_team_id":null}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			response.Success(c, http.StatusCreated, tc.data)
			if w.Code != 201 || w.Body.String() != tc.want {
				t.Fatalf("response: %d %s", w.Code, w.Body.String())
			}
		})
	}
}
