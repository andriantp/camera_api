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
	"camera_api/pkg/utils/token"

	"gorm.io/gorm"
)

// fakeRepo implements only the methods these tests need. Any other call
// panics on the nil embedded interface, which flags unexpected repo access.
type fakeRepo struct {
	user.Repository

	session    *model.ActiveSession
	sessionErr error

	findUserCalls int
}

func (f *fakeRepo) FindActiveSession(_ context.Context, _, _ string) (*model.ActiveSession, error) {
	return f.session, f.sessionErr
}

func (f *fakeRepo) FindUserByUserUUID(_ context.Context, _ string) (*model.UserToken, error) {
	f.findUserCalls++
	return nil, errors.New("FindUserByUserUUID must not be called")
}

var testJWT = token.Token{
	Secret: []byte("secret-test-camera-api"),
	Issuer: "camera",
}

func assertStatus(t *testing.T, err error, want int) {
	t.Helper()

	var apiErr response.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != want {
		t.Fatalf("err = %v, want APIError with status %d", err, want)
	}
}

func TestValidateSession(t *testing.T) {
	tests := []struct {
		name       string
		repo       *fakeRepo
		wantRole   string
		wantStatus int
	}{
		{
			name:     "active session",
			repo:     &fakeRepo{session: &model.ActiveSession{RoleName: "admin"}},
			wantRole: "admin",
		},
		{
			name:       "revoked or missing session",
			repo:       &fakeRepo{sessionErr: gorm.ErrRecordNotFound},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "database error",
			repo:       &fakeRepo{sessionErr: errors.New("connection refused")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUsecase(Expire{}, testJWT, tt.repo)

			role, err := u.ValidateSession(context.Background(), "session-1", "user-1")
			if tt.wantStatus != 0 {
				assertStatus(t, err, tt.wantStatus)
				return
			}

			if err != nil {
				t.Fatalf("ValidateSession: %v", err)
			}

			if role != tt.wantRole {
				t.Fatalf("role = %q, want %q", role, tt.wantRole)
			}
		})
	}
}

func TestRefreshRejectsRevokedSession(t *testing.T) {
	repo := &fakeRepo{sessionErr: gorm.ErrRecordNotFound}
	u := NewUsecase(Expire{Token: time.Minute, Refresh: time.Hour}, testJWT, repo)

	refreshToken, err := testJWT.GenerateRefresh(
		&token.Claims{UserUUID: "user-1"},
		time.Hour,
		"session-1",
	)
	if err != nil {
		t.Fatalf("GenerateRefresh: %v", err)
	}

	access, refresh, err := u.Refresh(context.Background(), refreshToken, request.UserActor{})
	assertStatus(t, err, http.StatusUnauthorized)

	if access != nil || refresh != nil {
		t.Fatal("no tokens must be issued for a revoked session")
	}

	if repo.findUserCalls != 0 {
		t.Fatal("Refresh must stop before loading the user")
	}
}
