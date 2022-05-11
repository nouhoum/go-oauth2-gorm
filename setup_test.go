package gooauth2gorm_test

import (
	"context"
	"fmt"
	"log"
	"os"
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

var dbPort int
var dbHost string
var cfg = PGConfig{User: "test", Database: "test", Password: "test", Port: 5432}

func TestMain(m *testing.M) {
	log.Println("==== TEST MAIN ====")
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
		fmt.Fprintf(os.Stderr, "testcontainers.GenericContainer %v\n", err)
		os.Exit(1)
	}
	// Even after log message found Postgres needs a touch more...
	time.Sleep(200 * time.Millisecond)

	containerPort, err := pg.MappedPort(ctx, nat.Port(natPort))
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get the port %v\n", err)
		os.Exit(1)
	}
	containerHost, err := pg.Host(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get the host %v\n", err)
		os.Exit(1)
	}

	dbHost = containerHost
	dbPort = containerPort.Int()

	code := m.Run()

	_ = pg.Terminate(ctx)

	log.Println("==== TEST MAIN END ====")
	os.Exit(code)
}
