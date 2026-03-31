package app

import (
	"answer/task3/config"
	"answer/task3/handlers/http"
	"answer/task3/infrastructure"
	"answer/task3/services"
	"answer/task3/usecases/create_order"
)

type Application struct {
	OrderHandler *http.OrderHandler
	OrderService *services.OrderService
}

func New(cfg config.Config) *Application {
	orderRepo := infrastructure.NewOrderRepository()
	paymentClient := &infrastructure.PaymentClient{BaseURL: cfg.PaymentBaseURL}
	inventoryClient := &infrastructure.InventoryClient{BaseURL: cfg.InventoryBaseURL}

	createOrderUC := create_order.New(orderRepo, paymentClient, inventoryClient)
	orderService := services.NewOrderService(createOrderUC)

	return &Application{
		OrderHandler: http.NewOrderHandler(orderService),
		OrderService: orderService,
	}
}
