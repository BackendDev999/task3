package services

import (
	"context"

	"answer/task3/usecases/create_order"
)

type OrderService struct {
	createOrder *create_order.Interactor
}

func NewOrderService(createOrder *create_order.Interactor) *OrderService {
	return &OrderService{createOrder: createOrder}
}

func (s *OrderService) CreateOrder(ctx context.Context, req create_order.Request) (create_order.Response, error) {
	return s.createOrder.Execute(ctx, req)
}
