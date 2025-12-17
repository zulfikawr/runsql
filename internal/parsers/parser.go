package parsers

// Source is the interface that all file parsers must implement.
// It defines how data enters the system from various file formats.
type Source interface {
	// GetHeaders returns the list of column names found in the source.
	GetHeaders() ([]string, error)

	// Read returns a channel that streams rows of data.
	// Each row is a slice of empty interfaces to support mixed types (int, float, string, etc.).
	// The channel is closed when reading is complete or an error occurs.
	Read() (chan []interface{}, error)
}
