# e2logrus

## Overview
`e2logrus` wraps the Logrus logger with additional conveniences for structured logging across the e2util ecosystem.

## Installation
```bash
go get github.com/e2u/e2util/e2logrus
```

## Usage
```go
import "github.com/e2u/e2util/e2logrus"

log := e2logrus.NewLogger()
log.Info("application started")
```

## Examples
*Show context‑aware logging with fields.*

## API Reference
*Exported types and functions.*
