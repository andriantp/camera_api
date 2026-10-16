package postgres_test

import (
	"context"
	"os"
	"testing"

	"camera_api/pkg/adapter/database/postgres"
)

/*
DATABASE_DSN="host=localhost port=5432 user= password= dbname=camera sslmode=" \
go test ./pkg/adapter/database/postgres -v
*/
func TestNewPostgres(t *testing.T) {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		t.Fatal("DATABASE_DSN environment variable is required")
	}

	db, err := postgres.NewPostgres(postgres.Setting{
		DSN:     dsn,
		Migrate: true,
	})
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}

	if db == nil {
		t.Fatal("database is nil")
	}

	ctx := context.Background()

	gormDB := db.Db(ctx)
	if gormDB == nil {
		t.Fatal("gorm database is nil")
	}
}



