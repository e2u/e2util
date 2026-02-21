# e2pprof

## Overview
`e2pprof` provides utilities for integrating Go's pprof profiling into services, including HTTP handlers and custom profiling controls.

## Installation
```bash
go get github.com/e2u/e2util/e2pprof
```

## Usage
```go
import "github.com/e2u/e2util/e2pprof"

router := gin.New()
e2pprof.Register(router)
```

## Examples
*Show enabling CPU and memory profiling via HTTP endpoints.*

## API Reference
*Exported functions and types.*
