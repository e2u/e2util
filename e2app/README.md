# e2app

## Overview
The `e2app` module provides utilities for application configuration and context management. It supports loading configuration from environment variables, command‑line flags, or configuration files (e.g., TOML) and integrates with other `e2util` modules such as `e2db`, `e2cache`, `e2http`, and `e2logrus`.

## Installation
```bash
go get github.com/e2u/e2util/e2app
```

## Usage
```go
import "github.com/e2u/e2util/e2app"

// Example: load configuration
cfg, err := e2app.LoadConfig()
if err != nil {
    // handle error
}
```

## Examples
* Loading configuration from a TOML file.
* Overriding configuration via environment variables.

## API Reference
* `LoadConfig() (*Config, error)` – Load configuration.
* `Config` – struct representing the application configuration.
* Additional helper functions and types are documented in the GoDoc.
