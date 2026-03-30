package setup

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	Container *postgres.PostgresContainer
	DSN       string
}

func StartPostgres(t *testing.T) *PostgresContainer {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("orders_test"),
		postgres.WithUsername("tester"),
		postgres.WithPassword("tester"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp"),
			wait.ForLog("database system is ready to accept connections"),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("build postgres dsn: %v", err)
	}

	return &PostgresContainer{
		Container: container,
		DSN:       dsn,
	}
}

type WireMockContainer struct {
	Container testcontainers.Container
	BaseURL   string
}

func StartWireMock(t *testing.T) *WireMockContainer {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)

	req := testcontainers.ContainerRequest{
		Image:        "wiremock/wiremock:3.9.1",
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForHTTP("/__admin").WithPort("8080/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start wiremock container: %v", err)
	}

	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("wiremock host: %v", err)
	}

	port, err := container.MappedPort(ctx, "8080/tcp")
	if err != nil {
		t.Fatalf("wiremock port: %v", err)
	}

	return &WireMockContainer{
		Container: container,
		BaseURL:   fmt.Sprintf("http://%s:%s", host, port.Port()),
	}
}
