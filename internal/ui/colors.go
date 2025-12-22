package ui

import "os"

// ColorScheme holds ANSI color codes for terminal output
type ColorScheme struct {
	Reset   string
	Bold    string
	Dim     string
	Green   string
	Yellow  string
	Red     string
	Magenta string
	Cyan    string
	White   string
}

// Colors is the global color scheme instance
var Colors = initColors()

// initColors initializes color codes based on NO_COLOR environment variable
func initColors() ColorScheme {
	if os.Getenv("NO_COLOR") != "" {
		return ColorScheme{} // All empty strings when NO_COLOR is set
	}
	return ColorScheme{
		Reset:   "\033[0m",
		Bold:    "\033[1m",
		Dim:     "\033[2m",
		Green:   "\033[32m",
		Yellow:  "\033[33m",
		Red:     "\033[31m",
		Magenta: "\033[35m",
		Cyan:    "\033[36m",
		White:   "\033[97m", // HiWhite equivalent
	}
}
