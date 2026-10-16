package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/security"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

type memoryRepository struct {
	Repository
	users        map[uuid.UUID]*User
	roles        map[string]Role
	revoked      []uuid.UUID
	keptSessions []string
	locks        int
	refreshSaved int
}

func newMemoryRepository() *memoryRepository {
	repository := &memoryRepository{users: map[uuid.UUID]*User{}, roles: map[string]Role{}}
	for _, definition := range access.Roles() {
		repository.roles[definition.Code] = Role{Code: definition.Code, Name: definition.Name}
		role := repository.roles[definition.Code]
		role.ID = uuid.New()
		repository.roles[definition.Code] = role
	}
	return repository
}

func (m *memoryRepository) addUser(roleCode string, active bool) uuid.UUID {
	user := &User{Name: "UJI-" + roleCode, Email: uuid.NewString() + "@uji.test", Active: active}
	user.ID = uuid.New()
	if role, ok := m.roles[roleCode]; ok {
		user.RoleID = &role.ID
	}
	m.users[user.ID] = user
	return user.ID
}

func (m *memoryRepository) roleCodeOf(user *User) *string {
	if user.RoleID == nil {
		return nil
	}
	for code, role := range m.roles {
		if role.ID == *user.RoleID {
			return &code
		}
	}
	return nil
}

func (m *memoryRepository) EmailExists(_ context.Context, email string) (bool, error) {
	for _, user := range m.users {
		if user.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (m *memoryRepository) CreateUserWithRole(ctx context.Context, user *User, chooseRole func(int64) string) error {
	if code := chooseRole(int64(len(m.users))); code != "" {
		role, err := m.FindRoleByCode(ctx, code)
		if err != nil {
			return err
		}
		user.RoleID = &role.ID
	}
	user.ID = uuid.New()
	m.users[user.ID] = user
	return nil
}

func (m *memoryRepository) FindUserByEmail(_ context.Context, email string) (User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return *user, nil
		}
	}
	return User{}, errUserNotFound
}

func (m *memoryRepository) FindUserByID(_ context.Context, id uuid.UUID) (User, error) {
	user, ok := m.users[id]
	if !ok {
		return User{}, errUserNotFound
	}
	return *user, nil
}

func (m *memoryRepository) FindRoleByID(_ context.Context, id uuid.UUID) (Role, error) {
	for _, role := range m.roles {
		if role.ID == id {
			return role, nil
		}
	}
	return Role{}, errRoleNotFound
}

func (m *memoryRepository) FindRoleByCode(_ context.Context, code string) (Role, error) {
	role, ok := m.roles[code]
	if !ok {
		return Role{}, errRoleNotFound
	}
	return role, nil
}

func (m *memoryRepository) StoreRefreshToken(context.Context, *RefreshToken) error {
	m.refreshSaved++
	return nil
}

func (m *memoryRepository) FindUserRow(_ context.Context, id uuid.UUID) (UserRow, error) {
	user, ok := m.users[id]
	if !ok {
		return UserRow{}, errUserNotFound
	}
	return UserRow{ID: user.ID, Name: user.Name, Email: user.Email, Active: user.Active, RoleCode: m.roleCodeOf(user)}, nil
}

func (m *memoryRepository) UpdateUserRole(_ context.Context, id, roleID uuid.UUID) error {
	m.users[id].RoleID = &roleID
	return nil
}

func (m *memoryRepository) UpdateUserActive(_ context.Context, id uuid.UUID, active bool) error {
	m.users[id].Active = active
	return nil
}

func (m *memoryRepository) CountActiveAdmins(context.Context) (int64, error) {
	var total int64
	for _, user := range m.users {
		if code := m.roleCodeOf(user); user.Active && code != nil && *code == access.RoleAdmin {
			total++
		}
	}
	return total, nil
}

func (m *memoryRepository) LockAdminChanges(context.Context) error {
	m.locks++
	return nil
}

func (m *memoryRepository) RevokeUserRefreshTokens(_ context.Context, userID uuid.UUID, _ time.Time) error {
	m.revoked = append(m.revoked, userID)
	return nil
}

func (m *memoryRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(m)
}

type forgetRecorder struct {
	forgotten []uuid.UUID
}

func (f *forgetRecorder) Forget(userID uuid.UUID) {
	f.forgotten = append(f.forgotten, userID)
}

func newTestService(repository *memoryRepository) (Service, *forgetRecorder) {
	tokens := token.NewManager(config.Token{AccessSecret: "auth-service-test", AccessTTL: time.Minute, RefreshTTL: time.Hour, Issuer: "auth-service-test"})
	cache := &forgetRecorder{}
	return NewService(repository, tokens, cache), cache
}

func assertAppError(t *testing.T, err error, status int, code string) {
	t.Helper()
	var failure *apperror.Error
	if !errors.As(err, &failure) || failure.Status != status || failure.Code != code {
		t.Fatalf("got %v, want %d %s", err, status, code)
	}
}

