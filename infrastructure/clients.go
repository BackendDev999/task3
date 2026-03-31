package infrastructure

import (
	"context"
	"fmt"
)

type PaymentClient struct {
	BaseURL string
}

func (c *PaymentClient) Authorize(ctx context.Context, orderID string, amountCents int64) error {
	_ = ctx
	_ = orderID
	_ = amountCents
	if c.BaseURL == "" {
		return fmt.Errorf("payment base url is not configured")
	}
	return nil
}

type InventoryClient struct {
	BaseURL string
}

func (c *InventoryClient) Reserve(ctx context.Context, orderID string) error {
	_ = ctx
	_ = orderID
	if c.BaseURL == "" {
		return fmt.Errorf("inventory base url is not configured")
	}
	return nil
}
