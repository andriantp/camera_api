package ucamera

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	adaptercamera "camera_api/pkg/adapter/camera"
	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"

	cameraModel "github.com/andriantp/camera/model"
	"github.com/google/uuid"
)

// fakeCameraAdapter records the search range sent to the camera. Other
// methods panic on the nil embedded interface.
type fakeCameraAdapter struct {
	adaptercamera.Camera

	searches []request.SearchContent
}

func (f *fakeCameraAdapter) GetContents(_ context.Context, _ adaptercamera.CameraTarget, req request.SearchContent) (*adaptercamera.Playlist, error) {
	f.searches = append(f.searches, req)

	uri := "rtsp://camera/playback"

	return &adaptercamera.Playlist{
		Protocol: model.ProtocolISAPI,
		Content: cameraModel.SearchResult{
			TotalMatches: 1,
			NumOfMatches: 1,
			Matches: []cameraModel.Record{{
				Filename: "20260818T010000Z_20260818T011000Z.mp4",
				Timeframe: cameraModel.Timeframe{
					Start: time.Date(2026, 8, 18, 1, 0, 0, 0, time.UTC),
					End:   time.Date(2026, 8, 18, 1, 10, 0, 0, time.UTC),
				},
				PlaybackURI: &uri,
			}},
		},
	}, nil
}

func newSearchUsecase() (*usecase, *fakeCameraAdapter) {
	adapter := &fakeCameraAdapter{}

	return &usecase{
		repoCamera: &fakeCameraRepo{camera: &model.Camera{
			UUID:     uuid.New(),
			Protocol: model.Protocol{Name: model.ProtocolISAPI},
		}},
		camera: adapter,
	}, adapter
}

func search(u *usecase, from, to time.Time) (*adaptercamera.Playlist, error) {
	return u.GetContents(
		context.Background(),
		request.UserActor{Role: "super_admin"},
		uuid.NewString(),
		request.SearchContent{From: from, To: to, Limit: 10},
	)
}

func TestGetContentsDefaultRange(t *testing.T) {
	wib := time.FixedZone("WIB", 7*3600)
	from := time.Date(2026, 8, 18, 0, 0, 0, 0, wib)
	to := time.Date(2026, 8, 18, 23, 59, 59, 0, wib)

	tests := []struct {
		name     string
		from, to time.Time
		// check validates the range sent to the camera.
		check func(t *testing.T, got request.SearchContent, before, after time.Time)
	}{
		{
			name: "no bounds: last 24 hours until now",
			check: func(t *testing.T, got request.SearchContent, before, after time.Time) {
				if got.To.Before(before) || got.To.After(after) {
					t.Errorf("to = %v, want now", got.To)
				}
				if !got.From.Equal(got.To.Add(-24 * time.Hour)) {
					t.Errorf("from = %v, want to - 24h", got.From)
				}
			},
		},
		{
			name: "only from: until now",
			from: from,
			check: func(t *testing.T, got request.SearchContent, before, after time.Time) {
				if !got.From.Equal(from) {
					t.Errorf("from = %v, want %v", got.From, from)
				}
				if got.To.Before(before) || got.To.After(after) {
					t.Errorf("to = %v, want now", got.To)
				}
			},
		},
		{
			name: "only to: 24 hours before to",
			to:   to,
			check: func(t *testing.T, got request.SearchContent, _, _ time.Time) {
				if !got.To.Equal(to) {
					t.Errorf("to = %v, want %v", got.To, to)
				}
				if !got.From.Equal(to.Add(-24 * time.Hour)) {
					t.Errorf("from = %v, want to - 24h", got.From)
				}
			},
		},
		{
			name: "both bounds: unchanged",
			from: from,
			to:   to,
			check: func(t *testing.T, got request.SearchContent, _, _ time.Time) {
				if !got.From.Equal(from) || !got.To.Equal(to) {
					t.Errorf("range = %v..%v, want %v..%v", got.From, got.To, from, to)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, adapter := newSearchUsecase()

			before := time.Now().UTC()
			if _, err := search(u, tt.from, tt.to); err != nil {
				t.Fatalf("GetContents: %v", err)
			}
			after := time.Now().UTC()

			if len(adapter.searches) != 1 {
				t.Fatalf("camera searches = %d, want 1", len(adapter.searches))
			}

			got := adapter.searches[0]
			if got.From.IsZero() || got.To.IsZero() {
				t.Fatalf("camera must always get both bounds, got %v..%v", got.From, got.To)
			}

			if got.From.Location() != time.UTC || got.To.Location() != time.UTC {
				t.Errorf("range must be sent in UTC, got %v..%v", got.From, got.To)
			}

			tt.check(t, got, before, after)
		})
	}
}

func TestGetContentsConvertsResultsToRequestTimezone(t *testing.T) {
	u, _ := newSearchUsecase()
	wib := time.FixedZone("WIB", 7*3600)

	playlist, err := search(u, time.Date(2026, 8, 18, 0, 0, 0, 0, wib), time.Time{})
	if err != nil {
		t.Fatalf("GetContents: %v", err)
	}

	match := playlist.Content.Matches[0]
	if match.Timeframe.Start.Location().String() != "Asia/Jakarta" {
		t.Errorf("start location = %v, want Asia/Jakarta", match.Timeframe.Start.Location())
	}

	if match.PlaybackURI != nil {
		t.Error("playback URI must not be exposed")
	}
}

func TestGetContentsInvalidRange(t *testing.T) {
	wib := time.FixedZone("WIB", 7*3600)
	wita := time.FixedZone("WITA", 8*3600)
	future := time.Now().Add(48 * time.Hour).In(wib)

	tests := []struct {
		name       string
		from, to   time.Time
		wantStatus int
	}{
		{name: "from after to", from: future, to: future.Add(-time.Hour), wantStatus: http.StatusBadRequest},
		{name: "only future from", from: future, wantStatus: http.StatusBadRequest},
		{name: "different timezones", from: future.Add(-72 * time.Hour), to: future.In(wita), wantStatus: http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, adapter := newSearchUsecase()

			_, err := search(u, tt.from, tt.to)

			var apiErr response.APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != tt.wantStatus {
				t.Fatalf("err = %v, want APIError with status %d", err, tt.wantStatus)
			}

			if len(adapter.searches) != 0 {
				t.Fatal("invalid range must not reach the camera")
			}
		})
	}
}
