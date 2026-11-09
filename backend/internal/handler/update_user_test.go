package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"

	"preditto/internal/auth"
	"preditto/internal/handler"
	"preditto/internal/middleware"
	"preditto/internal/model"
	"preditto/internal/repository"
	"preditto/internal/router"
	"preditto/internal/service"
)

func updateUserAPI(t *testing.T, db *sql.DB, directory string, loader middleware.UserInfoLoader, observers ...gin.HandlerFunc) (*gin.Engine, *auth.TokenManager) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	tokens, err := auth.NewTokenManager(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	users := service.NewUserService(repository.NewUserRepository(db), service.NewAvatarStore(directory))
	if loader == nil {
		loader = users
	}
	api := router.New()
	chain := append([]gin.HandlerFunc{middleware.RequireAuth(tokens), middleware.LoadUser(loader)}, observers...)
	router.RegisterUserRoutes(api.Group("/api", chain...), handler.NewUserHandler(users))
	return api, tokens
}

func updateToken(t *testing.T, tokens *auth.TokenManager, id int64) string {
	t.Helper()
	pair, _, err := tokens.GeneratePair(id, nil)
	if err != nil {
		t.Fatal(err)
	}
	return pair.AccessToken
}

func updateUserRequest(api http.Handler, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, "/api/user/update_user?user_id=81", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	return w
}

func uploadUserAvatar(t *testing.T, api http.Handler, token, body string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	if err := writer.WriteField("data", body); err != nil {
		t.Fatal(err)
	}
	file, err := writer.CreateFormFile("avatar", "../../untrusted.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/user/update_user", &buffer)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	return w
}

func TestUpdateUserValidationAndAccess(t *testing.T) {
	api, tokens := updateUserAPI(t, nil, t.TempDir(), selectionUserLoader{})
	token := updateToken(t, tokens, 81)
	for _, badToken := range []string{"", "invalid"} {
		assertError(t, updateUserRequest(api, badToken, `{}`), 401, "UNAUTHORIZED")
	}
	pair, _, _ := tokens.GeneratePair(81, nil)
	assertError(t, updateUserRequest(api, pair.RefreshToken, `{}`), 401, "UNAUTHORIZED")
	for _, body := range []string{
		``, `{}`, `null`, `[]`, `{"user_id":0,"email":"a@example.com"}`,
		`{"user_id":-1,"email":"a@example.com"}`, `{"user_id":"81","email":"a@example.com"}`,
		`{"user_id":81.5,"email":"a@example.com"}`, `{"user_id":9223372036854775808,"email":"a@example.com"}`,
		`{"user_id":81}`, `{"user_id":81,"email":null}`, `{"user_id":81,"email":"invalid"}`,
		`{"user_id":81,"email":"a@example.com","favorite_team_id":-1}`,
		`{"user_id":81,"email":"a@example.com","favorite_team_id":"5"}`,
		`{"user_id":81,"email":"a@example.com","first_name":1}`,
		`{"user_id":81,"email":"a@example.com","avatar_url":false}`,
		`{"user_id":81,"email":"a@example.com","last_name":"\u0000"}`,
		`{"user_id":81,"email":"a@example.com","roles":["admin"]}`,
		`{"user_id":81,"email":"a@example.com"} {}`,
		`{"user_id":81,"email":"a@example.com","first_name":"` + strings.Repeat("я", 101) + `"}`,
		`{"user_id":81,"email":"a@example.com","avatar_url":"` + strings.Repeat("x", 501) + `"}`,
		"{\"user_id\":81,\"email\":\"a@example.com\",\"first_name\":\"\xff\"}",
	} {
		assertError(t, updateUserRequest(api, token, body), 400, "INVALID_REQUEST")
	}
	assertError(t, updateUserRequest(api, token, `{"user_id":82,"email":"a@example.com"}`), 403, "FORBIDDEN")
	assertError(t, updateUserRequest(api, token, strings.Repeat(" ", 17<<10)), 413, "INVALID_REQUEST")
	broken := httptest.NewRequest(http.MethodPut, "/api/user/update_user", nil)
	broken.Header.Set("Content-Type", "application/json")
	broken.Header.Set("Authorization", "Bearer "+token)
	broken.Body = io.NopCloser(brokenUpdateBody{})
	w := httptest.NewRecorder()
	api.ServeHTTP(w, broken)
	assertError(t, w, 400, "INVALID_REQUEST")
	for _, contentType := range []string{"text/plain", "", "multipart/form-data; boundary=broken"} {
		req := httptest.NewRequest(http.MethodPut, "/api/user/update_user", strings.NewReader("invalid"))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		status := 415
		if strings.HasPrefix(contentType, "multipart") {
			status = 400
		}
		assertError(t, w, status, "INVALID_REQUEST")
	}
	apiWithoutContext := router.New()
	router.RegisterUserRoutes(apiWithoutContext.Group("/api"), handler.NewUserHandler(nil))
	assertError(t, updateUserRequest(apiWithoutContext, token, `{}`), 401, "UNAUTHORIZED")
}

