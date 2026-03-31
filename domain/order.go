package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidOrderID    = errors.New("order id is required")
	ErrInvalidCustomerID = errors.New("customer id is required")
	ErrInvalidAmount     = errors.New("amount must be positive")
)

type Order struct {
	ID           string
	CustomerID   string
	AmountCents  int64
	Status       string
	CreatedAtUTC time.Time
}

func NewOrder(id, customerID string, amountCents int64, createdAt time.Time) (*Order, error) {
	order := &Order{
		ID:           id,
		CustomerID:   customerID,
		AmountCents:  amountCents,
		Status:       "PENDING",
		CreatedAtUTC: createdAt.UTC(),
	}

	if err := order.Validate(); err != nil {
		return nil, err
	}

	return order, nil
}

func (o *Order) Validate() error {
	if o.ID == "" {
		return ErrInvalidOrderID
	}
	if o.CustomerID == "" {
		return ErrInvalidCustomerID
	}
	if o.AmountCents <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

func (o *Order) MarkAuthorized() {
	o.Status = "AUTHORIZED"
}
