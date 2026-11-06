# Go Architectural & Engineering Learnings from rock-sql

This document records architectural patterns, standard library nuances, cross-platform compilation insights, and best practices learned and solidified while reviewing and operating the [`rock-sql`](../main.go) codebase.

---

## 1. Database & Concurrency (`database/sql` + pure Go drivers)

### Pure Go vs. cgo (`github.com/lib/pq`)
- The driver `github.com/lib/pq` is implemented entirely in pure Go without relying on the C PostgreSQL client library (`libpq`).
- **Why this matters for portability**: Pure Go drivers require no C toolchain (`CGO_ENABLED=0`). This enables zero-friction cross-compilation across operating systems (e.g., building an OpenBSD `amd64` static binary directly from Linux or macOS without setting up an OpenBSD cross-compiler).

### Connection Pools as Dynamic Multi-DB Registries
In multi-database systems where databases are discovered or addressed dynamically (like `rock-sql`'s catalog), managing pools requires concurrency safety:
```go
var (
    pools   = map[string]*sql.DB{}
    poolsMu sync.Mutex
)
```
- In Go, `*sql.DB` represents a connection pool, not a single physical connection.
- Calling `sql.Open` does not establish connections immediately; it validates the driver name and DSN syntax and initializes the pool lazily upon the first query (`Ping`, `Query`, or `Exec`).
- Safeguard `SetMaxOpenConns(4)` prevents connection exhaustion across multiple database connections when running under resource-constrained environments (like a lightweight OpenBSD VM or home server).

### Read-Only Transactions with Context Deadlines
Executing arbitrary user queries safely requires defense-in-depth:
```go
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()

tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
if err != nil {
    return Result{Err: err.Error()}
}
defer tx.Rollback() // Always safe to defer Rollback for read-only tx

_, _ = tx.ExecContext(ctx, "SET LOCAL statement_timeout = '15s'")
```
Key learnings:
1. **`sql.TxOptions{ReadOnly: true}`**: Informs PostgreSQL to execute the transaction in `READ ONLY` mode at the database level, preventing unintended `UPDATE`, `INSERT`, `DROP`, or `DELETE` executions.
2. **`context.WithTimeout`**: Guards the Go process from blocked network socket reads or connection hangs.
3. **`SET LOCAL statement_timeout`**: Informs the Postgres query engine itself to abort execution if query planning/scanning exceeds the limit, protecting server CPU and disk I/O.
4. **Deferred `tx.Rollback()`**: Even for pure reads, deferred rollback ensures connection locks are cleanly released back to the pool without needing explicit commit logic.

### Dynamic Row Scanning for Unknown Schemas
When querying arbitrary SQL tables whose column types and counts are unknown at compile time:
```go
cols, _ := rows.Columns()
vals := make([]sql.NullString, len(cols))
ptrs := make([]any, len(cols))
for i := range vals {
    ptrs[i] = &vals[i]
}
rows.Scan(ptrs...)
```
Key insights:
- `rows.Scan(...)` accepts variadic `...any`, but requires pointers to addressable memory.
- `sql.NullString` gracefully handles `NULL` database values without panicking.
- By binding a pre-allocated slice of interface pointers (`[]any`) to the corresponding `[]sql.NullString`, Go can dynamically scan any database record with zero reflection overhead.

---

## 2. Templating & Frontend Hydration (`html/template`)

### Context-Aware Auto-Escaping vs. `template.JS`
Go's `html/template` package parses HTML, CSS, JavaScript, and URI attributes, applying context-aware escaping based on location:
- In HTML contexts, text is automatically escaped to prevent XSS (`<` becomes `&lt;`).
- When embedding JSON data directly into an inline `<script>` tag:
  ```go
  template.JS(js.String())
  ```
- **Nuance & Security**: `template.JS` tells the Go template engine to trust the string as safe JavaScript code without escaping. Because `rock-sql` generates the JSON payloads using PostgreSQL's native `row_to_json(t)::text`, the payload is formatted JSON. However, developers must ensure input data does not contain literal `</script>` sequences which browsers could interpret as closing the script block early.

---

## 3. Cross-Platform Compilation for OpenBSD

### Zero-Dependency Static Binaries
Go's toolchain makes cross-compiling for OpenBSD trivial when CGO is disabled:
```bash
GOOS=openbsd GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/openbsd_amd64/rock-sql .
```
- `-trimpath`: Strips absolute file system paths from the compiled binary, ensuring deterministic, reproducible builds.
- `-ldflags="-s -w"`:
  - `-s`: Strips symbol table information.
  - `-w`: Strips DWARF debug information.
  - Reduces binary size by ~35-45% (down to ~9.8 MB) with zero runtime performance penalty.

### OpenBSD Security Semantics (`pledge` and `unveil`)
OpenBSD features kernel-level sandboxing primitives:
- `pledge(2)`: Restricts system call capabilities (e.g. `stdio rpath inet`).
- `unveil(2)`: Restricts filesystem visibility to specific paths.
- While `rock-sql` runs as a standalone network daemon, OpenBSD's `golang.org/x/sys/unix` package allows native integration with `unix.Pledge` and `unix.Unveil` if kernel sandboxing is desired in future releases.

---

## 4. Configuration & Server Idioms

### Environment Variable Defaults
For services running across heterogeneous environments (OpenBSD servers, local containers, macOS development laptops):
```go
func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```
Using standard PostgreSQL environment variable conventions (`PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGSSLMODE`) allows the application to:
1. Work seamlessly out of the box with zero configuration when standard local defaults exist.
2. Integrate automatically with standard toolings, Docker Compose, systemd, or OpenBSD `/etc/rc.conf.local`.

---

## 5. Rich Interactive Multimodal UX in Zero-Dependency Go Services

Go web apps often serve minimalist, text-heavy UIs. However, standard Go template rendering (`html/template`) can deliver rich, zero-dependency interactive experiences by combining:
1. **Web Audio API (`window.AudioContext`)**: Synthesizing procedural UI sound effects (beeps, scans, frequencies) entirely client-side without needing external `.mp3` or `.wav` asset pipelines.
2. **Web Speech API (`speechSynthesis`)**: Dynamic multi-voice text-to-speech with pitch and rate modulation, allowing voice persona switching (robot, alien, cadet) across diverse operating systems and browsers with zero backend audio compute.
3. **HTML5 2D Canvas**: 3D orthographic and spherical geometry projections (like Earth continents, day/night shadows, orbiting moons) executed with pure JavaScript math, completely removing heavy 3D engine dependencies like Three.js.
