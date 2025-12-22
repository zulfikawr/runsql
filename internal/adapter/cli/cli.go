package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runsql/internal/core"
	"runsql/internal/parsers"
	"runsql/internal/ui"
	"strings"

	"github.com/xuri/excelize/v2"
)

// CLIConfig holds the CLI command-line arguments
type CLIConfig struct {
	FilePaths []string // -f: File paths (comma separated)
	Query     string   // -q: SQL query
	OutputFmt string   // -o: Output format (table, json, csv)
}

// Run executes the CLI workflow
func Run(config CLIConfig) error {
	// Validate inputs
	if len(config.FilePaths) == 0 {
		return fmt.Errorf("file path is required (-f)")
	}

	if config.Query == "" {
		// Default to selecting from the first table if available
		if len(config.FilePaths) > 0 {
			firstTable := getTableNameFromPath(config.FilePaths[0])
			config.Query = fmt.Sprintf("SELECT * FROM %s", firstTable)
		} else {
			config.Query = "SELECT 1" // Fallback if no files (though validation prevents this)
		}
	}

	if config.OutputFmt == "" {
		config.OutputFmt = "table"
	}

	// Colors
	c := ui.Colors

	// Step 1: Create engine
	engine, err := core.NewEngine()
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer engine.Close()

	// Step 2: Load all files

	for _, path := range config.FilePaths {
		fmt.Fprintf(os.Stderr, "%sProcessing %s%s%s...%s\n", c.Yellow, c.White, c.Bold, path, c.Reset)

		// Detect file type and create appropriate parser
		source, err := getSourceFromFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse file '%s': %w", path, err)
		}

		// Derive table name from filename
		tableName := getTableNameFromPath(path)

		err = engine.Load(tableName, source)
		if err != nil {
			return fmt.Errorf("failed to load data from '%s': %w", path, err)
		}

		fmt.Fprintf(os.Stderr, "%s✓%s Loaded '%s' as table '%s'\n", c.Green, c.Reset, path, tableName)
	}

	// Step 3: Execute query
	columns, rows, err := engine.Query(config.Query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	// Step 4: Format and output results
	return formatOutput(config.OutputFmt, columns, rows)
}

// getTableNameFromPath derives a table name from a file path
func getTableNameFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return sanitizeTableName(name)
}

func sanitizeTableName(name string) string {
	var sb strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	return sb.String()
}

// getSourceFromFile detects file type and returns appropriate parser
func getSourceFromFile(filePath string) (parsers.Source, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".csv":
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		return parsers.NewCSVSource(file)

	case ".json":
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		return parsers.NewJSONSource(file)

	case ".xlsx":
		file, err := excelize.OpenFile(filePath)
		if err != nil {
			return nil, err
		}
		return parsers.NewXLSXSource(file)

	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// formatOutput handles different output formats
func formatOutput(format string, columns []string, rows [][]interface{}) error {
	switch strings.ToLower(format) {
	case "table":
		return outputTable(columns, rows)
	case "json":
		return outputJSON(columns, rows)
	case "csv":
		return outputCSV(columns, rows)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// outputTable renders results as a styled text table
func outputTable(columns []string, rows [][]interface{}) error {
	if len(columns) == 0 {
		return nil
	}

	c := ui.Colors

	// Calculate column widths
	colWidths := make([]int, len(columns))
	for i, col := range columns {
		colWidths[i] = len(col)
	}

	for _, row := range rows {
		for i, val := range row {
			str := fmt.Sprintf("%v", val)
			if len(str) > colWidths[i] {
				colWidths[i] = len(str)
			}
		}
	}

	// Print separator function
	printSeparator := func() {
		fmt.Print(c.Dim)
		fmt.Print("+")
		for i := range columns {
			fmt.Print(strings.Repeat("-", colWidths[i]+2))
			fmt.Print("+")
		}
		fmt.Println(c.Reset)
	}

	// Print header
	printSeparator()
	fmt.Print(c.Dim + "| " + c.Reset)
	for i, col := range columns {
		fmt.Printf("%-*s %s ", colWidths[i], c.Cyan+c.Bold+col+c.Reset, c.Dim+"|"+c.Reset)
	}
	fmt.Println()
	printSeparator()

	// Print rows
	for _, row := range rows {
		fmt.Print(c.Dim + "| " + c.Reset)
		for i, val := range row {
			str := fmt.Sprintf("%v", val)
			fmt.Printf("%-*s %s ", colWidths[i], str, c.Dim+"|"+c.Reset)
		}
		fmt.Println()
	}
	printSeparator()

	return nil
}

// outputJSON renders results as JSON
func outputJSON(columns []string, rows [][]interface{}) error {
	// Convert to array of objects
	result := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		obj := make(map[string]interface{})
		for j, col := range columns {
			if j < len(row) {
				obj[col] = row[j]
			}
		}
		result[i] = obj
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonData))
	return nil
}

// outputCSV renders results as CSV
func outputCSV(columns []string, rows [][]interface{}) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	writer.Write(columns)

	// Write rows
	for _, row := range rows {
		strRow := make([]string, len(row))
		for i, val := range row {
			strRow[i] = fmt.Sprintf("%v", val)
		}
		writer.Write(strRow)
	}

	return nil
}
