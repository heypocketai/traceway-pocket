package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/tracewayapp/traceway/backend/app/chdb"
	"github.com/tracewayapp/traceway/backend/app/config"
	"github.com/tracewayapp/traceway/backend/app/db"

	"github.com/golang-migrate/migrate/v4"
	migrateCh "github.com/golang-migrate/migrate/v4/database/clickhouse"
	migratePg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

type ExtensionMigration struct {
	Source embed.FS
	Path   string
	Table  string // unique migration table name per extension
}

var ExtensionPostgresMigrations []ExtensionMigration

//go:embed ch/*.sql
var migrationsChFS embed.FS

//go:embed pg/*.sql
var migrationsPgFS embed.FS

//go:embed sqlite/*.sql
var migrationsSqliteFS embed.FS

func runMigrationsClickhouse(connStr string) error {
	db, err := sql.Open("clickhouse", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	source, err := iofs.New(migrationsChFS, "ch")
	if err != nil {
		return err
	}

	driver, err := migrateCh.WithInstance(db, &migrateCh.Config{
		MigrationsTableEngine: "MergeTree",
	})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", source, "clickhouse", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func runMigrationsPostgres(connStr string) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open postgres for migrations: %w", err)
	}
	defer db.Close()

	source, err := iofs.New(migrationsPgFS, "pg")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	driver, err := migratePg.WithInstance(db, &migratePg.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("postgres migration failed: %w", err)
	}

	for _, ext := range ExtensionPostgresMigrations {
		if err := runExtensionMigrations(db, ext); err != nil {
			return fmt.Errorf("extension migration failed: %w", err)
		}
	}

	return nil
}

func runExtensionMigrations(db *sql.DB, ext ExtensionMigration) error {
	source, err := iofs.New(ext.Source, ext.Path)
	if err != nil {
		return fmt.Errorf("failed to create extension migration source: %w", err)
	}

	tableName := ext.Table
	if tableName == "" {
		tableName = "schema_migrations_ext"
	}

	driver, err := migratePg.WithInstance(db, &migratePg.Config{
		MigrationsTable: tableName,
	})
	if err != nil {
		return fmt.Errorf("failed to create extension migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create extension migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("extension postgres migration failed: %w", err)
	}

	return nil
}

func runMigrationsSQLite(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at DATETIME DEFAULT (datetime('now'))
	)`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	entries, err := migrationsSqliteFS.ReadDir("sqlite")
	if err != nil {
		return fmt.Errorf("failed to read sqlite migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		version := strings.TrimSuffix(file, ".up.sql")

		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration version %s: %w", version, err)
		}
		if count > 0 {
			continue
		}

		content, err := migrationsSqliteFS.ReadFile("sqlite/" + file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", file, err)
			}
		}

		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("failed to record migration version %s: %w", version, err)
		}
	}

	return nil
}

func runMigrationsEmbeddedClickhouse(chDB *sql.DB) error {
	_, err := chDB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations_ch (
		version String,
		applied_at DateTime DEFAULT now()
	) ENGINE = MergeTree() ORDER BY version`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations_ch table: %w", err)
	}

	entries, err := migrationsChFS.ReadDir("ch")
	if err != nil {
		return fmt.Errorf("failed to read ch migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		version := strings.TrimSuffix(file, ".up.sql")

		rows, err := chDB.Query("SELECT count() FROM schema_migrations_ch WHERE version = ?", version)
		if err != nil {
			return fmt.Errorf("failed to check migration version %s: %w", version, err)
		}
		var count int
		if rows.Next() {
			if err := rows.Scan(&count); err != nil {
				rows.Close()
				return fmt.Errorf("failed to scan migration count for %s: %w", version, err)
			}
		}
		rows.Close()
		if count > 0 {
			continue
		}

		content, err := migrationsChFS.ReadFile("ch/" + file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		stmt := strings.TrimSpace(string(content))
		if stmt == "" {
			continue
		}
		if _, err := chDB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		if _, err := chDB.Exec("INSERT INTO schema_migrations_ch (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("failed to record migration version %s: %w", version, err)
		}
	}

	return nil
}

func Run(dbType string) error {
	cfg := config.Config

	// Run ClickHouse migrations
	if cfg.ClickhouseType == "embedded" {
		if err := runMigrationsEmbeddedClickhouse(chdb.EmbeddedDB); err != nil {
			return fmt.Errorf("embedded clickhouse migrations failed: %w", err)
		}
	} else {
		tlsConfig := "&secure=true"
		if cfg.ClickhouseTLS == "false" {
			tlsConfig = ""
		}

		err := runMigrationsClickhouse(fmt.Sprintf(`clickhouse://%s?username=%s&password=%s&database=%s%s`, cfg.ClickhouseServer, url.QueryEscape(cfg.ClickhouseUsername), url.QueryEscape(cfg.ClickhousePassword), cfg.ClickhouseDatabase, tlsConfig))
		if err != nil {
			return fmt.Errorf("clickhouse migrations failed: %w", err)
		}
	}

	if dbType == "sqlite" {
		if err := runMigrationsSQLite(db.DB); err != nil {
			return fmt.Errorf("sqlite migrations failed: %w", err)
		}

		return nil
	}

	// Run PostgreSQL migrations
	pgPort := cfg.PostgresPort
	pgSSLMode := cfg.PostgresSSLMode

	if pgSSLMode == "" {
		pgSSLMode = "disable"
	}
	if pgPort == "" {
		pgPort = "5432"
	}

	pgConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(cfg.PostgresUsername), url.QueryEscape(cfg.PostgresPassword), cfg.PostgresHost, pgPort, cfg.PostgresDatabase, pgSSLMode)

	if err := runMigrationsPostgres(pgConnStr); err != nil {
		return fmt.Errorf("postgres migrations failed: %w", err)
	}

	return nil
}
