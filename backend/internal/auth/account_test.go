package auth

import (
	"context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/cookie"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/security"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

const currentPassword = "rahasia123"

func (m *memoryRepository) UpdateUserName(_ context.Context, id uuid.UUID, name string) error {
	m.users[id].Name = name
	return nil
}

func (m *memoryRepository) UpdateUserPassword(_ context.Context, id uuid.UUID, passwordHash string) error {
	m.users[id].PasswordHash = passwordHash
	return nil
}

func (m *memoryRepository) RevokeOtherRefreshTokens(_ context.Context, userID uuid.UUID, keptHash string, _ time.Time) error {
	m.revoked = append(m.revoked, userID)
	m.keptSessions = append(m.keptSessions, keptHash)
	return nil
}

func userWithPassword(t *testing.T, repository *memoryRepository, roleCode string, active bool) uuid.UUID {
	t.Helper()
	userID := repository.addUser(roleCode, active)
	hash, err := security.HashPassword(currentPassword)
	if err != nil {
		t.Fatal(err)
	}
	repository.users[userID].PasswordHash = hash
	return userID
}

func TestChangePasswordKeepsTheCurrentSessionAndRevokesTheOthers(t *testing.T) {
	repository := newMemoryRepository()
	userID := userWithPassword(t, repository, access.RoleSales, true)
	service, _ := newTestService(repository)

	err := service.ChangePassword(context.Background(), userID, "refresh-sesi-ini",
		ChangePasswordRequest{CurrentPassword: currentPassword, NewPassword: "sandi-baru-456"})

	if err != nil {
		t.Fatal(err)
	}
	if !security.PasswordMatches(repository.users[userID].PasswordHash, "sandi-baru-456") {
		t.Fatal("the new password was not stored")
	}
	if len(repository.revoked) != 1 || repository.revoked[0] != userID {
		t.Fatalf("other sessions not revoked: %v", repository.revoked)
	}
	if repository.keptSessions[0] != token.HashRefresh("refresh-sesi-ini") {
		t.Fatalf("the current session must be kept, got %q", repository.keptSessions[0])
	}
}

func TestChangePasswordWithoutSessionCookieRevokesEverySession(t *testing.T) {
	repository := newMemoryRepository()
	userID := userWithPassword(t, repository, access.RoleSales, true)
	service, _ := newTestService(repository)

	err := service.ChangePassword(context.Background(), userID, "",
		ChangePasswordRequest{CurrentPassword: currentPassword, NewPassword: "sandi-baru-456"})

	if err != nil || len(repository.keptSessions) != 1 || repository.keptSessions[0] != "" {
		t.Fatalf("got %v, kept %v", err, repository.keptSessions)
	}
}

func TestChangePasswordRejectsWrongOrUnchangedPassword(t *testing.T) {
	cases := []struct {
		request ChangePasswordRequest
		code    string
	}{
		{ChangePasswordRequest{CurrentPassword: "salah-12345", NewPassword: "sandi-baru-456"}, "wrong_current_password"},
		{ChangePasswordRequest{CurrentPassword: currentPassword, NewPassword: currentPassword}, "same_password"},
	}
	for _, tc := range cases {
		repository := newMemoryRepository()
		userID := userWithPassword(t, repository, access.RoleSales, true)
		before := repository.users[userID].PasswordHash
		service, _ := newTestService(repository)

		err := service.ChangePassword(context.Background(), userID, "refresh-sesi-ini", tc.request)

		assertAppError(t, err, http.StatusUnprocessableEntity, tc.code)
		if repository.users[userID].PasswordHash != before || len(repository.revoked) != 0 {
			t.Fatalf("%s: nothing may change", tc.code)
		}
	}
}

func TestWrongCurrentPasswordHasAnIndonesianMessage(t *testing.T) {
	if errWrongPassword.Message != "Kata sandi saat ini salah" {
		t.Fatalf("got %q", errWrongPassword.Message)
	}
}

func TestDeactivatedUserCannotChangePasswordOrProfile(t *testing.T) {
	repository := newMemoryRepository()
	userID := userWithPassword(t, repository, access.RoleSales, false)
	service, _ := newTestService(repository)

	err := service.ChangePassword(context.Background(), userID, "", ChangePasswordRequest{CurrentPassword: currentPassword, NewPassword: "sandi-baru-456"})
	assertAppError(t, err, http.StatusUnauthorized, "account_inactive")

	_, err = service.UpdateProfile(context.Background(), userID, UpdateProfileRequest{Name: "UJI Nama"})
	assertAppError(t, err, http.StatusUnauthorized, "account_inactive")
}

func TestUpdateProfileTrimsTheNameAndReturnsThePermissions(t *testing.T) {
	repository := newMemoryRepository()
	userID := repository.addUser(access.RoleFinance, true)
	service, _ := newTestService(repository)

	profile, err := service.UpdateProfile(context.Background(), userID, UpdateProfileRequest{Name: "  UJI Nama Baru  "})

	if err != nil {
		t.Fatal(err)
	}
	if profile.Name != "UJI Nama Baru" || repository.users[userID].Name != "UJI Nama Baru" {
		t.Fatalf("got %q, stored %q", profile.Name, repository.users[userID].Name)
	}
	if profile.Role == nil || profile.Role.Code != access.RoleFinance || profile.Permissions["finance"] != "write" {
		t.Fatalf("got %+v", profile)
	}
}

func TestUpdateProfileRejectsNamesOutsideTwoToHundredCharacters(t *testing.T) {
	for _, name := range []string{"   a   ", strings.Repeat("é", 101)} {
		repository := newMemoryRepository()
		userID := repository.addUser(access.RoleFinance, true)
		original := repository.users[userID].Name
		service, _ := newTestService(repository)

		_, err := service.UpdateProfile(context.Background(), userID, UpdateProfileRequest{Name: name})

		assertAppError(t, err, http.StatusUnprocessableEntity, "invalid_name")
		if repository.users[userID].Name != original {
			t.Fatal("the name must not change")
		}
	}

	repository := newMemoryRepository()
	userID := repository.addUser(access.RoleFinance, true)
	service, _ := newTestService(repository)
	if _, err := service.UpdateProfile(context.Background(), userID, UpdateProfileRequest{Name: strings.Repeat("é", 100)}); err != nil {
		t.Fatalf("100 characters must be accepted, got %v", err)
	}
}

func TestRevokeOtherRefreshTokensKeepsOnlyTheGivenHash(t *testing.T) {
	db, recorder := databasetest.Open(t, databasetest.Result{Affected: 2})
	userID := uuid.New()

	if err := NewRepository(db).RevokeOtherRefreshTokens(context.Background(), userID, "hash-sesi-ini", time.Now()); err != nil {
		t.Fatal(err)
	}
	queries := recorder.Queries()
	if len(queries) != 1 {
		t.Fatalf("got %d queries", len(queries))
	}
	assertContains(t, queries[0].SQL, `UPDATE "refresh_tokens" SET "revoked_at"=$1`, "user_id = $", "AND revoked_at IS NULL AND token_hash <> $")
	if args := queries[0].Args; args[len(args)-1] != "hash-sesi-ini" {
		t.Fatalf("got args %v", queries[0].Args)
	}
}

const testRefreshCookie = "uji_refresh"

func userRow(t *testing.T, userID uuid.UUID) databasetest.Result {
	t.Helper()
	hash, err := security.HashPassword(currentPassword)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	return databasetest.Result{
		Columns: []string{"id", "name", "email", "password_hash", "role_id", "active", "last_login_at", "created_at", "updated_at"},
		Rows:    [][]driver.Value{{userID.String(), "UJI Tamu", "uji-tamu@uji.test", hash, nil, true, nil, now, now}},
	}
}

func serveAccount(t *testing.T, method, path, body string, withToken bool, results ...databasetest.Result) (*httptest.ResponseRecorder, *databasetest.Recorder, uuid.UUID) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, recorder := databasetest.Open(t, results...)
	tokens := token.NewManager(config.Token{AccessSecret: "auth-account-test", AccessTTL: time.Minute, RefreshTTL: time.Hour, Issuer: "auth-account-test"})
	refreshCookie := cookie.NewRefreshWriter(config.Cookie{Name: testRefreshCookie, Path: "/api/auth"})
	module := NewModule(db, tokens, refreshCookie)

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	module.RegisterRoutes(router.Group("/api", middleware.Authorization(tokens, module.Subjects())))

	userID := uuid.MustParse("6f1c3c1e-0f53-4c1e-9b7a-1f2d3c4b5a69")
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: testRefreshCookie, Value: "refresh-sesi-ini"})
	if withToken {
		issued, err := tokens.IssueAccess(userID)
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Authorization", "Bearer "+issued.Value)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response, recorder, userID
}

