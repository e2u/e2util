# e2context

## Overview
The `e2context` package offers utilities for managing request‑scoped contexts, cancellation, and value propagation.

## Installation
```bash
go get github.com/e2u/e2util/e2context
```

## Usage
```go
import "github.com/e2u/e2util/e2context"

ctx := e2context.WithValue(context.Background(), "key", "value")
```

## Examples
*Show usage of context cancellation and timeout helpers.*

## API Reference
*List exported symbols.*
