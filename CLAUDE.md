# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

- **Build the module**
  ```bash
  go build ./...
  ```
  Compiles all packages in the repository.

- **Static analysis / lint**
  ```bash
  go vet ./...
  # optional, if staticcheck is installed
  staticcheck ./...
  ```

- **Run the full test suite**
  ```bash
  go test ./...
  ```
  The test suite expects a PostgreSQL instance using the connection string shown in the README.

- **Run a single test**
  ```bash
  go test ./e2http -run ^TestBuilderGetHtml$
  ```
  Replace the package path and test name as needed. The `-run` flag takes a regular expression.

- **Run tests for a specific package**
  ```bash
  go test ./e2auth
  ```

- **Run benchmarks**
  ```bash
  go test -bench . ./...
  ```

- **Generate embedded assets** (used by `generate/embedfs`)
  ```bash
  go run ./generate/embedfs -source ./assets -output assets/generated.go -package assets
  ```

## High‑Level Architecture

The repository is a Go monorepo that provides a collection of reusable utilities and a small authentication/web framework.

- **`e2auth`** – Core authentication layer. It defines the user model, session handling, password‑reset flow, email verification, MFA, OAuth integration, CAPTCHA verification, and related HTTP routers. The package is split into:
  - `e2auth/dao.go` – Data‑access functions (user lookup, session CRUD, password updates, admin check, etc.).
  - `e2auth/middleware.go` – Gin middleware for request authentication and admin authorization.
  - `e2auth/routers.go` – Public and protected route definitions.
  - `e2auth/extra_handlers.go` – Additional handlers (OAuth, CAPTCHA, account unlock, MFA, etc.) that were moved out to keep `routers.go` tidy.

- **`e2http`** – A lightweight HTTP client builder (`Context`). It supports configurable URLs, headers, timeouts, basic/bearer auth, request dumping, JSON unmarshalling, and optional proxy handling (currently a no‑op in tests).

- **`e2rest`** – Helper utilities for building REST style handlers (not detailed here but follows the same pattern of thin wrappers around `gin`.

- **`e2bdb`** – Database helper utilities and integration tests.

- **Utility packages** (`e2json`, `e2jwt`, `e2logrus`, `e2slice`, `e2map`, `e2math`, `e2regexp`, `e2sync`, `e2time`, `e2var`, etc.) – Small, focused helpers that are reused across the codebase. They generally expose a few exported functions/types and have corresponding unit tests.

- **`generate/embedfs`** – A code‑generation tool that creates an `embed.FS` wrapper for static assets. Used via `go run ./generate/embedfs` as shown above.

- **`e2run` / `e2run/groupedcmd`** – CLI scaffolding utilities for building command‑line tools that may group sub‑commands.

The repository does not contain a dedicated `main` package; it is primarily a library that can be imported by services. Tests serve as the main entry point for exercising functionality.

## Development Tips for Claude Code

- When adding new routes or middleware, follow the existing pattern of defining a handler function in `e2auth` (or the appropriate package) and registering it in `routers.go` or the relevant `*router.go` file.
- Use the existing `cfg *routerConfig` pattern for dependency injection (database, logger, emailer, etc.).
- For new HTTP client usage, start from the `Builder` in `e2http` and chain methods (e.g., `URL`, `AddHeader`, `Do`).
- Keep tests deterministic; many integration tests rely on the PostgreSQL dev instance defined in the README.
- If you need to embed static files, run the `generate/embedfs` command and import the generated package.