func TestAccountRoutesRequireAnAccessToken(t *testing.T) {
	for _, path := range []string{"/api/auth/me", "/api/auth/password"} {
		response, recorder, _ := serveAccount(t, http.MethodPatch, path, `{"name":"UJI Nama"}`, false)
		if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), "missing_access_token") {
			t.Errorf("%s: got %d %s", path, response.Code, response.Body.String())
		}
		if len(recorder.Queries()) != 0 {
			t.Errorf("%s: got %d queries", path, len(recorder.Queries()))
		}
	}
}

func TestChangePasswordRouteKeepsTheSessionFromTheCookieForAUserWithoutRole(t *testing.T) {
	userID := uuid.MustParse("6f1c3c1e-0f53-4c1e-9b7a-1f2d3c4b5a69")
	response, recorder, _ := serveAccount(t, http.MethodPatch, "/api/auth/password",
		`{"currentPassword":"rahasia123","newPassword":"sandi-baru-456"}`, true,
		userRow(t, userID), databasetest.Result{Affected: 1}, databasetest.Result{Affected: 3})

	if response.Code != http.StatusNoContent {
		t.Fatalf("got %d %s", response.Code, response.Body.String())
	}
	queries := recorder.Queries()
	if len(queries) != 5 || queries[1].SQL != databasetest.Begin || queries[4].SQL != databasetest.Commit {
		t.Fatalf("got %+v", queries)
	}
	assertContains(t, queries[2].SQL, `UPDATE "users" SET "password_hash"`)
	assertContains(t, queries[3].SQL, `UPDATE "refresh_tokens"`, "token_hash <> $")
	if args := queries[3].Args; args[len(args)-1] != token.HashRefresh("refresh-sesi-ini") {
		t.Fatalf("kept %v", queries[3].Args)
	}
}

