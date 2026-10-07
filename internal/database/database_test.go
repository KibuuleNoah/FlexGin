package database_test

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"flexgin/internal/config"
	"flexgin/internal/database"
)

var testCfg config.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.Run(
		ctx, "postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("container host: %v", err)
	}
	mapped, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		log.Fatalf("container port: %v", err)
	}
	port, _ := strconv.Atoi(mapped.Port())

	testCfg = config.DB{
		Host: host, Port: port, Name: "testdb",
		User: "user", Password: "password",
		Schema: "public", SSLMode: "disable",
		MaxOpenConns: 5, MaxIdleConns: 5,
		ConnMaxLifetime: time.Minute, ConnMaxIdleTime: time.Minute,
	}

	code := m.Run()

	if err := container.Terminate(ctx); err != nil {
		log.Printf("terminate container: %v", err)
	}
	os.Exit(code)
}

func TestNew(t *testing.T) {
	db, err := database.New(testCfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()

	if err := db.Health(context.Background()); err != nil {
		t.Fatalf("Health: %v", err)
	}
}

func TestWithTx(t *testing.T) {
	db, err := database.New(testCfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS tx_probe (id INT PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}

	sentinel := errors.New("boom")
	err = db.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO tx_probe (id) VALUES (1)`); err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("want sentinel error, got %v", err)
	}

	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM tx_probe`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("rollback failed: %d rows", n)
	}
}

func TestIsUniqueViolation(t *testing.T) {
	db, err := database.New(testCfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS uniq_probe (v TEXT UNIQUE)`); err != nil {
		t.Fatal(err)
	}
	_, _ = db.ExecContext(ctx, `INSERT INTO uniq_probe (v) VALUES ('a')`)
	_, err = db.ExecContext(ctx, `INSERT INTO uniq_probe (v) VALUES ('a')`)
	if !database.IsUniqueViolation(err) {
		t.Fatalf("want unique violation, got %v", err)
	}
}
