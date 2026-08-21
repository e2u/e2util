# e2html

用 Go 組合 HTML 標籤與屬性，會轉義文字內容，void 元素（如 `br`、`img`）只輸出開始標籤。

Build HTML tags and attributes in Go. Text is escaped; void elements (`br`, `img`, …) emit a start tag only.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2html
```

## 功能 / Features

- **標籤構建 / Tag builder**：`T`、`Div`、`Doctype`、`TS`
- **屬性 / Attributes**：`A` / `Attr`，布林屬性只在 true 時輸出
- **安全 / Safety**：標籤名白名單、文字 HTML 轉義

## 用法 / Usage

```go
import h "github.com/e2u/e2util/e2html"

page := h.TS([]h.TAG{
    h.TAG(h.Doctype("html")),
    h.T("html",
        h.T("body",
            h.Div(h.A("class", "box"), "hello"),
            h.T("br"),
            h.T("img", h.A("src", "/logo.png"), h.A("alt", "logo")),
        ),
    ),
})
fmt.Print(page)
```
