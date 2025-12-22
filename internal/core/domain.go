package core

// Table represents the metadata of a loaded table.
type Table struct {
	Name    string
	Columns []string
	Types   []string // e.g., "TEXT", "INTEGER", "REAL"
}
