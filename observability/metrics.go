package observability

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	OrdersCreatedTotal *prometheus.CounterVec
	OrderDuration      *prometheus.HistogramVec
	OrdersPending      prometheus.Gauge
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		OrdersCreatedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "orders_created_total",
				Help: "Number of order creation attempts by result and customer tier.",
			},
			[]string{"status", "customer_tier"},
		),
		OrderDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "order_processing_duration_seconds",
				Help:    "Duration of order processing steps.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"step", "result"},
		),
		OrdersPending: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "orders_pending_count",
				Help: "Current number of pending orders.",
			},
		),
	}

	reg.MustRegister(m.OrdersCreatedTotal, m.OrderDuration, m.OrdersPending)
	return m
}

func (m *Metrics) ObserveUsecase(ctx context.Context, step string, startedAt time.Time, err error) {
	result := "success"
	if err != nil {
		result = "failure"
	}

	_ = ctx
	m.OrderDuration.WithLabelValues(step, result).Observe(time.Since(startedAt).Seconds())
}

func (m *Metrics) IncOrdersCreated(customerTier string, err error) {
	status := "success"
	if err != nil {
		status = "failure"
	}

	m.OrdersCreatedTotal.WithLabelValues(status, customerTier).Inc()
}
