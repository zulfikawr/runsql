package parsers

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// XLSXSource implements the Source interface for Excel files.
type XLSXSource struct {
	rows    *excelize.Rows
	headers []string
}

// NewXLSXSource creates a new XLSXSource from an excelize.File.
// It uses the active sheet.
func NewXLSXSource(f *excelize.File) (*XLSXSource, error) {
	// Get active sheet name
	sheetIndex := f.GetActiveSheetIndex()
	sheetName := f.GetSheetName(sheetIndex)

	// Get rows iterator
	rows, err := f.Rows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows for sheet %s: %w", sheetName, err)
	}

	// Read headers
	if !rows.Next() {
		return nil, fmt.Errorf("sheet %s is empty", sheetName)
	}
	headers, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	return &XLSXSource{
		rows:    rows,
		headers: headers,
	}, nil
}

// GetHeaders returns the column names.
func (s *XLSXSource) GetHeaders() ([]string, error) {
	return s.headers, nil
}

// Read streams rows from the Excel sheet.
func (s *XLSXSource) Read() (chan []interface{}, error) {
	out := make(chan []interface{})

	go func() {
		defer close(out)
		defer s.rows.Close()

		for s.rows.Next() {
			cols, err := s.rows.Columns()
			if err != nil {
				// Stop on error
				break
			}

			// Ensure row matches header length (pad with empty strings if needed)
			row := make([]interface{}, len(s.headers))
			for i := 0; i < len(s.headers); i++ {
				if i < len(cols) {
					row[i] = cols[i]
				} else {
					row[i] = ""
				}
			}
			out <- row
		}
	}()

	return out, nil
}
