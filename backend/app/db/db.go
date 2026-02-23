package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/tracewayapp/traceway/backend/app/config"

	_ "github.com/lib/pq"
	"github.com/tracewayapp/lit/v2"
	_ "modernc.org/sqlite"
)

var DB *sql.DB
var Driver lit.Driver = lit.PostgreSQL

func IsSQLite() bool {
	return Driver == lit.SQLite
}

func Init() error {
	cfg := config.Config
	if cfg.DBType == "sqlite" {
		return initSQLite()
	}
	return initPostgres()
}

func initPostgres() error {
	cfg := config.Config

	host := cfg.PostgresHost
	port := cfg.PostgresPort
	database := cfg.PostgresDatabase
	username := cfg.PostgresUsername
	password := cfg.PostgresPassword
	sslMode := cfg.PostgresSSLMode

	if sslMode == "" {
		sslMode = "disable"
	}
	if port == "" {
		port = "5432"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, username, password, database, sslMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open postgres connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	DB = db
	Driver = lit.PostgreSQL

	return nil
}

func initSQLite() error {
	path := config.Config.SQLitePath
	if path == "" {
		path = "./traceway.db"
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("failed to open sqlite connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping sqlite: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return fmt.Errorf("failed to set WAL mode: %w", err)
	}

	db.SetMaxOpenConns(1)

	DB = db
	Driver = lit.SQLite

	log.Printf("SQLite database opened at %s", path)

	return nil
}

func GetDB() *sql.DB {
	return DB
}

func ExecuteTransaction[T any](f func(tx *sql.Tx) (T, error)) (T, error) {
	tx, err := DB.Begin()

	if err != nil {
		var zero T
		return zero, err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	result, err := f(tx)

	if err != nil {
		tx.Rollback()
		var zero T
		return zero, err
	}

	if err := tx.Commit(); err != nil {
		var zero T
		return zero, err
	}

	return result, nil
}
