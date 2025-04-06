# e2concurrent Documentation

## 項目概覽 / Project Overview

`e2concurrent` 是 `e2util` 工具庫中的一個子包，提供了一個基於泛型的並發任務執行框架。它支持通過任務隊列並發執行多個任務，允許自定義並發數、緩衝區大小、超時和上下文控制，並提供任務執行結果和追蹤信息。此包適用於需要高效並發處理的場景，例如批量數據處理或並行計算。

`e2concurrent` is a sub-package of the `e2util` library, providing a generic-based concurrent task execution framework. It supports executing multiple tasks concurrently via a task queue, allowing customization of concurrency limits, buffer sizes, timeouts, and context control, while providing task execution results and tracing information. This package is suitable for scenarios requiring efficient concurrent processing, such as batch data processing or parallel computation.

---

## 使用方法 / Usage

### 1. 定義並執行簡單任務 / Defining and Executing a Simple Task

Define a task using a custom `WorkFunc` and execute it with `DefaultExec`.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2concurrent"
)

type MyTask struct{}

func (m *MyTask) ConcurrentRun(arg e2concurrent.Arg[int]) e2concurrent.Result[int] {
// Double the input value
return e2concurrent.Result[int]{Value: arg.Value * 2}
}

func main() {
// Create a task channel
tasks := make(chan e2concurrent.Task[int, int], 1)
task := e2concurrent.NewTask(&MyTask{}, 5)
tasks <- task
close(tasks)

// Execute tasks with default settings
results := e2concurrent.DefaultExec(nil, tasks)
for result := range results {
fmt.Println("Result:", result.Value) // Result: 10
}
}
```

### 2. 設置任務超時 / Setting Task Timeout

Use `WithTimeout` to set a timeout for a task.

```go
package main

import (
"fmt"
"time"
"github.com/e2u/e2util/e2concurrent"
)

type SlowTask struct{}

func (s *SlowTask) ConcurrentRun(arg e2concurrent.Arg[int]) e2concurrent.Result[int] {
// Simulate a long-running task
time.Sleep(2 * time.Second)
return e2concurrent.Result[int]{Value: arg.Value}
}

func main() {
tasks := make(chan e2concurrent.Task[int, int], 1)
task := e2concurrent.NewTask(&SlowTask{}, 1, e2concurrent.WithTimeout[int, int](1*time.Second))
tasks <- task
close(tasks)

results := e2concurrent.DefaultExec(nil, tasks)
for result := range results {
fmt.Println("Error:", result.Err) // Error: context deadline exceeded
}
}
```

### 3. 自定義並發數和緩衝區 / Customizing Concurrency and Buffer Size

Use `Exec` to specify the maximum concurrency and buffer size for task execution.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2concurrent"
)

type SimpleTask struct{}

func (s *SimpleTask) ConcurrentRun(arg e2concurrent.Arg[int]) e2concurrent.Result[int] {
return e2concurrent.Result[int]{Value: arg.Value + 1}
}

func main() {
tasks := make(chan e2concurrent.Task[int, int], 2)
tasks <- e2concurrent.NewTask(&SimpleTask{}, 1)
tasks <- e2concurrent.NewTask(&SimpleTask{}, 2)
close(tasks)

// Execute tasks with max concurrency of 1 and buffer size of 2
results := e2concurrent.Exec(nil, 1, tasks, 2)
for result := range results {
fmt.Println("Result:", result.Value) // Result: 2, then 3
}
}
```

### 4. 保留任務輸入 / Retaining Task Input in Result

Use `WithRetainInput` to include the original input in the task result.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2concurrent"
)

type EchoTask struct{}

func (e *EchoTask) ConcurrentRun(arg e2concurrent.Arg[string]) e2concurrent.Result[string] {
return e2concurrent.Result[string]{Value: arg.Value}
}

func main() {
tasks := make(chan e2concurrent.Task[string, string], 1)
task := e2concurrent.NewTask(&EchoTask{}, "hello", e2concurrent.WithRetainInput[string, string](true))
tasks <- task
close(tasks)

results := e2concurrent.DefaultExec(nil, tasks)
for result := range results {
fmt.Println("Result:", result.Value, "Input:", result.Arg.Value) // Result: hello Input: hello
}
}
```

### 5. 獲取任務執行追蹤信息 / Retrieving Task Execution Trace

Access the `Trace` field in the result to get execution metadata.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2concurrent"
)

type TraceTask struct{}

func (t *TraceTask) ConcurrentRun(arg e2concurrent.Arg[int]) e2concurrent.Result[int] {
return e2concurrent.Result[int]{Value: arg.Value}
}

func main() {
tasks := make(chan e2concurrent.Task[int, int], 1)
task := e2concurrent.NewTask(&TraceTask{}, 42)
tasks <- task
close(tasks)

results := e2concurrent.DefaultExec(nil, tasks)
for result := range results {
fmt.Println("Result:", result.Value, "Duration:", result.Trace.Duration) // Result: 42 Duration: <duration>
	}
	}
	```

	---

	## 安裝步驟 / Installation Steps

	1. **確保 Go 環境**
	確認已安裝 Go（建議使用 1.16 或以上版本），並設置好 `GOPATH`。
	2. **下載項目**
	在終端運行以下命令：
	```bash
	go get -u github.com/e2u/e2util/e2concurrent
	```
	3. **驗證安裝**
	在代碼中導入 `github.com/e2u/e2util/e2concurrent`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。

	1. **Ensure Go Environment**
	Confirm Go (version 1.16 or higher recommended) is installed and `GOPATH` is set.
	2. **Download the Package**
	Run the following command in your terminal:
	```bash
	go get -u github.com/e2u/e2util/e2concurrent
	```
	3. **Verify Installation**
	Import `github.com/e2u/e2util/e2concurrent` in your code and run `go build` or `go run`. Success indicates proper installation.

	---

	## 功能描述 / Features

	- **並發任務執行**：`Exec` 和 `DefaultExec` 支持以指定並發數執行任務，並提供結果通道。
	- **任務配置**：支持通過 `WithTimeout`、`WithContext`、`WithRefer` 和 `WithRetainInput` 自定義任務行為。
	- **泛型支持**：使用泛型實現類型安全的輸入和輸出，適配不同數據類型。
	- **執行追蹤**：`Trace` 結構記錄任務執行時間和元數據，便於調試和監控。
	- **上下文控制**：支持通過上下文（`context.Context`）取消任務或設置超時。
	- **錯誤處理**：任務執行結果包含錯誤信息，支持超時和取消錯誤。

	- **Concurrent Task Execution**: `Exec` and `DefaultExec` support executing tasks with a specified concurrency limit, providing a result channel.
	- **Task Configuration**: Supports customizing task behavior with `WithTimeout`, `WithContext`, `WithRefer`, and `WithRetainInput`.
	- **Generic Support**: Uses generics for type-safe input and output, adapting to different data types.
	- **Execution Tracing**: The `Trace` struct records task execution timing and metadata for debugging and monitoring.
	- **Context Control**: Supports task cancellation or timeout via `context.Context`.
	- **Error Handling**: Task results include error information, supporting timeout and cancellation errors.

	---
	```
