package postgres

import (
	"testing"
	"time"
)

func TestWithPoolDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   Setting
		want Setting
	}{
		{
			name: "all unset",
			want: Setting{MaxOpenConns: 25, MaxIdleConns: 10, ConnMaxLifetime: 30 * time.Minute, ConnMaxIdleTime: 5 * time.Minute},
		},
		{
			name: "custom values kept",
			in:   Setting{MaxOpenConns: 50, MaxIdleConns: 20, ConnMaxLifetime: time.Hour, ConnMaxIdleTime: time.Minute},
			want: Setting{MaxOpenConns: 50, MaxIdleConns: 20, ConnMaxLifetime: time.Hour, ConnMaxIdleTime: time.Minute},
		},
		{
			name: "idle capped at open",
			in:   Setting{MaxOpenConns: 5, MaxIdleConns: 10},
			want: Setting{MaxOpenConns: 5, MaxIdleConns: 5, ConnMaxLifetime: 30 * time.Minute, ConnMaxIdleTime: 5 * time.Minute},
		},
		{
			name: "default idle capped at small open",
			in:   Setting{MaxOpenConns: 4},
			want: Setting{MaxOpenConns: 4, MaxIdleConns: 4, ConnMaxLifetime: 30 * time.Minute, ConnMaxIdleTime: 5 * time.Minute},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.withPoolDefaults(); got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
