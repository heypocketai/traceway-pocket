//go:build !nochdb

package chdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/tracewayapp/traceway/backend/app/config"

	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	chdblib "github.com/chdb-io/chdb-go/chdb"
	_ "github.com/chdb-io/chdb-go/chdb/driver"
	"github.com/google/uuid"
)

type embeddedConn struct {
	session *chdblib.Session
}

func initEmbedded() error {
	path := config.Config.ClickhousePath

	// Create session first (global singleton in chdb-go)
	var session *chdblib.Session
	var err error
	if path != "" {
		session, err = chdblib.NewSession(path)
	} else {
		session, err = chdblib.NewSession()
	}
	if err != nil {
		return fmt.Errorf("failed to create chdb session: %w", err)
	}

	// Also open sql.DB for migrations (reuses the same session singleton)
	dsn := ""
	if path != "" {
		dsn = "session=" + path
	}
	db, err := sql.Open("chdb", dsn)
	if err != nil {
		return fmt.Errorf("failed to open embedded clickhouse: %w", err)
	}
	EmbeddedDB = db

	Conn = &embeddedConn{session: session}
	log.Println("Initialized embedded ClickHouse (chdb)")
	return nil
}

// Query uses the native session API with JSONCompactStrings format,
// bypassing the Parquet driver that cannot serialize UUID columns.
func (c *embeddedConn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	interpolated := interpolateQuery(query, args)
	result, err := c.session.Query(interpolated, "JSONCompactStrings")
	if err != nil {
		return nil, err
	}

	buf := result.Buf()
	if len(buf) == 0 {
		result.Free()
		return &embeddedRows{}, nil
	}

	var parsed jsonResult
	if err := json.Unmarshal(buf, &parsed); err != nil {
		result.Free()
		return nil, fmt.Errorf("failed to parse query result: %w", err)
	}
	result.Free()

	cols := make([]string, len(parsed.Meta))
	for i, m := range parsed.Meta {
		cols[i] = m.Name
	}

	return &embeddedRows{columns: cols, data: parsed.Data}, nil
}

func (c *embeddedConn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	interpolated := interpolateQuery(query, args)
	result, err := c.session.Query(interpolated, "JSONCompactStrings")
	if err != nil {
		return &embeddedRow{err: err}
	}

	buf := result.Buf()
	if len(buf) == 0 {
		result.Free()
		return &embeddedRow{err: sql.ErrNoRows}
	}

	var parsed jsonResult
	if err := json.Unmarshal(buf, &parsed); err != nil {
		result.Free()
		return &embeddedRow{err: fmt.Errorf("failed to parse query result: %w", err)}
	}
	result.Free()

	if len(parsed.Data) == 0 {
		return &embeddedRow{err: sql.ErrNoRows}
	}

	return &embeddedRow{data: parsed.Data[0]}
}

func (c *embeddedConn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	return &embeddedBatch{session: c.session, query: query}, nil
}

func (c *embeddedConn) Exec(ctx context.Context, query string, args ...any) error {
	interpolated := interpolateQuery(query, args)
	result, err := c.session.Query(interpolated)
	if err != nil {
		return err
	}
	if result != nil {
		result.Free()
	}
	return nil
}

// --- Query interpolation ---

func interpolateQuery(query string, args []any) string {
	if len(args) == 0 {
		return query
	}
	for _, arg := range args {
		query = strings.Replace(query, "?", formatValue(arg), 1)
	}
	return query
}

// --- JSON result types ---

type jsonResult struct {
	Meta []jsonMeta  `json:"meta"`
	Data [][]*string `json:"data"`
	Rows int         `json:"rows"`
}

type jsonMeta struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// --- Rows adapter ---

type embeddedRows struct {
	columns []string
	data    [][]*string
	index   int
}

func (r *embeddedRows) Next() bool {
	r.index++
	return r.index <= len(r.data)
}

