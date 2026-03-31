package infrastructure

import (
	"context"
	"fmt"
	"sync"

	"answer/task3/domain"
)

type OrderRepository struct {
	mu     sync.Mutex
	orders map[string]domain.Order
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		orders: make(map[string]domain.Order),
	}
}

func (r *OrderRepository) Insert(ctx context.Context, order *domain.Order) error {
	_ = ctx
	if err := order.Validate(); err != nil {
		return fmt.Errorf("validate order before insert: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.ID] = *order
	return nil
}
