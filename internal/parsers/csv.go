package parsers

import (
	"encoding/csv"
	"fmt"
	"io"
)

// CSVSource implements the Source interface for CSV files.
type CSVSource struct {
	reader *csv.Reader
	headers []string
}

// NewCSVSource creates a new CSVSource from an io.Reader.
// It assumes the first row contains headers.
func NewCSVSource(r io.Reader) (*CSVSource, error) {
	csvReader := csv.NewReader(r)
	// Allow variable field counts per record (some rows might have more/fewer fields)
	csvReader.FieldsPerRecord = -1
	
	// Read the first row as headers
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	return &CSVSource{
		reader:  csvReader,
		headers: headers,
	}, nil
}

// GetHeaders returns the column names.
func (s *CSVSource) GetHeaders() ([]string, error) {
	return s.headers, nil
}

// Read streams rows from the CSV file.
func (s *CSVSource) Read() (chan []interface{}, error) {
	out := make(chan []interface{})

	go func() {
		defer close(out)
		for {
			record, err := s.reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				// Log the error but continue reading for robustness
				// Some CSV files may have formatting issues on certain lines
				continue
			}

			// Convert []string to []interface{}
			row := make([]interface{}, len(record))
			for i, v := range record {
				row[i] = v
			}
			out <- row
		}
	}()

	return out, nil
}
