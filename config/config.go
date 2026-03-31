package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddress      string
	PaymentBaseURL   string
	InventoryBaseURL string
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		HTTPAddress:      getenv("HTTP_ADDRESS", ":8080"),
		PaymentBaseURL:   getenv("PAYMENT_BASE_URL", "http://payment.local"),
		InventoryBaseURL: getenv("INVENTORY_BASE_URL", "http://inventory.local"),
	}

	if cfg.HTTPAddress == "" {
		return Config{}, fmt.Errorf("HTTP_ADDRESS is required")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
