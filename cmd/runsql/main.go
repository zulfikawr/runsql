package main

import (
	"flag"
	"log"
	"runsql/internal/adapter/cli"
	"runsql/internal/adapter/web"
)

func main() {
	// Define CLI flags
	filePath := flag.String("f", "", "File path (for CLI mode)")
	query := flag.String("q", "", "SQL query (for CLI mode)")
	outputFmt := flag.String("o", "table", "Output format: table, json, csv (for CLI mode)")
	webMode := flag.Bool("web", false, "Run in web mode (default: CLI mode)")
	addr := flag.String("addr", ":8080", "Web server address (for web mode)")

	flag.Parse()

	if *webMode {
		// Web mode
		server := web.NewServer(*addr)
		if err := server.Start(); err != nil {
			log.Fatalf("Web server error: %v", err)
		}
	} else {
		// CLI mode
		config := cli.CLIConfig{
			FilePath:  *filePath,
			Query:     *query,
			OutputFmt: *outputFmt,
		}

		if err := cli.Run(config); err != nil {
			log.Fatalf("Error: %v", err)
		}
	}
}
