package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"runsql/internal/core"
	"runsql/internal/parsers"

	"github.com/xuri/excelize/v2"
)

// QueryRequest represents the request body for /query
type QueryRequest struct {
	Query  string
	Format string
}

// QueryResponse represents the response from /query
type QueryResponse struct {
	Status  string          `json:"status"`
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	TimeMs  int64           `json:"time_ms"`
	Error   string          `json:"error,omitempty"`
}

// Server handles the web interface
type Server struct {
	addr string
}

// NewServer creates a new web server
func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

// Start starts the web server
func (s *Server) Start() error {
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/schema", s.handleSchema)
	http.HandleFunc("/query", s.handleQuery)

	// Serve static files (CSS, JS)
	http.Handle("/style.css", http.HandlerFunc(s.handleStaticFile("style.css", "text/css")))
	http.Handle("/script.js", http.HandlerFunc(s.handleStaticFile("script.js", "application/javascript")))

	fmt.Printf("Starting web server on http://localhost%s\n", s.addr)
	return http.ListenAndServe(s.addr, nil)
}

// handleHealth is a simple health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// handleStaticFile returns a handler function for serving static files
func (s *Server) handleStaticFile(filename string, contentType string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(getWebDir(), filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Write(data)
	}
}

// handleIndex serves the HTML frontend
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Read the index.html file from disk
	indexPath := filepath.Join(getWebDir(), "index.html")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		http.Error(w, "Failed to load index", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

// handleSchema returns the schema of an uploaded file
func (s *Server) handleSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Parse multipart form
	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil { // 10MB max
		respondError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get files
	if r.MultipartForm == nil || len(r.MultipartForm.File["file"]) == 0 {
		respondError(w, "At least one file is required", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["file"]

	// Create engine
	engine, err := core.NewEngine()
	if err != nil {
		respondError(w, "Failed to create engine", http.StatusInternalServerError)
		return
	}
	defer engine.Close()

	schemas := make(map[string][]string)
	tmpFiles := []string{}
	defer func() {
		for _, f := range tmpFiles {
			os.Remove(f)
		}
	}()

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			respondError(w, fmt.Sprintf("Failed to open file %s", fileHeader.Filename), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Write temp file
		tmpDir := os.TempDir()
		tmpFile := filepath.Join(tmpDir, fileHeader.Filename)
		tmpFiles = append(tmpFiles, tmpFile)

		tmpF, err := os.Create(tmpFile)
		if err != nil {
			respondError(w, "Failed to create temp file", http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(tmpF, file); err != nil {
			tmpF.Close()
			respondError(w, "Failed to write temp file", http.StatusInternalServerError)
			return
		}
		tmpF.Close()

		// Load file
		source, err := getSourceFromFile(tmpFile)
		if err != nil {
			respondError(w, fmt.Sprintf("Failed to parse file: %v", err), http.StatusBadRequest)
			return
		}

		tableName := getTableNameFromPath(fileHeader.Filename)
		if err := engine.Load(tableName, source); err != nil {
			respondError(w, fmt.Sprintf("Failed to load data: %v", err), http.StatusBadRequest)
			return
		}

		// Get schema
		columns, _, err := engine.Query(fmt.Sprintf("SELECT * FROM %s LIMIT 0", tableName))
		if err != nil {
			respondError(w, fmt.Sprintf("Failed to get schema for %s: %v", tableName, err), http.StatusBadRequest)
			return
		}

		schemas[tableName] = columns
		fmt.Printf("[WEB] Schema loaded for: %s\n", tableName)
	}

	// Return schemas
	response := map[string]interface{}{
		"status":  "success",
		"schemas": schemas,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleQuery processes file uploads and executes SQL queries
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Parse multipart form
	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil { // 10MB max
		respondError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get files
	if r.MultipartForm == nil || len(r.MultipartForm.File["file"]) == 0 {
		respondError(w, "At least one file is required", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["file"]

	// Get query
	query := r.FormValue("query")
	if query == "" {
		respondError(w, "Query is required", http.StatusBadRequest)
		return
	}

	format := r.FormValue("format")
	if format == "" {
		format = "table"
	}

	// Create engine
	engine, err := core.NewEngine()
	if err != nil {
		respondError(w, "Failed to create engine", http.StatusInternalServerError)
		return
	}
	defer engine.Close()

	// Process each file
	var loadedTables []string
	tmpFiles := []string{}
	defer func() {
		for _, f := range tmpFiles {
			os.Remove(f)
		}
	}()

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			respondError(w, fmt.Sprintf("Failed to open file %s", fileHeader.Filename), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Write temp file
		tmpDir := os.TempDir()
		tmpFile := filepath.Join(tmpDir, fileHeader.Filename)
		tmpFiles = append(tmpFiles, tmpFile)

		tmpF, err := os.Create(tmpFile)
		if err != nil {
			respondError(w, "Failed to create temp file", http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(tmpF, file); err != nil {
			tmpF.Close()
			respondError(w, "Failed to write temp file", http.StatusInternalServerError)
			return
		}
		tmpF.Close() // Close explicitly to flush

		// Load file into engine
		source, err := getSourceFromFile(tmpFile)
		if err != nil {
			respondError(w, fmt.Sprintf("Failed to parse file %s: %v", fileHeader.Filename, err), http.StatusBadRequest)
			return
		}

		// Derive table name
		tableName := getTableNameFromPath(fileHeader.Filename)

		if err := engine.Load(tableName, source); err != nil {
			respondError(w, fmt.Sprintf("Failed to load data from %s: %v", fileHeader.Filename, err), http.StatusBadRequest)
			return
		}
		loadedTables = append(loadedTables, tableName)
		fmt.Printf("[WEB] Loaded table: %s\n", tableName)
	}

	// Execute query
	columns, rows, err := engine.Query(query)
	if err != nil {
		respondError(w, fmt.Sprintf("Query error: %v", err), http.StatusBadRequest)
		return
	}

	elapsed := time.Since(startTime).Milliseconds()

	// Log query execution
	fmt.Printf("[WEB] SQL Query: %s\n", query)
	fmt.Printf("[WEB] Result: %d rows returned in %dms\n", len(rows), elapsed)

	// Return results
	response := QueryResponse{
		Status:  "success",
		Columns: columns,
		Rows:    rows,
		TimeMs:  elapsed,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// respondError writes an error response
func respondError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	response := QueryResponse{
		Status: "error",
		Error:  message,
	}
	json.NewEncoder(w).Encode(response)
}

// getWebDir returns the path to the web directory
func getWebDir() string {
	// Try to find the web directory relative to the current working directory
	if _, err := os.Stat("web"); err == nil {
		return "web"
	}
	// Fallback for when running from a different directory
	return filepath.Join(".", "web")
}

// getSourceFromFile detects the file type and returns the appropriate parser
func getSourceFromFile(filePath string) (parsers.Source, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".csv":
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open CSV file: %w", err)
		}
		source, err := parsers.NewCSVSource(file)
		if err != nil {
			file.Close()
			return nil, err
		}
		return source, nil

	case ".json":
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open JSON file: %w", err)
		}
		source, err := parsers.NewJSONSource(file)
		if err != nil {
			file.Close()
			return nil, err
		}
		return source, nil

	case ".xlsx":
		xlsxFile, err := excelize.OpenFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open XLSX file: %w", err)
		}
		source, err := parsers.NewXLSXSource(xlsxFile)
		if err != nil {
			xlsxFile.Close()
			return nil, err
		}
		return source, nil

	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
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
