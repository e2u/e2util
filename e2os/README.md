# e2os

## Overview
`e2os` offers OS‑related helpers such as environment variable management, file path utilities, and signal handling.

## Installation
```bash
go get github.com/e2u/e2util/e2os
```

## Usage
```go
import "github.com/e2u/e2util/e2os"

val := e2os.GetEnv("HOME", "/tmp")
```

## Examples
*Show getting env vars with defaults and handling OS signals.*

## API Reference
*Exported functions and types.*
