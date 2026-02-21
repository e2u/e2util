# e2hash

## Overview
`e2hash` offers hashing utilities for passwords, tokens, and data integrity using bcrypt, SHA‑256, and other algorithms.

## Installation
```bash
go get github.com/e2u/e2util/e2hash
```

## Usage
```go
import "github.com/e2u/e2util/e2hash"

hash, _ := e2hash.BcryptPassword("secret")
```

## Examples
*Demonstrate password hashing and verification.*

## API Reference
*Exported functions and types.*
