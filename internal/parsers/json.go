package parsers

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// JSONSource implements the Source interface for JSON files.
type JSONSource struct {
	decoder *json.Decoder
	headers []string
	firstRow []interface{}
}

// NewJSONSource creates a new JSONSource from an io.Reader.
// It expects a JSON array of objects.
func NewJSONSource(r io.Reader) (*JSONSource, error) {
	dec := json.NewDecoder(r)

	// Expect start of array '['
	if token, err := dec.Token(); err != nil || token != json.Delim('[') {
		return nil, fmt.Errorf("expected JSON array start, got %v", token)
	}

	if !dec.More() {
		return &JSONSource{decoder: dec, headers: []string{}}, nil
	}

	// Read first object to infer headers
	var firstObj map[string]interface{}
	if err := dec.Decode(&firstObj); err != nil {
		return nil, fmt.Errorf("failed to decode first JSON object: %w", err)
	}

	// Extract headers and sort them for stability
	var headers []string
	for k := range firstObj {
		headers = append(headers, k)
	}
	sort.Strings(headers)

	// Create first row based on sorted headers
	row := make([]interface{}, len(headers))
	for i, header := range headers {
		row[i] = firstObj[header]
	}

	return &JSONSource{
		decoder:  dec,
		headers:  headers,
		firstRow: row,
	}, nil
}

// GetHeaders returns the inferred column names.
func (s *JSONSource) GetHeaders() ([]string, error) {
	return s.headers, nil
}

// Read streams rows from the JSON array.
func (s *JSONSource) Read() (chan []interface{}, error) {
	out := make(chan []interface{})

	go func() {
		defer close(out)

		// Emit the first row we already read
		if s.firstRow != nil {
			out <- s.firstRow
			s.firstRow = nil // Clear it so we don't re-emit if Read is called again (though it shouldn't be)
		}

		for s.decoder.More() {
			var obj map[string]interface{}
			if err := s.decoder.Decode(&obj); err != nil {
				// Stop on error
				break
			}

			// Map object to row based on headers
			row := make([]interface{}, len(s.headers))
			for i, header := range s.headers {
				row[i] = obj[header]
			}
			out <- row
		}

		// Consume closing ']'
		_, _ = s.decoder.Token()
	}()

	return out, nil
}
