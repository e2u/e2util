# e2sign

## Overview
`e2sign` implements digital signature utilities, including RSA/ECDSA signing and verification helpers.

## Installation
```bash
go get github.com/e2u/e2util/e2sign
```

## Usage
```go
import "github.com/e2u/e2util/e2sign"

sig, _ := e2sign.SignMessage(privateKey, []byte("data"))
valid := e2sign.VerifySignature(publicKey, []byte("data"), sig)
```

## Examples
*Show signing and verifying a message using RSA keys.*

## API Reference
*Exported functions and types.*
