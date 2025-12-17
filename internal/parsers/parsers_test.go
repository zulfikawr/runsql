package parsers

import (
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestCSVSource(t *testing.T) {
	data := `id,name
1,Apple
2,Banana`
	r := strings.NewReader(data)

	src, err := NewCSVSource(r)
	if err != nil {
		t.Fatalf("NewCSVSource failed: %v", err)
	}

	headers, err := src.GetHeaders()
	if err != nil {
		t.Fatalf("GetHeaders failed: %v", err)
	}
	if len(headers) != 2 || headers[0] != "id" || headers[1] != "name" {
		t.Errorf("Unexpected headers: %v", headers)
	}

	ch, err := src.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	rowCount := 0
	for row := range ch {
		rowCount++
		if len(row) != 2 {
			t.Errorf("Row %d has wrong length: %d", rowCount, len(row))
		}
	}
	if rowCount != 2 {
		t.Errorf("Expected 2 rows, got %d", rowCount)
	}
}

func TestJSONSource(t *testing.T) {
	data := `[
		{"id": 1, "name": "Apple"},
		{"id": 2, "name": "Banana"}
	]`
	r := strings.NewReader(data)

	src, err := NewJSONSource(r)
	if err != nil {
		t.Fatalf("NewJSONSource failed: %v", err)
	}

	headers, err := src.GetHeaders()
	if err != nil {
		t.Fatalf("GetHeaders failed: %v", err)
	}
	// JSON keys are sorted alphabetically: id, name
	if len(headers) != 2 || headers[0] != "id" || headers[1] != "name" {
		t.Errorf("Unexpected headers: %v", headers)
	}

	ch, err := src.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	rowCount := 0
	for row := range ch {
		rowCount++
		if len(row) != 2 {
			t.Errorf("Row %d has wrong length: %d", rowCount, len(row))
		}
	}
	if rowCount != 2 {
		t.Errorf("Expected 2 rows, got %d", rowCount)
	}
}

func TestXLSXSource(t *testing.T) {
	f := excelize.NewFile()
	// Sheet1 is created by default. Get its index.
	index, err := f.GetSheetIndex("Sheet1")
	if err != nil {
		// NewFile should create Sheet1
		index, _ = f.NewSheet("Sheet1")
	}
	f.SetActiveSheet(index)

	// Set headers
	f.SetCellValue("Sheet1", "A1", "id")
	f.SetCellValue("Sheet1", "B1", "name")

	// Set data
	f.SetCellValue("Sheet1", "A2", 1)
	f.SetCellValue("Sheet1", "B2", "Apple")
	f.SetCellValue("Sheet1", "A3", 2)
	f.SetCellValue("Sheet1", "B3", "Banana")

	src, err := NewXLSXSource(f)
	if err != nil {
		t.Fatalf("NewXLSXSource failed: %v", err)
	}

	headers, err := src.GetHeaders()
	if err != nil {
		t.Fatalf("GetHeaders failed: %v", err)
	}
	if len(headers) != 2 || headers[0] != "id" || headers[1] != "name" {
		t.Errorf("Unexpected headers: %v", headers)
	}

	ch, err := src.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	rowCount := 0
	for range ch {
		rowCount++
	}
	if rowCount != 2 {
		t.Errorf("Expected 2 rows, got %d", rowCount)
	}
}