type brokenUpdateBody struct{}

func (brokenUpdateBody) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func seedUpdateUsers(t *testing.T, db *sql.DB) {
	t.Helper()
	execSQL(t, db, `
		INSERT INTO users (id, username, email, password_hash, status) VALUES
		(81, 'owner', 'owner@example.com', 'private-hash', 'ACTIVE'),
		(82, 'admin', 'admin@example.com', 'admin-hash', 'ACTIVE'),
		(83, 'no_profile', 'no_profile@example.com', 'other-hash', 'BLOCKED');
		INSERT INTO tournaments (id, name, slug) VALUES (201, 'Selected', 'selected');
		INSERT INTO user_info (user_id, first_name, last_name, favorite_team_id, selected_tournament_id)
		VALUES (81, 'First', 'Last', 5, 201);
		INSERT INTO roles (id, code, name) VALUES (92, 'admin', 'Admin');
		INSERT INTO user_roles (user_id, role_id) VALUES (81, 91), (82, 92);
	`)
}

func readUpdateUser(t *testing.T, db *sql.DB, id int64) model.UserInfo {
	t.Helper()
	user, err := repository.NewUserRepository(db).GetUserInfo(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return user
}

func sameUpdateUser(a, b model.UserInfo) bool {
	a.CreatedAt, b.CreatedAt = a.CreatedAt.UTC(), b.CreatedAt.UTC()
	return reflect.DeepEqual(a, b)
}

func updatedUser(t *testing.T, w *httptest.ResponseRecorder) model.UserInfo {
	t.Helper()
	if w.Code != 200 {
		t.Fatalf("update: %d %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("missing no-store")
	}
	var user model.UserInfo
	if err := json.Unmarshal(successData(t, w, ""), &user); err != nil {
		t.Fatal(err)
	}
	return user
}

func TestUpdateUserPostgres(t *testing.T) {
	db := postgres(t)
	seedUpdateUsers(t, db)
	var before, after *model.UserInfo
	api, tokens := updateUserAPI(t, db, t.TempDir(), nil, func(c *gin.Context) {
		before, _ = middleware.CurrentUser(c)
		c.Next()
		after, _ = middleware.CurrentUser(c)
	})
	owner, admin := updateToken(t, tokens, 81), updateToken(t, tokens, 82)
	initial := readUpdateUser(t, db, 81)
	if user := updatedUser(t, updateUserRequest(api, owner, `{"user_id":81,"email":"owner@example.com"}`)); !sameUpdateUser(user, initial) {
		t.Fatal("email-only update changed existing profile fields")
	}
	user := updatedUser(t, updateUserRequest(api, owner, `{"user_id":81,"email":"owner@example.com","first_name":"Иван","last_name":"","favorite_team_id":0}`))
	if before != after || !sameUpdateUser(*after, user) || *user.FirstName != "Иван" || *user.LastName != "" || user.FavoriteTeamID != nil || *user.SelectedTournamentID != 201 || !user.CreatedAt.Equal(initial.CreatedAt) || !reflect.DeepEqual(user.Roles, initial.Roles) {
		t.Fatalf("incorrect update or snapshot: %+v", user)
	}
	for _, tc := range []struct {
		body   string
		status int
		code   string
	}{
		{`{"user_id":81,"email":"admin@example.com","first_name":"Uncommitted"}`, 409, "EMAIL_ALREADY_EXISTS"},
		{`{"user_id":81,"email":"admin@example.com","avatar_url":"/users/untrusted.jpg"}`, 409, "EMAIL_ALREADY_EXISTS"},
		{`{"user_id":81,"email":"rollback@example.com","favorite_team_id":999}`, 400, "FAVORITE_TEAM_NOT_FOUND"},
		{`{"user_id":81,"email":"rollback@example.com","avatar_url":"../../outside.jpg"}`, 400, "INVALID_REQUEST"},
		{`{"user_id":81,"email":"rollback@example.com","avatar_url":"https://example.com/avatar.jpg"}`, 400, "INVALID_REQUEST"},
	} {
		assertError(t, updateUserRequest(api, owner, tc.body), tc.status, tc.code)
		if !sameUpdateUser(readUpdateUser(t, db, 81), user) || !sameUpdateUser(*after, user) {
			t.Fatal("failed update changed database or snapshot")
		}
	}
	assertError(t, updateUserRequest(api, owner, `{"user_id":83,"email":"forbidden@example.com"}`), 403, "FORBIDDEN")
	assertError(t, updateUserRequest(api, admin, `{"user_id":999,"email":"missing@example.com"}`), 404, "USER_NOT_FOUND")
	missingProfile := readUpdateUser(t, db, 83)
	assertError(t, updateUserRequest(api, admin, `{"user_id":83,"email":"rollback@example.com","avatar_url":"/users/untrusted.jpg"}`), 400, "INVALID_REQUEST")
	assertError(t, updateUserRequest(api, admin, `{"user_id":83,"email":"rollback@example.com","favorite_team_id":999}`), 400, "FAVORITE_TEAM_NOT_FOUND")
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM user_info WHERE user_id=83`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("profile insert did not roll back: %d %v", count, err)
	}
	if !sameUpdateUser(readUpdateUser(t, db, 83), missingProfile) {
		t.Fatal("failed update of missing profile changed the account")
	}
	// An email-only update creates an absent profile with nullable fields intact.
	adminBefore := readUpdateUser(t, db, 82)
	if created := updatedUser(t, updateUserRequest(api, admin, `{"user_id":82,"email":"admin@example.com"}`)); !sameUpdateUser(created, adminBefore) {
		t.Fatal("email-only profile creation changed nullable fields")
	}
	if err := db.QueryRow(`SELECT count(*) FROM user_info WHERE user_id=82`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("missing profile was not created: %d %v", count, err)
	}
	other := updatedUser(t, updateUserRequest(api, admin, `{"user_id":83,"email":"changed@example.com","first_name":"Other","favorite_team_id":5}`))
	if other.ID != 83 || other.Status != "BLOCKED" || *other.FavoriteTeamID != 5 || other.LastName != nil || before != after || after.ID != 82 || after.Email != "admin@example.com" {
		t.Fatal("admin update corrupted target or caller")
	}
	user = updatedUser(t, updateUserRequest(api, owner, `{"user_id":81,"email":"owner@example.com","first_name":null,"favorite_team_id":null}`))
	if *user.FirstName != "Иван" || *user.LastName != "" || user.FavoriteTeamID != nil {
		t.Fatal("omitted/null optional fields changed")
	}
	var username, hash, status string
	if err := db.QueryRow(`SELECT username,password_hash,status FROM users WHERE id=81`).Scan(&username, &hash, &status); err != nil || username != "owner" || hash != "private-hash" || status != "ACTIVE" {
		t.Fatal("protected account fields changed", err)
	}
	// A database-side profile error must roll back the preceding email write.
	execSQL(t, db, `CREATE FUNCTION reject_profile_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'private database detail'; END $$;
		CREATE TRIGGER reject_update BEFORE UPDATE ON user_info FOR EACH ROW EXECUTE FUNCTION reject_profile_update();`)
	assertError(t, updateUserRequest(api, owner, `{"user_id":81,"email":"failed@example.com"}`), 500, "INTERNAL_ERROR")
	if !sameUpdateUser(readUpdateUser(t, db, 81), user) || !sameUpdateUser(*after, user) {
		t.Fatal("database failure did not roll back")
	}
}

func avatarPNG(t *testing.T) []byte {
	t.Helper()
	picture := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	picture.Set(0, 0, color.NRGBA{R: 255, A: 255})
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, picture); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestUpdateUserAvatarPostgres(t *testing.T) {
	db := postgres(t)
	seedUpdateUsers(t, db)
	directory := t.TempDir()
	api, tokens := updateUserAPI(t, db, directory, nil)
	token := updateToken(t, tokens, 81)
	body := `{"user_id":81,"email":"owner@example.com"}`
	picture := avatarPNG(t)
	user := updatedUser(t, uploadUserAvatar(t, api, token, body, picture))
	pattern := regexp.MustCompile(`^/users/[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}\.jpg$`)
	if user.AvatarURL == nil || !pattern.MatchString(*user.AvatarURL) {
		t.Fatalf("unsafe avatar URL: %v", user.AvatarURL)
	}
	stored, err := os.ReadFile(filepath.Join(directory, filepath.Base(*user.AvatarURL)))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := jpeg.Decode(bytes.NewReader(stored))
	if err != nil || decoded.Bounds() != image.Rect(0, 0, 3, 2) {
		t.Fatalf("avatar is not a JPEG: %v", err)
	}
	if !sameUpdateUser(user, readUpdateUser(t, db, 81)) {
		t.Fatal("URL not persisted")
	}
	if unchanged := updatedUser(t, updateUserRequest(api, token, body)); !sameUpdateUser(unchanged, user) {
		t.Fatal("omitted avatar was not preserved")
	}
	keep, _ := json.Marshal(map[string]any{"user_id": 81, "email": "owner@example.com", "avatar_url": *user.AvatarURL})
	updatedUser(t, updateUserRequest(api, token, string(keep)))
	otherAvatar, _ := json.Marshal(map[string]any{"user_id": 83, "email": "no_profile@example.com", "avatar_url": *user.AvatarURL})
	assertError(t, updateUserRequest(api, updateToken(t, tokens, 83), string(otherAvatar)), 400, "INVALID_REQUEST")
	second := updatedUser(t, uploadUserAvatar(t, api, token, body, stored))
	if *user.AvatarURL == *second.AvatarURL {
		t.Fatal("avatar UUID reused")
	}
	// Reject excessive dimensions before allocating image pixels.
	huge := append([]byte(nil), picture...)
	binary.BigEndian.PutUint32(huge[16:20], 50000)
	binary.BigEndian.PutUint32(huge[29:33], crc32.ChecksumIEEE(huge[12:29]))
	for _, data := range [][]byte{[]byte("not an image"), picture[:len(picture)/2], []byte("GIF89a"), huge} {
		assertError(t, uploadUserAvatar(t, api, token, body, data), 400, "INVALID_REQUEST")
	}
	assertError(t, uploadUserAvatar(t, api, token, body, bytes.Repeat([]byte{'x'}, service.MaxAvatarBytes+1)), 413, "INVALID_REQUEST")
	assertError(t, uploadUserAvatar(t, api, token, body, bytes.Repeat([]byte{'x'}, service.MaxAvatarBytes+(32<<10))), 413, "INVALID_REQUEST")
	assertError(t, uploadUserAvatar(t, api, token, `{"user_id":82,"email":"admin@example.com"}`, picture), 403, "FORBIDDEN")
	assertError(t, uploadUserAvatar(t, api, token, `{"user_id":81,"email":"admin@example.com"}`, picture), 409, "EMAIL_ALREADY_EXISTS")
	assertError(t, uploadUserAvatar(t, api, token, `{"user_id":81,"email":"changed@example.com","favorite_team_id":999}`, picture), 400, "FAVORITE_TEAM_NOT_FOUND")
	assertError(t, uploadUserAvatar(t, api, token, string(keep), picture), 400, "INVALID_REQUEST")
	files, err := os.ReadDir(directory)
	if err != nil || len(files) != 2 || !sameUpdateUser(second, readUpdateUser(t, db, 81)) {
		t.Fatal("failed upload left a file or modified database", err)
	}
	// A filesystem error occurs before database changes.
	blockedDirectory := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blockedDirectory, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	blockedAPI, _ := updateUserAPI(t, db, blockedDirectory, nil)
	assertError(t, uploadUserAvatar(t, blockedAPI, token, `{"user_id":81,"email":"unwritten@example.com"}`, picture), 500, "INTERNAL_ERROR")
	if !sameUpdateUser(second, readUpdateUser(t, db, 81)) {
		t.Fatal("filesystem failure modified database")
	}
	cleared := updatedUser(t, updateUserRequest(api, token, `{"user_id":81,"email":"owner@example.com","avatar_url":""}`))
	if cleared.AvatarURL != nil {
		t.Fatal("avatar was not cleared")
	}
}

func TestUpdateUserConcurrentEmailPostgres(t *testing.T) {
	db := postgres(t)
	seedUpdateUsers(t, db)
	api, tokens := updateUserAPI(t, db, t.TempDir(), nil)
	owner, admin := updateToken(t, tokens, 81), updateToken(t, tokens, 82)
	var wait sync.WaitGroup
	results := make(chan *httptest.ResponseRecorder, 2)
	for _, target := range []struct{ token, body string }{
		{owner, `{"user_id":81,"email":"shared@example.com"}`},
		{admin, `{"user_id":83,"email":"shared@example.com"}`},
	} {
		wait.Add(1)
		go func(token, body string) {
			defer wait.Done()
			results <- updateUserRequest(api, token, body)
		}(target.token, target.body)
	}
	wait.Wait()
	close(results)
	statuses := map[int]int{}
	for result := range results {
		statuses[result.Code]++
		if result.Code == 409 {
			assertError(t, result, 409, "EMAIL_ALREADY_EXISTS")
		}
	}
	if statuses[200] != 1 || statuses[409] != 1 {
		t.Fatalf("concurrent email results: %v", statuses)
	}
}
