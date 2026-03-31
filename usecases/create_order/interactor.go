package create_order

import (
	"context"
	"fmt"
	"time"

	"answer/task3/contracts"
	"answer/task3/domain"
)

type Request struct {
	OrderID      string `json:"order_id"`
	CustomerID   string `json:"customer_id"`
	AmountCents  int64  `json:"amount_cents"`
	CustomerTier string `json:"customer_tier"`
}

type Response struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

type Interactor struct {
	orders    contracts.OrderRepository
	payments  contracts.PaymentClient
	inventory contracts.InventoryClient
}

func New(
	orders contracts.OrderRepository,
	payments contracts.PaymentClient,
	inventory contracts.InventoryClient,
) *Interactor {
	return &Interactor{
		orders:    orders,
		payments:  payments,
		inventory: inventory,
	}
}

func (uc *Interactor) Execute(ctx context.Context, req Request) (Response, error) {
	order, err := domain.NewOrder(req.OrderID, req.CustomerID, req.AmountCents, time.Now().UTC())
	if err != nil {
		return Response{}, fmt.Errorf("build order: %w", err)
	}

	if err := uc.payments.Authorize(ctx, order.ID, order.AmountCents); err != nil {
		return Response{}, fmt.Errorf("authorize payment: %w", err)
	}

	if err := uc.inventory.Reserve(ctx, order.ID); err != nil {
		return Response{}, fmt.Errorf("reserve inventory: %w", err)
	}

	order.MarkAuthorized()

	if err := uc.orders.Insert(ctx, order); err != nil {
		return Response{}, fmt.Errorf("persist order: %w", err)
	}

	return Response{
		OrderID: order.ID,
		Status:  order.Status,
	}, nil
}
