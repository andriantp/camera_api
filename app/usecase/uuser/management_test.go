package uuser

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/adapter/database/repository/user"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// fakeManagementRepo implements only what the tests use; other methods panic
// on the nil embedded interface.
type fakeManagementRepo struct {
	user.Repository

	role    *model.Role
	roleErr error

	users []model.UserSummary
}

func (f *fakeManagementRepo) FindRoleByName(context.Context, string) (*model.Role, error) {
	return f.role, f.roleErr
}

func (f *fakeManagementRepo) FindUsers(context.Context, int, int, int) ([]model.UserSummary, error) {
	return f.users, nil
}

func (f *fakeManagementRepo) FindRoles(context.Context, int) ([]model.Role, error) {
	return []model.Role{{ID: 3, UUID: uuid.New(), Name: "user"}}, nil
}

func (f *fakeManagementRepo) FindUserSummaryByUserUUID(context.Context, string) (*model.UserSummary, error) {
	return &model.UserSummary{UUID: "user-2", Username: "bob", RoleID: 3, Role: "user"}, nil
}

var adminActor = request.UserActor{UserID: "user-1", Role: "admin"}

func TestGetUsersEmptyPage(t *testing.T) {
	repo := &fakeManagementRepo{role: &model.Role{ID: 2, Name: "admin"}}
	u := NewUsecase(Expire{}, testJWT, repo)

	result, err := u.GetUsers(context.Background(), adminActor, 20, 1000)
	if err != nil {
		t.Fatalf("GetUsers: %v", err)
	}

	if result.Total != 0 || result.Returned != 0 || len(result.Users) != 0 {
		t.Fatalf("result = %+v, want empty list", result)
	}
}

func TestGetUsersReturnsTotal(t *testing.T) {
	repo := &fakeManagementRepo{
		role: &model.Role{ID: 2, Name: "admin"},
		users: []model.UserSummary{
			{Total: 5, UUID: "a", RoleID: 2},
			{Total: 5, UUID: "b", RoleID: 3},
		},
	}
	u := NewUsecase(Expire{}, testJWT, repo)

	result, err := u.GetUsers(context.Background(), adminActor, 2, 0)
	if err != nil {
		t.Fatalf("GetUsers: %v", err)
	}

	if result.Total != 5 || result.Returned != 2 {
		t.Fatalf("total=%d returned=%d, want 5 and 2", result.Total, result.Returned)
	}

	if result.Users[0].CanUpdate || !result.Users[1].CanUpdate {
		t.Fatal("admin may update only users with a lower role")
	}
}

func TestActorRoleLookupErrors(t *testing.T) {
	calls := map[string]func(Usecase) error{
		"GetUsers": func(u Usecase) error {
			_, err := u.GetUsers(context.Background(), adminActor, 20, 0)
			return err
		},
		"GetRoles": func(u Usecase) error {
			_, err := u.GetRoles(context.Background(), adminActor)
			return err
		},
		"GetUserProfile": func(u Usecase) error {
			_, err := u.GetUserProfile(context.Background(), adminActor, "user-2")
			return err
		},
	}

	cases := []struct {
		name       string
		roleErr    error
		wantStatus int
	}{
		{name: "database error", roleErr: errors.New("connection refused"), wantStatus: http.StatusInternalServerError},
		{name: "role not found", roleErr: gormNotFound, wantStatus: http.StatusForbidden},
	}

	for fn, call := range calls {
		for _, tc := range cases {
			t.Run(fn+"/"+tc.name, func(t *testing.T) {
				u := NewUsecase(Expire{}, testJWT, &fakeManagementRepo{roleErr: tc.roleErr})

				var apiErr response.APIError
				if err := call(u); !errors.As(err, &apiErr) || apiErr.StatusCode != tc.wantStatus {
					t.Fatalf("err = %v, want APIError with status %d", err, tc.wantStatus)
				}
			})
		}
	}
}

var gormNotFound = gorm.ErrRecordNotFound
