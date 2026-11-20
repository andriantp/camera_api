package euser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"camera_api/app/usecase/uuser"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/token"

	"github.com/labstack/echo/v4"
)

// fakeUserUsecase records the paging GetUsers receives. Other methods panic
// on the nil embedded interface.
type fakeUserUsecase struct {
	uuser.Usecase

	limit, offset int
}

func (f *fakeUserUsecase) GetUsers(_ context.Context, _ request.UserActor, limit, offset int) (*response.UserList, error) {
	f.limit, f.offset = limit, offset
	return &response.UserList{}, nil
}

func TestGetUsersLimit(t *testing.T) {
	tests := []struct {
		query     string
		wantLimit int
	}{
		{query: "", wantLimit: 20},
		{query: "?limit=50", wantLimit: 50},
		{query: "?limit=100", wantLimit: 100},
		{query: "?limit=1000000", wantLimit: 100},
		{query: "?limit=-5", wantLimit: 20},
		{query: "?limit=abc", wantLimit: 20},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			e := echo.New()
			c := e.NewContext(httptest.NewRequest(http.MethodGet, "/user"+tt.query, nil), httptest.NewRecorder())
			c.Set("tokenClaims", &token.CustomClaims{User: "admin-1", Role: "admin"})

			ucase := &fakeUserUsecase{}
			if err := (&handler{ucaseUser: ucase}).GetUsers(c); err != nil {
				t.Fatalf("GetUsers: %v", err)
			}

			if ucase.limit != tt.wantLimit {
				t.Fatalf("limit = %d, want %d", ucase.limit, tt.wantLimit)
			}
		})
	}
}
