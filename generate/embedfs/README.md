# generate/embedfs

產生 `embed.FS` 包裝程式碼，把靜態目錄嵌進 Go 套件。

Code generator that wraps a directory as `embed.FS`.

## 用法 / Usage

```bash
go run ./generate/embedfs \
  -source ./assets \
  -output assets/generated.go \
  -package assets \
  -name EmbedFS \
  -patterns '*.js,*.css'
```

或在 `main.go`：

Or from `main.go`:

```go
//go:generate go run github.com/e2u/e2util/generate/embedfs -source ./assets
```

## 參數 / Flags

| 參數 / Flag | 預設 / Default | 說明 / Meaning |
| --- | --- | --- |
| `-source` | `.` | 來源目錄 / source directory |
| `-output` | `generated.go` | 輸出檔 / output file |
| `-package` | `assets` | 套件名 / package name |
| `-name` | `EmbedFS` | 變數名 / variable name |
| `-patterns` | `*.js,*.css` | 檔案 glob / file globs |
