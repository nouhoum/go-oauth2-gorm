package internal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	image  = "postgres:latest"
	logMsg = "database system is ready to accept connections"
)

type PGConfig struct {
	User     string
	Password string
	Database string
	Port     int
}

//EmbedPostgres spins up a postgres container.
func EmbedPostgres(t *testing.T, cfg PGConfig) (string, int) {
	t.Helper()

	ctx := context.Background()
	natPort := "5432/tcp"
	dbURL := func(port nat.Port) string {
		return fmt.Sprintf("postgres://test:test@localhost:%s/%s?sslmode=disable", port.Port(), cfg.Database)
	}
	// Setup and startup container
	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{natPort},
		Env: map[string]string{
			"POSTGRES_USER":     cfg.User,
			"POSTGRES_PASSWORD": cfg.Password,
			"POSTGRES_DATABASE": cfg.Database,
		},
		WaitingFor: wait.ForAll(
			wait.ForSQL(nat.Port(natPort), "postgres", dbURL),
			wait.ForLog(logMsg),
		).WithStartupTimeout(time.Minute * 10),
	}
	pg, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	// Even after log message found Postgres needs a touch more...
	time.Sleep(200 * time.Millisecond)
	// When test is done terminate container
	t.Cleanup(func() {
		_ = pg.Terminate(ctx)
	})
	// Get the container info needed
	containerPort, err := pg.MappedPort(ctx, nat.Port(natPort))
	if err != nil {
		t.Error(err)
	}
	containerHost, err := pg.Host(ctx)
	if err != nil {
		t.Error(err)
	}

	return containerHost, containerPort.Int()
}
