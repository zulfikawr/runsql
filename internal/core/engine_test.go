package core

import (
	"testing"
)

// MockSource for testing
type MockSource struct {
	headers []string
	rows    [][]interface{}
}

func (m *MockSource) GetHeaders() ([]string, error) {
	return m.headers, nil
}

func (m *MockSource) Read() (chan []interface{}, error) {
	ch := make(chan []interface{})
	go func() {
		defer close(ch)
		for _, row := range m.rows {
			ch <- row
		}
	}()
	return ch, nil
}

func TestEngine(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	source := &MockSource{
		headers: []string{"id", "data", "value"},
		rows: [][]interface{}{
			{1, "A", 1.1},
			{2, "B", 2.2},
			{3, "C", 3.3},
		},
	}

	err = engine.Load("test_table", source)
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	// Test Query
	cols, rows, err := engine.Query("SELECT * FROM test_table ORDER BY id ASC")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(cols) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(cols))
	}

	if len(rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}

	// Verify data
	if rows[0][0].(int64) != 1 { // SQLite driver returns int64
		t.Errorf("Expected id 1, got %v", rows[0][0])
	}
	if rows[0][1].(string) != "A" {
		t.Errorf("Expected data A, got %v", rows[0][1])
	}
}

func TestTypeInference(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123", "INTEGER"},
		{"-456", "INTEGER"},
		{"12.34", "REAL"},
		{"-56.78", "REAL"},
		{"abc", "TEXT"},
		{"123a", "TEXT"},
		{"", "TEXT"},
	}

	for _, tt := range tests {
		if got := InferType(tt.input); got != tt.expected {
			t.Errorf("InferType(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSanitizeHeader(t *testing.T) {
	h := sanitizeHeader("First Name")
	if h != "First_Name" {
		t.Errorf("Expected First_Name, got %s", h)
	}
}
