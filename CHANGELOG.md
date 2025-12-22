# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2025-12-23

### Added

- **Multi-File Support**: Upload and query multiple files (CSV, JSON, XLSX) simultaneously
- **JOIN Capabilities**: Execute SQL JOINs across different files (tables are named after filenames)
- **Enhanced Web UI**:
  - **Card Layout**: Improved schema display with collapsible cards and badges
  - **Syntax Highlighting**: Add Rainbow CSV, and VSCode-style JSON syntax highlighting output
  - **Skeleton Loading**: Visual feedback while loading large files
- **CLI TUI Overhaul**:
  - **Custom Help**: Detailed, color-coded `--help` output
  - **Colored Output**: Status messages and table borders now use a color scheme

### Changes

- **API Update**: `/query` endpoint now accepts list of files instead of single file
- **CLI**: `-f` flag now supports comma-separated paths (e.g., `-f file1.csv,file2.json`)

[2.0.0]: https://github.com/zulfikawr/runsql/releases/tag/v2.0.0

---

## [1.0.0] - 2025-12-17

### Added

- Initial release of RunSQL
- **CLI Mode**: Execute SQL queries from the terminal with Unix philosophy
- **Web Mode**: Spin up a localhost server with a GUI for non-technical users
- **Multi-Format Support**: Parse and query CSV, XLSX, and JSON files
- **In-Memory SQLite**: Load files into SQLite for fast querying
- **Type Inference**: Automatically detect column types (INTEGER, REAL, TEXT)
- **Multiple Output Formats**: Table, JSON, or CSV output
- **Hexagonal Architecture**: Clean separation of concerns (Ports & Adapters)
- GitHub CI/CD workflows for automated testing and building
- Cross-platform support: Windows, Linux, macOS
- Pre-built binaries for multiple architectures (amd64, arm64)

### Features

- Parse CSV, XLSX, and JSON files with automatic type detection
- Query files using standard SQL syntax
- Interactive web interface for ease of use
- Command-line interface for automation and scripting
- Support for Go 1.21, 1.22, and 1.25

[1.0.0]: https://github.com/zulfikAWR/runsql/releases/tag/v1.0.0
