# e2bdb Documentation

## 項目概覽 / Project Overview

`e2bdb` 是 `e2util` 工具庫中的一個子包，基於 `badger` 鍵值數據庫，提供了一個簡單的文件存儲解決方案。它支持文件的存儲、加載、刪除和存在性檢查，並使用 SHA256 哈希驗證文件內容的完整性。此包適用於需要輕量級鍵值存儲的應用場景，例如文件緩存或小型數據存儲。

`e2bdb` is a sub-package of the `e2util` library, built on the `badger` key-value database, providing a simple file storage solution. It supports storing, loading, deleting, and checking the existence of files, using SHA256 hashing to verify the integrity of file content. This package is suitable for scenarios requiring lightweight key-value storage, such as file caching or small-scale data storage.

---

## 使用方法 / Usage

### 1. 初始化 Badger 數據庫 / Initializing Badger Database

使用 `New` 函數初始化一個 Badger 數據庫實例。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2bdb"
)

func main() {
// 初始化 Badger 數據庫
db, err := e2bdb.New("./badger")
if err != nil {
fmt.Println("初始化失敗:", err)
return
}
defer db.Close()
fmt.Println("數據庫初始化成功")
}
```

Use the `New` function to initialize a Badger database instance.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2bdb"
)

func main() {
// Initialize Badger database
db, err := e2bdb.New("./badger")
if err != nil {
fmt.Println("Failed to initialize:", err)
return
}
defer db.Close()
fmt.Println("Database initialized successfully")
}
```

### 2. 存儲文件 / Storing a File

使用 `StorageFile` 方法將文件存儲到 Badger 數據庫中。

```go
package main

import (
"bytes"
"fmt"
"github.com/e2u/e2util/e2bdb"
)

func main() {
db, err := e2bdb.New("./badger")
if err != nil {
fmt.Println("初始化失敗:", err)
return
}
defer db.Close()

// 創建一個文件對象
file := &e2bdb.File{
Key:     "file1",
Name:    "example.txt",
Reader:  bytes.NewReader([]byte("Hello, World!")),
}

// 存儲文件
if err := db.StorageFile(file); err != nil {
fmt.Println("存儲文件失敗:", err)
return
}
fmt.Println("文件存儲成功")
}
```

Use the `StorageFile` method to store a file in the Badger database.

```go
package main

import (
"bytes"
"fmt"
"github.com/e2u/e2util/e2bdb"
)

func main() {
db, err := e2bdb.New("./badger")
if err != nil {
fmt.Println("Failed to initialize:", err)
return
}
defer db.Close()

// Create a file object
file := &e2bdb.File{
Key:     "file1",
Name:    "example.txt",
Reader:  bytes.NewReader([]byte("Hello, World!")),
}

// Store the file
if err := db.StorageFile(file); err != nil {
fmt.Println("Failed to store file:", err)
return
}
fmt.Println("File stored successfully")
}
```

### 3. 加載文件 / Loading a File

使用 `LoadFile` 方法從 Badger 數據庫中加載文件。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2bdb"
)

func main() {
db, err := e2bdb.New("./badger")
if err != nil {
fmt.Println("初始化失敗:", err)
return
}
defer db.Close()

// 加載文件
file, err := db.LoadFile("file1")
if err != nil {
fmt.Println("加載文件失敗:", err)
return
}
fmt.Println("文件內容:", string(file.Content()))
}
```

Use the `LoadFile` method to load a file from the Badger database.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2bdb"
)

func main() {
db, err := e2bdb.New("./badger")
if err != nil {
fmt.Println("Failed to initialize:", err)
return
}
defer db.Close()

// Load the file
file, err := db.LoadFile("file1")
if err != nil {
fmt.Println("Failed to load file:", err)
return
}
fmt.Println("File content:", string(file.Content()))
}
```

### 4. 檢查文件是否存在 / Checking if a File Exists

使用 `Exists` 方法檢查指定鍵的文件是否存在。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2bdb"
)

func main() {
db, err := e2bdb.New("./badger")
if err != nil {
fmt.Println("初始化失敗:", err)
return
}
defer db.Close()

// 檢查文件是否存在
exists, err := db.Exists("file1")
if err != nil {
fmt.Println("檢查失敗:", err)
return
}
fmt.Println("文件是否存在:", exists)
}
```

Use the `Exists` method to check if a file with the specified key exists.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2bdb"
)

func main() {
db, err := e2bdb.New("./badger")
if err != nil {
fmt.Println("Failed to initialize:", err)
return
}
defer db.Close()

// Check if the file exists
exists, err := db.Exists("file1")
if err != nil {
fmt.Println("Failed to check:", err)
return
}
fmt.Println("File exists:", exists)
}
```

### 5. 刪除文件 / Deleting a File

使用 `Delete` 方法從 Badger 數據庫中刪除文件。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2bdb"
)

func main() {
db, err := e2bdb.New("./badger")
if err != nil {
fmt.Println("初始化失敗:", err)
return
}
defer db.Close()

// 刪除文件
if err := db.Delete("file1"); err != nil {
fmt.Println("刪除文件失敗:", err)
return
}
fmt.Println("文件刪除成功")
}
```

Use the `Delete` method to delete a file from the Badger database.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2bdb"
)

func main() {
db, err := e2bdb.New("./badger")
if err != nil {
fmt.Println("Failed to initialize:", err)
return
}
defer db.Close()

// Delete the file
if err := db.Delete("file1"); err != nil {
fmt.Println("Failed to delete file:", err)
return
}
fmt.Println("File deleted successfully")
}
```

---

## 安裝步驟 / Installation Steps

1. **確保 Go 環境**
確認已安裝 Go（建議使用 1.16 或以上版本），並設置好 `GOPATH`。
2. **下載項目**
在終端運行以下命令：
```bash
go get -u github.com/e2u/e2util/e2bdb
```
3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2bdb`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。

1. **Ensure Go Environment**
Confirm Go (version 1.16 or higher recommended) is installed and `GOPATH` is set.
2. **Download the Package**
Run the following command in your terminal:
```bash
go get -u github.com/e2u/e2util/e2bdb
```
3. **Verify Installation**
Import `github.com/e2u/e2util/e2bdb` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **數據庫初始化**：`New` 函數初始化 Badger 數據庫，支持自定義選項（如路徑和壓縮設置）。
- **文件存儲**：`StorageFile` 方法將文件存儲到數據庫，包含元數據（如名稱、大小、哈希）和內容。
- **文件加載**：`LoadFile` 方法從數據庫中加載文件，包括元數據和內容。
- **存在性檢查**：`Exists` 方法檢查指定鍵的文件是否存在，並驗證內容完整性。
- **文件刪除**：`Delete` 方法從數據庫中刪除文件的元數據和內容。
- **內容完整性**：使用 SHA256 哈希（`contentHash`）驗證文件內容的完整性。
- **錯誤處理**：所有方法均針對無效輸入（如缺失鍵或讀取錯誤）返回錯誤。

- **Database Initialization**: The `New` function initializes a Badger database, supporting custom options (e.g., path and compression settings).
- **File Storage**: The `StorageFile` method stores a file in the database, including metadata (e.g., name, size, hash) and content.
- **File Loading**: The `LoadFile` method loads a file from the database, including metadata and content.
- **Existence Check**: The `Exists` method checks if a file with the specified key exists and verifies content integrity.
- **File Deletion**: The `Delete` method removes a file's metadata and content from the database.
- **Content Integrity**: Uses SHA256 hashing (`contentHash`) to verify the integrity of file content.
- **Error Handling**: All methods return errors for invalid inputs (e.g., missing key or read errors).

---
```
