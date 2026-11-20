package camera

import (
	"context"
	"strings"
	"sync"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// dryRunProvider hands out a gorm DB that builds SQL without connecting and
// records every query it would run.
type dryRunProvider struct {
	db *gorm.DB

	mu      sync.Mutex
	queries []string
}

func newDryRunProvider(t *testing.T) *dryRunProvider {
	t.Helper()

	db, err := gorm.Open(
		postgres.New(postgres.Config{DSN: "host=127.0.0.1 dbname=dryrun"}),
		&gorm.Config{DryRun: true, DisableAutomaticPing: true, Logger: logger.Discard},
	)
	if err != nil {
		t.Fatalf("open dry-run db: %v", err)
	}

	p := &dryRunProvider{db: db}

	if err := db.Callback().Query().After("gorm:query").Register("test:capture", func(tx *gorm.DB) {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.queries = append(p.queries, tx.Statement.SQL.String())
	}); err != nil {
		t.Fatalf("register callback: %v", err)
	}

	return p
}

func (p *dryRunProvider) Db(ctx context.Context) *gorm.DB { return p.db.WithContext(ctx) }

func (p *dryRunProvider) WithTransaction(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	return fn(ctx, p.db)
}

func (p *dryRunProvider) cameraQuery(t *testing.T) string {
	t.Helper()

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, q := range p.queries {
		if strings.Contains(q, "camera_access") {
			return q
		}
	}

	t.Fatalf("no query joined camera_access; queries: %v", p.queries)
	return ""
}

func TestPerUserCameraQueriesOnlyUseActiveAccess(t *testing.T) {
	calls := map[string]func(r *repo) error{
		"FindCamerasByUser": func(r *repo) error {
			_, err := r.FindCamerasByUser(context.Background(), "user-1")
			return err
		},
		"FindCameraAccessUser": func(r *repo) error {
			_, err := r.FindCameraAccessUser(context.Background(), "camera-1", "user-1")
			if err == gorm.ErrRecordNotFound {
				return nil // dry run returns no rows
			}
			return err
		},
	}

	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			p := newDryRunProvider(t)

			if err := call(&repo{provider: p}); err != nil {
				t.Fatalf("%s: %v", name, err)
			}

			sql := p.cameraQuery(t)
			for _, want := range []string{"ca.active = TRUE", "ca.deleted_at IS NULL", "ca.user_id ="} {
				if !strings.Contains(sql, want) {
					t.Errorf("missing %q in SQL:\n%s", want, sql)
				}
			}
		})
	}
}

func TestActiveAccessJoinConditions(t *testing.T) {
	for _, want := range []string{"ca.camera_id = cameras.uuid", "ca.active = TRUE", "ca.deleted_at IS NULL"} {
		if !strings.Contains(activeAccessJoin, want) {
			t.Errorf("activeAccessJoin missing %q", want)
		}
	}
}