func boolPointer(value bool) *bool {
	return &value
}

func TestFirstUserBecomesAdminAndLaterUsersWaitForARole(t *testing.T) {
	if roleForNewcomer(0) != access.RoleAdmin || roleForNewcomer(1) != pendingRole || roleForNewcomer(40) != pendingRole {
		t.Fatal("bootstrap rule broken")
	}

	repository := newMemoryRepository()
	service, _ := newTestService(repository)

	first, err := service.Register(context.Background(), RegisterRequest{Name: "UJI Pertama", Email: "pertama@uji.test", Password: "rahasia123"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Register(context.Background(), RegisterRequest{Name: "UJI Kedua", Email: "kedua@uji.test", Password: "rahasia123"})
	if err != nil {
		t.Fatal(err)
	}

	if first.Role == nil || first.Role.Code != access.RoleAdmin || first.Permissions["users"] != "write" {
		t.Fatalf("first user got %+v", first)
	}
	if second.Role != nil || second.Permissions == nil || len(second.Permissions) != 0 {
		t.Fatalf("second user must wait for a role without any permission, got %+v", second)
	}
	for _, module := range access.Modules() {
		if access.Allows("", module, access.LevelRead) {
			t.Errorf("a user without role must not read %s", module)
		}
	}
}

func TestPendingUserCanLogInWithoutPermissions(t *testing.T) {
	repository := newMemoryRepository()
	service, _ := newTestService(repository)
	ctx := context.Background()
	if _, err := service.Register(ctx, RegisterRequest{Name: "UJI Admin", Email: "admin@uji.test", Password: "rahasia123"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Register(ctx, RegisterRequest{Name: "UJI Tamu", Email: "tamu@uji.test", Password: "rahasia123"}); err != nil {
		t.Fatal(err)
	}

	session, err := service.Login(ctx, LoginRequest{Email: "tamu@uji.test", Password: "rahasia123"})

	if err != nil {
		t.Fatal(err)
	}
	if session.Response.AccessToken == "" || session.RefreshToken == "" || repository.refreshSaved != 1 {
		t.Fatalf("pending user must still get a session, got %+v", session.Response)
	}
	if session.Response.User.Role != nil || len(session.Response.User.Permissions) != 0 {
		t.Fatalf("got %+v", session.Response.User)
	}
}

func TestAdminCanGiveAPendingUserARole(t *testing.T) {
	repository := newMemoryRepository()
	repository.addUser(access.RoleAdmin, true)
	pendingID := repository.addUser("", true)
	service, cache := newTestService(repository)

	updated, err := service.ChangeRole(context.Background(), pendingID, ChangeRoleRequest{Role: access.RoleSales})

	if err != nil {
		t.Fatal(err)
	}
	if updated.Role == nil || updated.Role.Code != access.RoleSales {
		t.Fatalf("got %+v", updated)
	}
	if len(cache.forgotten) != 1 || cache.forgotten[0] != pendingID {
		t.Fatalf("role cache not invalidated: %v", cache.forgotten)
	}
}

func TestRegisterFailsLoudlyWhenRolesAreNotSeeded(t *testing.T) {
	repository := newMemoryRepository()
	repository.roles = map[string]Role{}
	service, _ := newTestService(repository)

	_, err := service.Register(context.Background(), RegisterRequest{Name: "UJI", Email: "uji@uji.test", Password: "rahasia123"})

	assertAppError(t, err, http.StatusInternalServerError, "internal_error")
}

func TestRegisterRejectsPasswordsLongerThanBcryptCanRead(t *testing.T) {
	repository := newMemoryRepository()
	service, _ := newTestService(repository)

	_, err := service.Register(context.Background(), RegisterRequest{
		Name: "UJI", Email: "uji@uji.test", Password: strings.Repeat("é", 72),
	})

	assertAppError(t, err, http.StatusUnprocessableEntity, "password_too_long")
	if len(repository.users) != 0 {
		t.Fatalf("a user was created with a password bcrypt cannot hash")
	}
}

func TestProfileReturnsRoleAndPermissions(t *testing.T) {
	repository := newMemoryRepository()
	userID := repository.addUser(access.RoleFinance, true)
	service, _ := newTestService(repository)

	profile, err := service.Profile(context.Background(), userID)

	if err != nil {
		t.Fatal(err)
	}
	if profile.Role.Code != access.RoleFinance || profile.Role.Name != "Keuangan" {
		t.Fatalf("got role %+v", profile.Role)
	}
	if profile.Permissions["finance"] != "write" || profile.Permissions["sales"] != "read" || profile.Permissions["hr"] != "" {
		t.Fatalf("got permissions %v", profile.Permissions)
	}
}

func TestDeactivatedUserCannotLogInOrReadProfile(t *testing.T) {
	repository := newMemoryRepository()
	userID := repository.addUser(access.RoleSales, false)
	hash, _ := security.HashPassword("rahasia123")
	repository.users[userID].PasswordHash = hash
	service, _ := newTestService(repository)

	_, err := service.Login(context.Background(), LoginRequest{Email: repository.users[userID].Email, Password: "rahasia123"})
	assertAppError(t, err, http.StatusForbidden, "account_inactive")

	_, err = service.Profile(context.Background(), userID)
	assertAppError(t, err, http.StatusUnauthorized, "account_inactive")

	if repository.refreshSaved != 0 {
		t.Fatal("no session may be issued")
	}
}

func TestDeactivatedUserWithWrongPasswordStillGetsTheGenericError(t *testing.T) {
	repository := newMemoryRepository()
	userID := repository.addUser(access.RoleSales, false)
	hash, _ := security.HashPassword("rahasia123")
	repository.users[userID].PasswordHash = hash
	service, _ := newTestService(repository)

	_, err := service.Login(context.Background(), LoginRequest{Email: repository.users[userID].Email, Password: "salah12345"})

	assertAppError(t, err, http.StatusUnauthorized, "invalid_credential")
}

func TestLastActiveAdminCannotBeDemoted(t *testing.T) {
	repository := newMemoryRepository()
	adminID := repository.addUser(access.RoleAdmin, true)
	repository.addUser(access.RoleAdmin, false)
	service, cache := newTestService(repository)

	_, err := service.ChangeRole(context.Background(), adminID, ChangeRoleRequest{Role: access.RoleObserver})

	assertAppError(t, err, http.StatusUnprocessableEntity, "last_admin")
	if code := repository.roleCodeOf(repository.users[adminID]); *code != access.RoleAdmin {
		t.Fatalf("role changed to %s", *code)
	}
	if repository.locks != 1 || len(cache.forgotten) != 1 {
		t.Fatalf("got %d locks, %d cache invalidations", repository.locks, len(cache.forgotten))
	}
}

func TestLastActiveAdminCannotBeDeactivated(t *testing.T) {
	repository := newMemoryRepository()
	adminID := repository.addUser(access.RoleAdmin, true)
	service, _ := newTestService(repository)

	_, err := service.ChangeActive(context.Background(), adminID, ChangeActiveRequest{Active: boolPointer(false)})

	assertAppError(t, err, http.StatusUnprocessableEntity, "last_admin")
	if !repository.users[adminID].Active || len(repository.revoked) != 0 {
		t.Fatal("last admin must stay active")
	}
}

func TestAdminCanStepDownWhenAnotherAdminIsActive(t *testing.T) {
	repository := newMemoryRepository()
	adminID := repository.addUser(access.RoleAdmin, true)
	repository.addUser(access.RoleAdmin, true)
	service, cache := newTestService(repository)

	updated, err := service.ChangeRole(context.Background(), adminID, ChangeRoleRequest{Role: access.RoleFinance})

	if err != nil {
		t.Fatal(err)
	}
	if updated.Role == nil || updated.Role.Code != access.RoleFinance {
		t.Fatalf("got %+v", updated)
	}
	if len(cache.forgotten) != 1 || cache.forgotten[0] != adminID {
		t.Fatalf("role cache not invalidated: %v", cache.forgotten)
	}
}

func TestKeepingAdminRoleIsAlwaysAllowed(t *testing.T) {
	repository := newMemoryRepository()
	adminID := repository.addUser(access.RoleAdmin, true)
	service, _ := newTestService(repository)

	if _, err := service.ChangeRole(context.Background(), adminID, ChangeRoleRequest{Role: access.RoleAdmin}); err != nil {
		t.Fatal(err)
	}
}

func TestDeactivatingRevokesSessionsAndReactivatingRestores(t *testing.T) {
	repository := newMemoryRepository()
	repository.addUser(access.RoleAdmin, true)
	staffID := repository.addUser(access.RoleHR, true)
	service, _ := newTestService(repository)

	deactivated, err := service.ChangeActive(context.Background(), staffID, ChangeActiveRequest{Active: boolPointer(false)})
	if err != nil || deactivated.Active {
		t.Fatalf("got %+v, %v", deactivated, err)
	}
	if len(repository.revoked) != 1 || repository.revoked[0] != staffID {
		t.Fatalf("sessions not revoked: %v", repository.revoked)
	}

	reactivated, err := service.ChangeActive(context.Background(), staffID, ChangeActiveRequest{Active: boolPointer(true)})
	if err != nil || !reactivated.Active {
		t.Fatalf("got %+v, %v", reactivated, err)
	}
}

func TestChangeRoleRejectsUnknownRoleAndUser(t *testing.T) {
	repository := newMemoryRepository()
	service, _ := newTestService(repository)

	_, err := service.ChangeRole(context.Background(), uuid.New(), ChangeRoleRequest{Role: "tamu"})
	assertAppError(t, err, http.StatusUnprocessableEntity, "unknown_role")

	_, err = service.ChangeRole(context.Background(), uuid.New(), ChangeRoleRequest{Role: access.RoleSales})
	assertAppError(t, err, http.StatusNotFound, "user_not_found")
}
