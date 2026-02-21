# e2regexp

## Overview
`e2regexp` offers helper functions for working with regular expressions, including compiled caches and safe matching utilities.

## Installation
```bash
go get github.com/e2u/e2util/e2regexp
```

## Usage
```go
import "github.com/e2u/e2util/e2regexp"

re := e2regexp.MustCompile(`^a.*b$`)
matched := re.MatchString("abc")
```

## Examples
*Show using compiled regex with caching.*

## API Reference
*Exported functions and types.*
