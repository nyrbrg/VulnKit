package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"vulnkit/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://vulnkit:vulnkit@localhost:5432/vulnkit_test?sslmode=disable"
	}

	ctx := context.Background()
	database, err := db.New(ctx, dsn)
	if err != nil {
		t.Skipf("Skipping integration test: no test DB available (%v)", err)
	}

	if err := database.Migrate(ctx); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	_, err = database.Pool.Exec(ctx, `TRUNCATE TABLE presets`)
	if err != nil {
		t.Fatalf("truncate failed: %v", err)
	}

	t.Cleanup(func() {
		database.Pool.Exec(context.Background(), `TRUNCATE TABLE presets`)
		database.Close()
	})

	fmt.Printf("[tests] connected to test DB: %s\n", dsn)
	return database.Pool
}
