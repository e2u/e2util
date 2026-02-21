# e2time

## Overview
`e2time` offers time‑related helper functions, such as parsing, formatting, duration calculations, and clock abstractions for testing.

## Installation
```bash
go get github.com/e2u/e2util/e2time
```

## Usage
```go
import "github.com/e2u/e2util/e2time"

now := e2time.Now()
formatted := now.Format(time.RFC3339)
```

## Examples
*Show parsing ISO8601 dates and using a mock clock.*

## API Reference
*Exported functions and types.*
