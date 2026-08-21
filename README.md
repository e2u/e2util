# e2util

Go 工具庫／認證與 Web 輔助的 monorepo。每個子套件目錄都有中英雙語 `README.md`。

A Go utility monorepo (auth, web, AWS, DB, and small helpers). Each package directory has a bilingual `README.md`.

```bash
go get github.com/e2u/e2util
go test ./...
```

**資料庫測試 / Database tests**：需要 PostgreSQL：

```
host=pgsql-dev port=5432 user=pgsql password=123456 dbname=e2util_dev sslmode=disable TimeZone=UTC application_name=e2util
```
