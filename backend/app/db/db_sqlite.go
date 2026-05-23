//go:build !pgch

package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/tracewayapp/traceway/backend/app/config"
	"github.com/tracewayapp/lit/v2"
	_ "modernc.org/sqlite"
)

func Init() error {
	cfg := config.Config
	if cfg.DBType == "sqlite" {
		return initSQLite()
	}
	return initPostgres()
}

func initSQLite() error {
	path := config.Config.SQLitePath
	if path == "" {
		path = "./traceway.db"
	}

	mainDB, err := openSQLite(path, false)
	if err != nil {
		return err
	}
	DB = mainDB
	Driver = lit.SQLite
	config.Logf("SQLite database opened at %s", path)

	telemetryPath := strings.TrimSuffix(path, ".db") + "_telemetry.db"
	if path == ":memory:" {
		telemetryPath = ":memory:"
	}
	telDB, err := openSQLite(telemetryPath, true)
	if err != nil {
		return err
	}
	TelemetryDB = telDB
	config.Logf("SQLite telemetry database opened at %s", telemetryPath)

	return nil
}

// openSQLite opens a SQLite database with pragmas tuned for its role.
// The main DB gets foreign-key enforcement and the safest fsync mode; the
// telemetry DB relaxes synchronous to NORMAL and grows the page cache because
// it is append-only and cheaply re-derivable from clients on a hard crash.
//
// Pragmas go on the DSN (modernc.org/sqlite supports `_pragma=name(value)`)
// rather than via db.Exec because Exec only configures the first pooled
// connection — once SetMaxOpenConns goes above 1 new connections would
// otherwise boot without the pragmas applied.
func openSQLite(path string, telemetry bool) (*sql.DB, error) {
	var dsn string
	if path == ":memory:" {
		dsn = path
	} else {
		params := []string{
			"_pragma=journal_mode(WAL)",
			"_pragma=busy_timeout(5000)",
		}
		if telemetry {
			params = append(params,
				"_pragma=synchronous(NORMAL)",
				"_pragma=cache_size(-524288)",
				"_pragma=temp_store(MEMORY)",
				"_pragma=mmap_size(1073741824)",
				"_pragma=wal_autocheckpoint(50000)",
			)
		} else {
			params = append(params, "_pragma=foreign_keys(ON)")
		}
		dsn = "file:" + path + "?" + strings.Join(params, "&")
	}

	d, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite at %s: %w", path, err)
	}
	if err := d.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite at %s: %w", path, err)
	}

	if path == ":memory:" {
		if !telemetry {
			if _, err := d.Exec("PRAGMA foreign_keys = ON"); err != nil {
				return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
			}
		}
		if _, err := d.Exec("PRAGMA journal_mode = WAL"); err != nil {
			return nil, fmt.Errorf("failed to set WAL mode: %w", err)
		}
	}

	if telemetry {
		// WAL allows concurrent readers; SQLite still serializes writes at the
		// file level and busy_timeout absorbs short contention windows.
		d.SetMaxOpenConns(4)
	} else {
		d.SetMaxOpenConns(1)
	}
	return d, nil
}
