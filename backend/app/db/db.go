package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/tracewayapp/go-lightning/lit"
	_ "modernc.org/sqlite"
)

var DB *sql.DB
var Driver lit.Driver = lit.PostgreSQL

func IsSQLite() bool {
	return Driver == lit.SQLite
}

func Init() error {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "sqlite" {
		return initSQLite()
	}
	return initPostgres()
}

func initPostgres() error {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	database := os.Getenv("POSTGRES_DATABASE")
	username := os.Getenv("POSTGRES_USERNAME")
	password := os.Getenv("POSTGRES_PASSWORD")
	sslMode := os.Getenv("POSTGRES_SSLMODE")

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
	path := os.Getenv("SQLITE_PATH")
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
