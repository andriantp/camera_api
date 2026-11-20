package postgres_test

import (
	"context"
	"os"
	"testing"

	"camera_api/pkg/adapter/database/postgres"
)

/*
Integration test: needs a running PostgreSQL. Skipped when DATABASE_DSN is empty.

DATABASE_DSN="host=localhost port=5432 user= password= dbname=camera sslmode=disable" \
go test ./pkg/adapter/database/postgres -v

The test never runs migrations, so it is safe against any database.
*/
func TestNewPostgres(t *testing.T) {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		t.Skip("DATABASE_DSN not set; skipping PostgreSQL integration test")
	}

	db, err := postgres.NewPostgres(postgres.Setting{
		DSN:     dsn,
		Migrate: false,
	})
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}

	if db == nil {
		t.Fatal("database is nil")
	}

	var one int
	if err := db.Db(context.Background()).Raw("SELECT 1").Scan(&one).Error; err != nil {
		t.Fatalf("SELECT 1: %v", err)
	}

	if one != 1 {
		t.Fatalf("SELECT 1 returned %d", one)
	}
}
