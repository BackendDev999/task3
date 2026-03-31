package contracts

import (
	"context"

	"answer/task3/domain"
)

type OrderRepository interface {
	Insert(ctx context.Context, order *domain.Order) error
}

type PaymentClient interface {
	Authorize(ctx context.Context, orderID string, amountCents int64) error
}

type InventoryClient interface {
	Reserve(ctx context.Context, orderID string) error
}

type Clock interface {
	Now() int64
}
