package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runsql/internal/core"
	"runsql/internal/parsers"
	"strings"

	"github.com/xuri/excelize/v2"
)

// CLIConfig holds the CLI command-line arguments
type CLIConfig struct {
	FilePath   string // -f: File path
	Query      string // -q: SQL query
	OutputFmt  string // -o: Output format (table, json, csv)
}

// Run executes the CLI workflow
func Run(config CLIConfig) error {
	// Validate inputs
	if config.FilePath == "" {
		return fmt.Errorf("file path is required (-f)")
	}

	if config.Query == "" {
		config.Query = "SELECT * FROM tbl" // Default query
	}

	if config.OutputFmt == "" {
		config.OutputFmt = "table" // Default output format
	}

	// Step 1: Detect file type and create appropriate parser
	source, err := getSourceFromFile(config.FilePath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Step 2: Create engine and load data
	engine, err := core.NewEngine()
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer engine.Close()

	err = engine.Load("tbl", source)
	if err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}

	// Step 3: Execute query
	columns, rows, err := engine.Query(config.Query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	// Step 4: Format and output results
	return formatOutput(config.OutputFmt, columns, rows)
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
		defer file.Close()
		return parsers.NewCSVSource(file)

	case ".json":
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
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

// outputTable renders results as a simple text table
func outputTable(columns []string, rows [][]interface{}) error {
	if len(columns) == 0 {
		return nil
	}

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

	// Print header
	fmt.Print("| ")
	for i, col := range columns {
		fmt.Printf("%-*s | ", colWidths[i], col)
	}
	fmt.Println()

	// Print separator
	fmt.Print("+")
	for i := range columns {
		fmt.Print(strings.Repeat("-", colWidths[i]+2))
		fmt.Print("+")
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		fmt.Print("| ")
		for i, val := range row {
			str := fmt.Sprintf("%v", val)
			fmt.Printf("%-*s | ", colWidths[i], str)
		}
		fmt.Println()
	}

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
