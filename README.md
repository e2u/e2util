# e2util

**Important testing note**: The test suite now uses the PostgreSQL connection string:
```
host=pgsql-dev port=5432 user=pgsql password=123456 dbname=e2util_dev sslmode=disable TimeZone=UTC application_name=e2util
```
All integration tests that require a database should be run against this configuration.

(Existing documentation for individual packages remains unchanged.)