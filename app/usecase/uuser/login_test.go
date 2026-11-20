package uuser

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/adapter/database/repository/user"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/password"

	"gorm.io/gorm"
)

const correctPassword = "correct-password"

// fakeLoginRepo records the writes Login makes. Unimplemented methods panic
// on the nil embedded interface.
type fakeLoginRepo struct {
	user.Repository

	user *model.UserSession

	transactions  int
	loginFailed   []int
	statusUpdates []bool
	revokes       int
	activities    []string
	sessions      int
}

func (f *fakeLoginRepo) FindUserByUserName(context.Context, string) (*model.UserSession, error) {
	if f.user == nil {
		return nil, gorm.ErrRecordNotFound
	}

	copied := *f.user
	return &copied, nil
}

func (f *fakeLoginRepo) DoTransaction(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	f.transactions++
	return fn(ctx, nil)
}

func (f *fakeLoginRepo) UpdateLoginFailed(_ context.Context, _ string, n int) error {
	f.loginFailed = append(f.loginFailed, n)
	return nil
}

func (f *fakeLoginRepo) UpdateStatus(_ context.Context, _ string, active bool) error {
	f.statusUpdates = append(f.statusUpdates, active)
	return nil
}

func (f *fakeLoginRepo) RevokeUserSessions(context.Context, string, time.Time, string) error {
	f.revokes++
	return nil
}

func (f *fakeLoginRepo) CreateActivity(_ context.Context, a *model.UserActivity) error {
	f.activities = append(f.activities, a.Activity)
	return nil
}

func (f *fakeLoginRepo) UpdateLastLogin(context.Context, string, time.Time) error {
	return nil
}

func (f *fakeLoginRepo) CreateSession(context.Context, *model.Session) error {
	f.sessions++
	return nil
}

// fakeBcrypt replaces the slow bcrypt functions and counts comparisons.
func fakeBcrypt(t *testing.T) *int {
	t.Helper()

	checks := 0
	origCheck, origCreate := password.CheckPassword, password.CreatePassword

	password.CheckPassword = func(plain, hash string) (bool, error) {
		checks++
		if plain == correctPassword && hash == "hash-of-correct-password" {
			return true, nil
		}

		return false, errors.New("mismatch")
	}
	password.CreatePassword = func(string) (string, error) {
		return "dummy-hash", nil
	}

	t.Cleanup(func() {
		password.CheckPassword, password.CreatePassword = origCheck, origCreate
	})

	return &checks
}

func newLoginUser(active bool, loginFailed int) *model.UserSession {
	return &model.UserSession{
		UserName:    "alice",
		UserUUID:    "user-1",
		Password:    "hash-of-correct-password",
		RoleName:    "admin",
		Active:      active,
		LoginFailed: loginFailed,
	}
}

func login(repo *fakeLoginRepo, pass string) (*string, *string, error) {
	u := NewUsecase(Expire{Token: time.Minute, Refresh: time.Hour}, testJWT, repo)

	return u.Login(context.Background(), request.Login{
		UserName:  "alice",
		Password:  pass,
		SessionID: "session-1",
		IPAddress: "127.0.0.1",
	})
}

func assertInvalidCredentials(t *testing.T, err error) {
	t.Helper()

	var apiErr response.APIError
	if !errors.As(err, &apiErr) ||
		apiErr.StatusCode != http.StatusUnauthorized ||
		apiErr.Message != "Invalid username or password" {
		t.Fatalf("err = %#v, want generic 401 invalid credentials", err)
	}
}

func TestLoginUnknownUserLooksLikeWrongPassword(t *testing.T) {
	checks := fakeBcrypt(t)
	repo := &fakeLoginRepo{}

	_, _, err := login(repo, "whatever")
	assertInvalidCredentials(t, err)

	if *checks != 1 {
		t.Fatalf("password comparisons = %d, want 1 (dummy check for timing)", *checks)
	}

	if repo.transactions != 0 {
		t.Fatal("unknown user must not write to the database")
	}
}

func TestLoginWrongPasswordCountsFailure(t *testing.T) {
	fakeBcrypt(t)
	repo := &fakeLoginRepo{user: newLoginUser(true, 0)}

	_, _, err := login(repo, "wrong")
	assertInvalidCredentials(t, err)

	if len(repo.loginFailed) != 1 || repo.loginFailed[0] != 1 {
		t.Fatalf("login_failed updates = %v, want [1]", repo.loginFailed)
	}

	if len(repo.statusUpdates) != 0 {
		t.Fatal("account must not be deactivated yet")
	}
}

func TestLoginLockoutUsesGenericResponse(t *testing.T) {
	fakeBcrypt(t)
	repo := &fakeLoginRepo{user: newLoginUser(true, 3)}

	_, _, err := login(repo, "wrong")
	assertInvalidCredentials(t, err)

	if len(repo.statusUpdates) != 1 || repo.statusUpdates[0] != false {
		t.Fatalf("status updates = %v, want [false]", repo.statusUpdates)
	}

	if repo.revokes != 1 {
		t.Fatalf("session revokes = %d, want 1", repo.revokes)
	}

	if len(repo.activities) != 1 || repo.activities[0] != "LOGIN_FAILED" {
		t.Fatalf("activities = %v, want [LOGIN_FAILED]", repo.activities)
	}
}

func TestLoginInactiveUserWrongPassword(t *testing.T) {
	fakeBcrypt(t)
	repo := &fakeLoginRepo{user: newLoginUser(false, 4)}

	_, _, err := login(repo, "wrong")
	assertInvalidCredentials(t, err)

	if repo.transactions != 0 {
		t.Fatal("failed attempts on an inactive account must not be recorded")
	}
}

func TestLoginInactiveUserCorrectPassword(t *testing.T) {
	fakeBcrypt(t)
	repo := &fakeLoginRepo{user: newLoginUser(false, 4)}

	_, _, err := login(repo, correctPassword)

	var apiErr response.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusForbidden {
		t.Fatalf("err = %v, want 403 User Not Active", err)
	}

	if repo.transactions != 0 || repo.sessions != 0 {
		t.Fatal("inactive user must not get a session")
	}
}

func TestLoginSuccess(t *testing.T) {
	fakeBcrypt(t)
	repo := &fakeLoginRepo{user: newLoginUser(true, 2)}

	access, refresh, err := login(repo, correctPassword)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	if access == nil || refresh == nil || *access == "" || *refresh == "" {
		t.Fatal("expected access and refresh tokens")
	}

	if len(repo.loginFailed) != 1 || repo.loginFailed[0] != 0 {
		t.Fatalf("login_failed updates = %v, want reset to [0]", repo.loginFailed)
	}

	if repo.sessions != 1 {
		t.Fatalf("sessions created = %d, want 1", repo.sessions)
	}
}
