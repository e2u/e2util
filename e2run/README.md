# e2run

## Overview
`e2run` provides utilities for building command‑line applications, including sub‑command grouping, flag parsing, and execution helpers.

## Installation
```bash
go get github.com/e2u/e2util/e2run
```

## Usage
```go
import "github.com/e2u/e2util/e2run"

app := e2run.NewApp("mycli")
app.AddCommand("serve", serveCmd)
app.Run()
```

## Examples
*Show creating a CLI with sub‑commands and flags.*

## API Reference
*Exported types and functions.*
