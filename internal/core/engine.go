package core

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
	"runsql/internal/parsers"
)

// Engine wraps the SQLite database and handles data loading and querying.
type Engine struct {
	db *sql.DB
}

// NewEngine creates a new in-memory SQLite engine.
func NewEngine() (*Engine, error) {
	// Connect to in-memory SQLite database
	// "cache=shared" allows multiple connections to the same in-memory DB (useful if we add concurrency later)
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Engine{db: db}, nil
}

// Load reads data from a source and loads it into a table.
func (e *Engine) Load(tableName string, source parsers.Source) error {
	// 1. Get Headers
	headers, err := source.GetHeaders()
	if err != nil {
		return fmt.Errorf("failed to get headers: %w", err)
	}

	// Sanitize headers: replace spaces with underscores, remove non-alphanumeric chars
	sanitizedHeaders := make([]string, len(headers))
	for i, h := range headers {
		sanitizedHeaders[i] = sanitizeHeader(h)
	}

	// 2. Read first batch to infer types
	// Since Read() returns a channel, we can't "peek" easily without consuming.
	// We'll read all data into a buffer first? NO, that defeats streaming.
	// Strategy:
	// - Read the first N rows into a buffer.
	// - Infer types from the buffer.
	// - Create Table.
	// - Insert buffered rows.
	// - Continue streaming the rest.
	
	const inferenceSampleSize = 100
	var bufferedRows [][]interface{}
	
	rowCh, err := source.Read()
	if err != nil {
		return fmt.Errorf("failed to start reading: %w", err)
	}

	columnTypes := make([]string, len(headers))
	// Initialize with "TEXT" as fallback
	for i := range columnTypes {
		columnTypes[i] = "TEXT"
	}

	// Read up to sample size
	for i := 0; i < inferenceSampleSize; i++ {
		row, ok := <-rowCh
		if !ok {
			break
		}
		bufferedRows = append(bufferedRows, row)
	}

	// Infer types based on buffered rows
	// If a column has ANY non-integer value (that isn't empty/null), it downgrades to Real or Text.
	// Hierarchy: INTEGER -> REAL -> TEXT
	for colIdx := range headers {
		isInt := true
		isReal := true

		for _, row := range bufferedRows {
			if colIdx >= len(row) {
				continue
			}
			val := fmt.Sprintf("%v", row[colIdx]) // Convert to string for regex check
			
			if val == "" {
				continue // Skip empty values
			}

			typeStr := InferType(val)
			if typeStr == "TEXT" {
				isInt = false
				isReal = false
				break
			}
			if typeStr == "REAL" {
				isInt = false
			}
		}

		if isInt {
			columnTypes[colIdx] = "INTEGER"
		} else if isReal {
			columnTypes[colIdx] = "REAL"
		} else {
			columnTypes[colIdx] = "TEXT"
		}
	}

	// 3. Create Table
	createSQL := buildCreateTableSQL(tableName, sanitizedHeaders, columnTypes)
	_, err = e.db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// 4. Insert Data (Buffered + Remaining)
	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	insertSQL := buildInsertSQL(tableName, sanitizedHeaders)
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// Insert buffered rows
	for _, row := range bufferedRows {
		if _, err := stmt.Exec(normalizeRow(row, len(headers))...); err != nil {
			return fmt.Errorf("failed to insert buffered row: %w", err)
		}
	}

	// Insert remaining rows
	for row := range rowCh {
		if _, err := stmt.Exec(normalizeRow(row, len(headers))...); err != nil {
			return fmt.Errorf("failed to insert row: %w", err)
		}
	}

	return tx.Commit()
}

// Query executes a SQL query and returns the results.
func (e *Engine) Query(query string) ([]string, [][]interface{}, error) {
	rows, err := e.db.Query(query)
	if err != nil {
		return nil, nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results [][]interface{}

	for rows.Next() {
		// Scan needs pointers to interfaces
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Prepare row for result, converting []byte to string if needed (SQLite often returns text as bytes)
		finalRow := make([]interface{}, len(columns))
		for i, v := range values {
			if b, ok := v.([]byte); ok {
				finalRow[i] = string(b)
			} else {
				finalRow[i] = v
			}
		}
		results = append(results, finalRow)
	}

	return columns, results, nil
}

// Close closes the database connection.
func (e *Engine) Close() error {
	return e.db.Close()
}

// Helpers

func sanitizeHeader(h string) string {
	h = strings.TrimSpace(h)
	// Replace spaces with underscores
	h = strings.ReplaceAll(h, " ", "_")
	// Remove anything that isn't alphanumeric or underscore
	// Simple approach: keep it simple for now
	return h
}

func buildCreateTableSQL(tableName string, headers []string, types []string) string {
	var cols []string
	for i, h := range headers {
		cols = append(cols, fmt.Sprintf(`"%s" %s`, h, types[i]))
	}
	return fmt.Sprintf(`CREATE TABLE "%s" (%s);`, tableName, strings.Join(cols, ", "))
}

func buildInsertSQL(tableName string, headers []string) string {
	placeholders := make([]string, len(headers))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	quotedHeaders := make([]string, len(headers))
	for i, h := range headers {
		quotedHeaders[i] = fmt.Sprintf(`"%s"`, h)
	}
	return fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s);`, 
		tableName, 
		strings.Join(quotedHeaders, ", "), 
		strings.Join(placeholders, ", "))
}

func normalizeRow(row []interface{}, length int) []interface{} {
	// Ensure row has correct length by padding or trimming
	if len(row) == length {
		return row
	}
	newRow := make([]interface{}, length)
	copy(newRow, row)
	return newRow
}
