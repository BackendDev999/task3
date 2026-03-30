package integration

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"

	"answer/task3/testing/setup"
)

func TestOrderRepo_Integration(t *testing.T) {
	pg := setup.StartPostgres(t)

	db, err := sql.Open("postgres", pg.DSN)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()

	_, err = db.ExecContext(ctx, `
		CREATE TABLE orders (
			id TEXT PRIMARY KEY,
			customer_id TEXT NOT NULL,
			status TEXT NOT NULL,
			total_cents BIGINT NOT NULL,
			shipping_address TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}

	// Example assertions the real integration test must verify:
	// 1. CreateMut produces executable SQL against PostgreSQL.
	// 2. UpdateMut only changes dirty columns by comparing persisted values before/after update.
	// 3. Database constraints reject invalid rows, proving mocks would not catch the issue.
	//
	// Pseudocode:
	// repo := repo.NewOrderRepo(db)
	// mut, err := repo.CreateMut(order)
	// require.NoError(t, err)
	// require.NoError(t, repo.ExecMut(ctx, mut))
	//
	// loaded, err := repo.GetByID(ctx, order.ID)
	// require.NoError(t, err)
	// require.Equal(t, expected, loaded)
	//
	// order.UpdateShippingAddress("new address")
	// mut, err = repo.UpdateMut(order)
	// require.NoError(t, err)
	// require.NoError(t, repo.ExecMut(ctx, mut))
	//
	// var customerID, shippingAddress string
	// err = db.QueryRowContext(ctx, "SELECT customer_id, shipping_address FROM orders WHERE id = $1", order.ID).
	//     Scan(&customerID, &shippingAddress)
	// require.NoError(t, err)
	// require.Equal(t, "original-customer", customerID) // unchanged
	// require.Equal(t, "new address", shippingAddress)
}
