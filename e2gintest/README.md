# e2gintest

## Overview
`e2gintest` provides testing utilities for Gin HTTP handlers, including request builders and response assertions.

## Installation
```bash
go get github.com/e2u/e2util/e2gintest
```

## Usage
```go
import "github.com/e2u/e2util/e2gintest"

router := gin.New()
// register routes
recorder := e2gintest.NewRecorder()
router.ServeHTTP(recorder, request)
```

## Examples
*Example of testing a JSON endpoint.*

## API Reference
*Exported functions and types.*
