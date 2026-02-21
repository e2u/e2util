# e2var

## Overview
`e2var` provides utilities for handling environment variables, configuration variables, and default value management.

## Installation
```bash
go get github.com/e2u/e2util/e2var
```

## Usage
```go
import "github.com/e2u/e2util/e2var"

value := e2var.GetString("APP_ENV", "development")
```

## Examples
*Show retrieving string, int, and bool env vars with defaults.*

## API Reference
*Exported functions and types.*