func (r *embeddedRows) Scan(dest ...any) error {
	if r.index < 1 || r.index > len(r.data) {
		return fmt.Errorf("no current row")
	}
	row := r.data[r.index-1]
	if len(dest) != len(row) {
		return fmt.Errorf("scan: expected %d destinations, got %d", len(row), len(dest))
	}
	for i, d := range dest {
		if err := scanValue(d, row[i]); err != nil {
			return fmt.Errorf("column %d: %w", i, err)
		}
	}
	return nil
}

func (r *embeddedRows) Close() error            { return nil }
func (r *embeddedRows) Err() error              { return nil }
func (r *embeddedRows) Columns() []string       { return r.columns }
func (r *embeddedRows) ScanStruct(dest any) error        { return fmt.Errorf("ScanStruct not supported in embedded mode") }
func (r *embeddedRows) ColumnTypes() []driver.ColumnType { return nil }
func (r *embeddedRows) Totals(dest ...any) error         { return nil }

// --- Row adapter ---

type embeddedRow struct {
	data []*string
	err  error
}

func (r *embeddedRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.data) {
		return fmt.Errorf("scan: expected %d destinations, got %d", len(r.data), len(dest))
	}
	for i, d := range dest {
		if err := scanValue(d, r.data[i]); err != nil {
			return fmt.Errorf("column %d: %w", i, err)
		}
	}
	return nil
}

func (r *embeddedRow) Err() error                { return r.err }
func (r *embeddedRow) ScanStruct(dest any) error { return fmt.Errorf("ScanStruct not supported in embedded mode") }

// --- Scan helper: convert JSON string values to Go types ---

func scanValue(dest any, src *string) error {
	if src == nil {
		return nil // NULL — leave dest as zero value
	}
	val := *src

	switch d := dest.(type) {
	case *string:
		*d = val
	case *uuid.UUID:
		id, err := uuid.Parse(val)
		if err != nil {
			return fmt.Errorf("parse UUID %q: %w", val, err)
		}
		*d = id
	case **uuid.UUID:
		id, err := uuid.Parse(val)
		if err != nil {
			return fmt.Errorf("parse UUID %q: %w", val, err)
		}
		*d = &id
	case *time.Time:
		for _, layout := range []string{
			"2006-01-02 15:04:05.000000000",
			"2006-01-02 15:04:05.000000",
			"2006-01-02 15:04:05.000",
			"2006-01-02 15:04:05",
		} {
			if t, err := time.Parse(layout, val); err == nil {
				*d = t
				return nil
			}
		}
		return fmt.Errorf("parse time %q: no matching layout", val)
	case *time.Duration:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		*d = time.Duration(v)
	case *bool:
		*d = val == "true" || val == "1"
	case *int:
		v, err := strconv.Atoi(val)
		if err != nil {
			return err
		}
		*d = v
	case *int8:
		v, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return err
		}
		*d = int8(v)
	case *int16:
		v, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return err
		}
		*d = int16(v)
	case *int32:
		v, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return err
		}
		*d = int32(v)
	case *int64:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		*d = v
	case *uint:
		v, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		*d = uint(v)
	case *uint8:
		v, err := strconv.ParseUint(val, 10, 8)
		if err != nil {
			return err
		}
		*d = uint8(v)
	case *uint16:
		v, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return err
		}
		*d = uint16(v)
	case *uint32:
		v, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return err
		}
		*d = uint32(v)
	case *uint64:
		v, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		*d = v
	case *float32:
		v, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return err
		}
		*d = float32(v)
	case *float64:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		*d = v
	default:
		return fmt.Errorf("unsupported scan type %T", dest)
	}
	return nil
}

// --- Batch adapter ---

type embeddedBatch struct {
	session *chdblib.Session
	query   string
	rows    [][]any
	sent    bool
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

	result, err := b.session.Query(sb.String())
	if err != nil {
		return err
	}
	if result != nil {
		result.Free()
	}
	b.sent = true
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

// --- Value formatting for batch INSERTs and query interpolation ---

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
		return "'" + val.UTC().Format("2006-01-02 15:04:05") + "'"
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
