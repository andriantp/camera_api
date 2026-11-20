package user

import (
	"strings"
	"testing"
	"time"

	"camera_api/pkg/adapter/database/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newDryRunDB returns a gorm DB that builds SQL without connecting.
func newDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(
		postgres.New(postgres.Config{DSN: "host=127.0.0.1 dbname=dryrun"}),
		&gorm.Config{
			DryRun:               true,
			DisableAutomaticPing: true,
			Logger:               logger.Discard,
		},
	)
	if err != nil {
		t.Fatalf("open dry-run db: %v", err)
	}

	return db
}

// activitySQL returns the SQL that FindUserActivities runs for a filter.
func activitySQL(t *testing.T, filter ActivityFilter) string {
	t.Helper()

	db := newDryRunDB(t)

	return db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var out []model.UserActivitySummary
		return activityQuery(tx, filter).Find(&out)
	})
}

func TestFindUserActivitiesOptionalTimeRange(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name     string
		from, to time.Time
		wantFrom bool
		wantTo   bool
	}{
		{name: "no bounds"},
		{name: "only from", from: now, wantFrom: true},
		{name: "only to", to: now, wantTo: true},
		{name: "both bounds", from: now.Add(-time.Hour), to: now, wantFrom: true, wantTo: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := activitySQL(t, ActivityFilter{
				ActorUserID: "user-1",
				Target:      ActivityTargetAll,
				From:        tt.from,
				To:          tt.to,
				Limit:       20,
			})

			if got := strings.Contains(sql, "user_activities.created_at >="); got != tt.wantFrom {
				t.Errorf("from condition present = %v, want %v\nSQL: %s", got, tt.wantFrom, sql)
			}

			if got := strings.Contains(sql, "user_activities.created_at <="); got != tt.wantTo {
				t.Errorf("to condition present = %v, want %v\nSQL: %s", got, tt.wantTo, sql)
			}

			if !strings.Contains(sql, "user_activities.actor_user_id =") {
				t.Errorf("actor condition missing\nSQL: %s", sql)
			}
		})
	}
}

func TestUserQueriesUseSameLastSeen(t *testing.T) {
	db := newDryRunDB(t)

	queries := map[string]string{
		"list (FindUsers)": db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			var out []model.UserSummary
			return usersQuery(tx, 2, 20, 0).Find(&out)
		}),
		"profile (FindUserSummaryByUserUUID)": db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			var out model.UserSummary
			return userSummaryQuery(tx, "user-1").Find(&out)
		}),
	}

	for name, sql := range queries {
		t.Run(name, func(t *testing.T) {
			checkLastSeenSQL(t, sql)
		})
	}
}

func checkLastSeenSQL(t *testing.T, sql string) {
	t.Helper()

	for _, want := range []string{
		"LEFT JOIN LATERAL",
		"FROM sessions s",
		"s.last_seen_at IS NOT NULL",
		"ORDER BY s.last_seen_at DESC",
		"last_session.last_seen_at AS last_seen",
	} {
		if !strings.Contains(sql, want) {
			t.Errorf("missing %q in SQL:\n%s", want, sql)
		}
	}

	// Joining every session and grouping is what the list query replaced;
	// NULLS LAST can't use idx_sessions_user_last_seen; filtering on
	// logout_at made the profile disagree with the list.
	for _, banned := range []string{"GROUP BY", "LEFT JOIN sessions ON", "NULLS LAST", "logout_at"} {
		if strings.Contains(sql, banned) {
			t.Errorf("unexpected %q in SQL:\n%s", banned, sql)
		}
	}
}
