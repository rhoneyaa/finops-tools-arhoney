package snowflake

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// QueryResult is the outcome of a SQL query.
type QueryResult struct {
	Columns []string
	Rows    [][]string
}

// Query executes SQL and returns all rows as strings.
func Query(ctx context.Context, db *sql.DB, sqlText string) (QueryResult, error) {
	rows, err := db.QueryContext(ctx, sqlText)
	if err != nil {
		return QueryResult{}, fmt.Errorf("snowflake query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return QueryResult{}, err
	}
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return QueryResult{}, err
	}

	out := QueryResult{Columns: resolveColumnNames(cols, colTypes)}

	for rows.Next() {
		row, err := scanRowValues(rows, len(out.Columns))
		if err != nil {
			return QueryResult{}, err
		}
		out.Rows = append(out.Rows, row)
	}
	if err := rows.Err(); err != nil {
		return QueryResult{}, err
	}
	return out, nil
}

func resolveColumnNames(cols []string, colTypes []*sql.ColumnType) []string {
	names := make([]string, len(cols))
	for i, col := range cols {
		name := strings.TrimSpace(col)
		if name == "" && i < len(colTypes) && colTypes[i] != nil {
			name = strings.TrimSpace(colTypes[i].Name())
		}
		if name == "" {
			name = fmt.Sprintf("COLUMN_%d", i+1)
		}
		names[i] = name
	}
	return names
}

func scanRowValues(rows *sql.Rows, n int) ([]string, error) {
	holders := make([]any, n)
	ptrs := make([]any, n)
	for i := range holders {
		ptrs[i] = &holders[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		return nil, err
	}

	out := make([]string, n)
	for i, v := range holders {
		out[i] = valueString(v)
	}
	return out, nil
}

func valueString(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	case time.Time:
		return x.Format(time.RFC3339Nano)
	case bool:
		if x {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprint(x)
	}
}
