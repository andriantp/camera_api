package uuser

import (
	"context"
	"errors"
	"net/http"
	"strings"
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

	// alreadyLocked simulates a concurrent request that locked the user first.
	alreadyLocked bool

	transactions int
	resets       int
	increments   int
	locks        []time.Time
	revokes      int
	activities   []string
	sessions     int
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

func (f *fakeLoginRepo) ResetLoginFailures(context.Context, string) error {
	f.resets++
	return nil
}

func (f *fakeLoginRepo) IncrementLoginFailed(context.Context, string) (int, error) {
	f.increments++
	return f.user.LoginFailed + f.increments, nil
}

func (f *fakeLoginRepo) LockUser(_ context.Context, _ string, until, _ time.Time) (bool, error) {
	if f.alreadyLocked {
		return false, nil
	}

	f.locks = append(f.locks, until)
	return true, nil
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

	if repo.increments != 1 {
		t.Fatalf("login_failed increments = %d, want 1", repo.increments)
	}

	if len(repo.locks) != 0 {
		t.Fatal("account must not be locked yet")
	}
}

func TestLoginLockoutUsesGenericResponse(t *testing.T) {
	fakeBcrypt(t)
	repo := &fakeLoginRepo{user: newLoginUser(true, 3)}

	before := time.Now().UTC()
	_, _, err := login(repo, "wrong")
	assertInvalidCredentials(t, err)

	if len(repo.locks) != 1 {
		t.Fatalf("locks = %d, want 1", len(repo.locks))
	}

	if until := repo.locks[0]; until.Before(before.Add(loginLockDuration)) || until.After(time.Now().UTC().Add(loginLockDuration)) {
		t.Errorf("locked until %v, want about now + %v", until, loginLockDuration)
	}

	if repo.revokes != 0 {
		t.Error("a temporary lock must not revoke the owner's sessions")
	}

	if len(repo.activities) != 1 || repo.activities[0] != "LOGIN_LOCKED" {
		t.Fatalf("activities = %v, want [LOGIN_LOCKED]", repo.activities)
	}
}

func lockedUser(until time.Time) *model.UserSession {
	u := newLoginUser(true, 0)
	u.LockedUntil = &until
	return u
}

func TestLoginWhileLockedWrongPassword(t *testing.T) {
	fakeBcrypt(t)
	repo := &fakeLoginRepo{user: lockedUser(time.Now().UTC().Add(10 * time.Minute))}

	_, _, err := login(repo, "wrong")
	assertInvalidCredentials(t, err)

	if repo.transactions != 0 {
		t.Fatal("failed attempts while locked must not be counted")
	}
}

func TestLoginWhileLockedCorrectPassword(t *testing.T) {
	fakeBcrypt(t)
	repo := &fakeLoginRepo{user: lockedUser(time.Now().UTC().Add(10*time.Minute + 30*time.Second))}

	_, _, err := login(repo, correctPassword)

	var apiErr response.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("err = %v, want 429", err)
	}

	if !strings.Contains(apiErr.Message, "11 minutes") {
		t.Errorf("message = %q, want the remaining time rounded up (11 minutes)", apiErr.Message)
	}

	if repo.sessions != 0 {
		t.Fatal("locked user must not get a session")
	}
}

func TestLoginAfterLockExpired(t *testing.T) {
	fakeBcrypt(t)
	repo := &fakeLoginRepo{user: lockedUser(time.Now().UTC().Add(-time.Minute))}

	if _, _, err := login(repo, correctPassword); err != nil {
		t.Fatalf("Login after lock expired: %v", err)
	}

	if repo.resets != 1 || repo.sessions != 1 {
		t.Fatalf("resets=%d sessions=%d, want 1 and 1", repo.resets, repo.sessions)
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

	if repo.resets != 1 {
		t.Fatalf("login failure resets = %d, want 1", repo.resets)
	}

	if repo.sessions != 1 {
		t.Fatalf("sessions created = %d, want 1", repo.sessions)
	}
}

func TestLoginLockoutAlreadyDoneByConcurrentRequest(t *testing.T) {
	fakeBcrypt(t)
	repo := &fakeLoginRepo{user: newLoginUser(true, 3), alreadyLocked: true}

	_, _, err := login(repo, "wrong")
	assertInvalidCredentials(t, err)

	if len(repo.activities) != 0 {
		t.Fatalf("activities = %v, want none: another request already locked the account", repo.activities)
	}
}
