package http

import (
	"encoding/json"
	stdhttp "net/http"

	"answer/task3/services"
	"answer/task3/usecases/create_order"
)

type OrderHandler struct {
	service *services.OrderService
}

func NewOrderHandler(service *services.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) CreateOrder(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req create_order.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		stdhttp.Error(w, "invalid request", stdhttp.StatusBadRequest)
		return
	}

	res, err := h.service.CreateOrder(r.Context(), req)
	if err != nil {
		stdhttp.Error(w, err.Error(), stdhttp.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stdhttp.StatusCreated)
	_ = json.NewEncoder(w).Encode(res)
}

func (h *OrderHandler) Health(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stdhttp.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
