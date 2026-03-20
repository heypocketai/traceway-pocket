//go:build !pgch

package migrations

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/tracewayapp/traceway/backend/app/db"
)

type ExtensionMigration struct {
	Source embed.FS
	Path   string
	Table  string
}

var ExtensionPostgresMigrations []ExtensionMigration

//go:embed sqlite/*.sql
var migrationsSqliteFS embed.FS

func Run(dbType string) error {
	sqliteDB := db.DB

	_, err := sqliteDB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
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
		err := sqliteDB.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
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
			if _, err := sqliteDB.Exec(stmt); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", file, err)
			}
		}

		if _, err := sqliteDB.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("failed to record migration version %s: %w", version, err)
		}
	}

	return nil
}
