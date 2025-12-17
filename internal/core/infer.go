package core

import (
	"regexp"
	"strings"
)

var (
	intRegex   = regexp.MustCompile(`^-?\d+$`)
	floatRegex = regexp.MustCompile(`^-?\d*\.\d+$`)
)

// InferType determines the SQLite type for a string value.
func InferType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "TEXT" // Default to TEXT for empty/null, though could be NULL sensitive
	}

	if intRegex.MatchString(value) {
		return "INTEGER"
	}

	if floatRegex.MatchString(value) {
		return "REAL"
	}

	return "TEXT"
}