func TestChangePasswordRouteAnswersWrongPasswordWith422(t *testing.T) {
	userID := uuid.MustParse("6f1c3c1e-0f53-4c1e-9b7a-1f2d3c4b5a69")
	response, recorder, _ := serveAccount(t, http.MethodPatch, "/api/auth/password",
		`{"currentPassword":"salah-12345","newPassword":"sandi-baru-456"}`, true, userRow(t, userID))

	if response.Code != http.StatusUnprocessableEntity || !strings.Contains(response.Body.String(), "Kata sandi saat ini salah") {
		t.Fatalf("got %d %s", response.Code, response.Body.String())
	}
	if len(recorder.Queries()) != 1 {
		t.Fatalf("got %d queries", len(recorder.Queries()))
	}
}

func TestChangePasswordRouteValidatesTheNewPassword(t *testing.T) {
	for _, body := range []string{
		`{"currentPassword":"rahasia123","newPassword":"pendek"}`,
		`{"currentPassword":"rahasia123","newPassword":"` + strings.Repeat("a", 73) + `"}`,
		`{"newPassword":"sandi-baru-456"}`,
	} {
		response, recorder, _ := serveAccount(t, http.MethodPatch, "/api/auth/password", body, true)
		if response.Code != http.StatusBadRequest || len(recorder.Queries()) != 0 {
			t.Errorf("%s: got %d after %d queries", body, response.Code, len(recorder.Queries()))
		}
	}
}

func TestUpdateProfileRouteReturnsTheProfile(t *testing.T) {
	userID := uuid.MustParse("6f1c3c1e-0f53-4c1e-9b7a-1f2d3c4b5a69")
	response, recorder, _ := serveAccount(t, http.MethodPatch, "/api/auth/me", `{"name":"  UJI Nama Baru "}`, true,
		userRow(t, userID), databasetest.Result{Affected: 1})

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"name":"UJI Nama Baru"`) {
		t.Fatalf("got %d %s", response.Code, response.Body.String())
	}
	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries", len(queries))
	}
	assertContains(t, queries[1].SQL, `UPDATE "users" SET "name"=$1`)
}
