package main

import (
	"log"
	stdhttp "net/http"

	"answer/task3/app"
	"answer/task3/config"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	application := app.New(cfg)

	mux := stdhttp.NewServeMux()
	mux.HandleFunc("GET /health", application.OrderHandler.Health)
	mux.HandleFunc("POST /orders", application.OrderHandler.CreateOrder)

	log.Printf("task3 service listening on %s", cfg.HTTPAddress)
	if err := stdhttp.ListenAndServe(cfg.HTTPAddress, mux); err != nil {
		log.Fatalf("serve http: %v", err)
	}
}
