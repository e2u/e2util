# e2webdriver

## Overview
`e2webdriver` offers a thin wrapper around Selenium WebDriver for browser automation in Go.

## Installation
```bash
go get github.com/e2u/e2util/e2webdriver
```

## Usage
```go
import "github.com/e2u/e2util/e2webdriver"

wd, _ := e2webdriver.NewChrome()
wd.Get("https://example.com")
```

## Examples
*Show navigating to a page and extracting the title.*

## API Reference
*Exported types and functions.*
