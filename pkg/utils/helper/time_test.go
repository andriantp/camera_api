package helper

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"camera_api/pkg/api/response"
)

func TestConvertTimeRange(t *testing.T) {
	wib := time.FixedZone("WIB", 7*3600)
	wita := time.FixedZone("WITA", 8*3600)
	odd := time.FixedZone("ODD", 5*3600)

	from := time.Date(2026, 8, 18, 0, 0, 0, 0, wib)
	to := time.Date(2026, 8, 18, 23, 59, 59, 0, wib)

	tests := []struct {
		name       string
		from, to   time.Time
		wantFrom   time.Time
		wantTo     time.Time
		wantLoc    string
		wantStatus int
	}{
		{
			name:    "no bounds",
			wantLoc: "UTC",
		},
		{
			name:     "only from",
			from:     from,
			wantFrom: from.UTC(),
			wantLoc:  "Asia/Jakarta",
		},
		{
			name:    "only to",
			to:      time.Date(2026, 8, 18, 12, 0, 0, 0, wita),
			wantTo:  time.Date(2026, 8, 18, 4, 0, 0, 0, time.UTC),
			wantLoc: "Asia/Makassar",
		},
		{
			name:     "both bounds",
			from:     from,
			to:       to,
			wantFrom: from.UTC(),
			wantTo:   to.UTC(),
			wantLoc:  "Asia/Jakarta",
		},
		{
			name:       "different timezones",
			from:       from,
			to:         time.Date(2026, 8, 19, 0, 0, 0, 0, wita),
			wantStatus: http.StatusConflict,
		},
		{
			name:       "from after to",
			from:       to,
			to:         from,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unsupported timezone",
			from:       time.Date(2026, 8, 18, 0, 0, 0, 0, odd),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFrom, gotTo, loc, err := ConvertTimeRange(tt.from, tt.to)

			if tt.wantStatus != 0 {
				var apiErr response.APIError
				if !errors.As(err, &apiErr) || apiErr.StatusCode != tt.wantStatus {
					t.Fatalf("err = %v, want APIError with status %d", err, tt.wantStatus)
				}

				return
			}

			if err != nil {
				t.Fatalf("ConvertTimeRange: %v", err)
			}

			if !gotFrom.Equal(tt.wantFrom) || gotFrom.IsZero() != tt.wantFrom.IsZero() {
				t.Errorf("from = %v, want %v", gotFrom, tt.wantFrom)
			}

			if !gotTo.Equal(tt.wantTo) || gotTo.IsZero() != tt.wantTo.IsZero() {
				t.Errorf("to = %v, want %v", gotTo, tt.wantTo)
			}

			if !gotFrom.IsZero() && gotFrom.Location() != time.UTC {
				t.Errorf("from location = %v, want UTC", gotFrom.Location())
			}

			if loc.String() != tt.wantLoc {
				t.Errorf("loc = %v, want %v", loc, tt.wantLoc)
			}
		})
	}
}
