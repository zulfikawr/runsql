package main

import (
	"flag"
	"fmt"
	"os"
	"runsql/internal/adapter/cli"
	"runsql/internal/adapter/web"
	"runsql/internal/ui"
	"strings"
)

func main() {
	// Custom Usage/Help Message
	flag.Usage = func() {
		// Colors aliases for brevity
		c := ui.Colors

		// Header
		fmt.Fprintf(os.Stderr, "\n  %s%s%s %s%s%s\n\n",
			c.Cyan, c.Bold, "runsql",
			c.Reset, c.White, "A hybrid CLI & Web tool to run SQL queries on CSV, XLSX, and JSON files, written in Go.")

		fmt.Fprintf(os.Stderr, "  %s%s:\n", c.Yellow, "Usage")
		fmt.Fprintf(os.Stderr, "    %srunsql [flags]%s\n\n", c.Reset, c.Reset)

		fmt.Fprintf(os.Stderr, "  %s%s:\n", c.Yellow, "Flags")

		printFlag := func(flagName, shorthand, description string, defaultVal any) {
			// Format: --flag, -f
			left := fmt.Sprintf("    --%s, -%s", flagName, shorthand)

			// Format the full line
			fmt.Fprintf(os.Stderr, "%s%-25s%s %s %s(default: %v)%s\n",
				c.Green, left, c.Reset,
				description,
				c.Dim, defaultVal, c.Reset)
		}

		printFlag("file", "f", " Input file paths (comma-separated for multiple files)", "\"\"")
		printFlag("query", "q", " SQL query to execute", "\"\"")
		printFlag("output", "o", " Output format (table, json, csv)", "table")
		printFlag("web", "web", " Start the web interface", "false")
		printFlag("addr", "addr", "Address for web server", ":8080")

		fmt.Fprintf(os.Stderr, "\n  %s%s:\n", c.Yellow, "Examples")
		fmt.Fprintf(os.Stderr, "    runsql -f users.csv -q \"SELECT * FROM users LIMIT 5\"\n")
		fmt.Fprintf(os.Stderr, "    runsql -f users.csv,orders.json -q \"SELECT * FROM users JOIN orders ON users.id = orders.user_id\"\n")
		fmt.Fprintf(os.Stderr, "    runsql -web -addr :9090\n\n")

		fmt.Fprint(os.Stderr, c.Reset)
	}

	// Define CLI flags
	filePath := flag.String("f", "", "File path (for CLI mode)")
	query := flag.String("q", "", "SQL query (for CLI mode)")
	outputFmt := flag.String("o", "table", "Output format: table, json, csv (for CLI mode)")
	webMode := flag.Bool("web", false, "Run in web mode (default: CLI mode)")
	addr := flag.String("addr", ":8080", "Web server address (for web mode)")

	flag.Parse()

	// Check for help flag equivalent helper (Go flag handles -h/-help automatically calling Usage, but we customized it)

	if *webMode {
		fmt.Printf("%sStarting web server on %s...%s\n", ui.Colors.Green, *addr, ui.Colors.Reset)
		server := web.NewServer(*addr)
		if err := server.Start(); err != nil {
			fmt.Printf("%sWeb server failed: %v%s\n", ui.Colors.Red, err, ui.Colors.Reset)
			os.Exit(1)
		}
	} else {
		// CLI mode
		filePaths := []string{}
		if *filePath != "" {
			parts := strings.SplitSeq(*filePath, ",")
			for p := range parts {
				trimmed := strings.TrimSpace(p)
				if trimmed != "" {
					filePaths = append(filePaths, trimmed)
				}
			}
		}

		config := cli.CLIConfig{
			FilePaths: filePaths,
			Query:     *query,
			OutputFmt: *outputFmt,
		}

		if err := cli.Run(config); err != nil {
			fmt.Printf("%sExecution failed: %v%s\n", ui.Colors.Red, err, ui.Colors.Reset)
			os.Exit(1)
		}
	}
}
