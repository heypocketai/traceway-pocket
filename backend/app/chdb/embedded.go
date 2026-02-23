//go:build !nochdb

package chdb

import (
	"github.com/tracewayapp/traceway/backend/app/config"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	_ "github.com/chdb-io/chdb-go/chdb/driver"
	"github.com/google/uuid"
)

type embeddedConn struct {
	db *sql.DB
}

func initEmbedded() error {
	path := config.Config.ClickhousePath
	dsn := ""
	if path != "" {
		dsn = "session=" + path
	}
	db, err := sql.Open("chdb", dsn)
	if err != nil {
		return fmt.Errorf("failed to open embedded clickhouse: %w", err)
	}
	EmbeddedDB = db
	Conn = &embeddedConn{db: db}
	log.Println("Initialized embedded ClickHouse (chdb)")
	return nil
}

func (c *embeddedConn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	rows, err := c.db.QueryContext(ctx, query, sanitizeArgs(args)...)
	if err != nil {
		return nil, err
	}
	return &embeddedRows{rows: rows}, nil
}

func (c *embeddedConn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	row := c.db.QueryRowContext(ctx, query, sanitizeArgs(args)...)
	return &embeddedRow{row: row}
}

func (c *embeddedConn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	return &embeddedBatch{db: c.db, query: query}, nil
}

func (c *embeddedConn) Exec(ctx context.Context, query string, args ...any) error {
	_, err := c.db.ExecContext(ctx, query, sanitizeArgs(args)...)
	return err
}

func sanitizeArgs(args []any) []any {
	out := make([]any, len(args))
	for i, arg := range args {
		switch v := arg.(type) {
		case time.Time:
			out[i] = v.UTC().Format("2006-01-02 15:04:05")
		default:
			out[i] = arg
		}
	}
	return out
}

// --- Rows adapter ---

type embeddedRows struct {
	rows *sql.Rows
}

func (r *embeddedRows) Next() bool             { return r.rows.Next() }
func (r *embeddedRows) Scan(dest ...any) error  { return r.rows.Scan(dest...) }
func (r *embeddedRows) Close() error            { return r.rows.Close() }
func (r *embeddedRows) Err() error              { return r.rows.Err() }
func (r *embeddedRows) Columns() []string {
	cols, _ := r.rows.Columns()
	return cols
}
func (r *embeddedRows) ScanStruct(dest any) error        { return fmt.Errorf("ScanStruct not supported in embedded mode") }
func (r *embeddedRows) ColumnTypes() []driver.ColumnType { return nil }
func (r *embeddedRows) Totals(dest ...any) error         { return nil }

// --- Row adapter ---

type embeddedRow struct {
	row *sql.Row
}

func (r *embeddedRow) Scan(dest ...any) error    { return r.row.Scan(dest...) }
func (r *embeddedRow) Err() error                { return r.row.Err() }
func (r *embeddedRow) ScanStruct(dest any) error { return fmt.Errorf("ScanStruct not supported in embedded mode") }

// --- Batch adapter ---

type embeddedBatch struct {
	db    *sql.DB
	query string
	rows  [][]any
	sent  bool
}

func (b *embeddedBatch) Append(v ...any) error {
	b.rows = append(b.rows, v)
	return nil
}

func (b *embeddedBatch) Send() error {
	if len(b.rows) == 0 {
		b.sent = true
		return nil
	}

	var sb strings.Builder
	sb.WriteString(b.query)
	sb.WriteString(" VALUES ")

	for i, row := range b.rows {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("(")
		for j, v := range row {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(formatValue(v))
		}
		sb.WriteString(")")
	}

	_, err := b.db.Exec(sb.String())
	if err == nil {
		b.sent = true
	}
	return err
}

func (b *embeddedBatch) Abort() error                  { return nil }
func (b *embeddedBatch) AppendStruct(v any) error      { return fmt.Errorf("AppendStruct not supported in embedded mode") }
func (b *embeddedBatch) Column(int) driver.BatchColumn { return nil }
func (b *embeddedBatch) Flush() error                  { return nil }
func (b *embeddedBatch) IsSent() bool                  { return b.sent }
func (b *embeddedBatch) Rows() int                     { return len(b.rows) }
func (b *embeddedBatch) Columns() []column.Interface   { return nil }
func (b *embeddedBatch) Close() error                  { return nil }

// --- Value formatting for batch INSERTs ---

func formatValue(v any) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case string:
		return "'" + strings.ReplaceAll(val, "'", "\\'") + "'"
	case uuid.UUID:
		return "'" + val.String() + "'"
	case time.Time:
		return "'" + val.UTC().Format("2006-01-02 15:04:05.000") + "'"
	case bool:
		if val {
			return "1"
		}
		return "0"
	case int:
		return fmt.Sprintf("%d", val)
	case int8:
		return fmt.Sprintf("%d", val)
	case int16:
		return fmt.Sprintf("%d", val)
	case int32:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case uint:
		return fmt.Sprintf("%d", val)
	case uint8:
		return fmt.Sprintf("%d", val)
	case uint16:
		return fmt.Sprintf("%d", val)
	case uint32:
		return fmt.Sprintf("%d", val)
	case uint64:
		return fmt.Sprintf("%d", val)
	case float32:
		return fmt.Sprintf("%v", val)
	case float64:
		return fmt.Sprintf("%v", val)
	default:
		return "'" + strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "\\'") + "'"
	}
}
