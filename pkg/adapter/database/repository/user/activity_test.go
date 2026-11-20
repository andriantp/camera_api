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
