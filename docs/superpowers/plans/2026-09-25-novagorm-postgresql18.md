# novagorm PostgreSQL 18 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add production-ready PostgreSQL support to `novagorm`, prove it against PostgreSQL 18, publish the capability, and prepare a reviewable release.

**Architecture:** Keep `starter/gorm/novagorm` as the single database lifecycle boundary. Add a PostgreSQL configuration branch beside MySQL using GORM's official PostgreSQL dialector, apply the existing pool and Model First guard behavior, and expose one `OpenPostgresFromSQLDB` bridge. A build-tagged integration test runs against a real PostgreSQL 18 server; a repository-level test validates the machine-readable starter index.

**Tech Stack:** Go 1.24.5, GORM 1.31.x, `gorm.io/driver/postgres` v1.6.0, pgx v5, PostgreSQL 18, Docker, YAML.

**Spec:** `docs/superpowers/specs/2026-09-06-nova-lightweight-monorepo-postgresql18-design.md`

## Global Constraints

- Nova remains one repository and one Go module; do not add a driver submodule or runtime plugin loader.
- The only PostgreSQL driver ID is `postgres`; reject `postgresql` and unknown IDs.
- PostgreSQL configuration exists only below `gorm.postgres` or `gorm.<name>.postgres`.
- Do not include DSNs or passwords in logs or returned errors.
- Pool values less than or equal to zero leave `database/sql` defaults unchanged.
- Preserve MySQL, `Register`, named-instance, reload, close, and Model First behavior.
- PostgreSQL 18 is the real-server baseline; runtime code does not hard-code a server version.
- Pin `gorm.io/driver/postgres` to v1.6.0 because v1.6.1+ declares Go 1.25 while Nova declares Go 1.24.5.
- Do not add application schemas, business queries, foreign-key generation, or MySQL-to-PostgreSQL migration tooling.

## Review Focus

- Missing PostgreSQL DSN: return an instance-aware configuration error without leaking configuration; Task 1 tests it.
- Unsupported driver aliases: fail closed instead of falling back to MySQL; Task 1 tests it.
- Mixed named MySQL/PostgreSQL instances and `default`: preserve deterministic selection; Task 1 tests it.
- PostgreSQL migrator extension: preserve `BuildIndexOptions` through the guard; Task 2 tests it.
- PostgreSQL lifecycle: closing a Nova instance releases its underlying pool; Tasks 2 and 3 test it.

---

### Task 1: PostgreSQL configuration and driver dispatch

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`
- Modify: `starter/gorm/novagorm/gorm.go`
- Modify: `starter/gorm/novagorm/gorm_dependency_test.go`

**Interfaces:**
- Consumes: `buildDefinitions(map[string]any)`, `parseGormConfig(map[string]any)`, `newConfiguredConnection(gormConfig)`.
- Produces: `postgresConfig`, `gormConfig.Postgres`, `newPostgresConnection(gormConfig) (*gorm.DB, error)`.

- [ ] **Step 1: Pin the compatible official dialector**

Run:

```bash
go get gorm.io/driver/postgres@v1.6.0
```

Expected: `go.mod` contains `gorm.io/driver/postgres v1.6.0` and still declares `go 1.24.5`.

- [ ] **Step 2: Write failing configuration tests**

Add tests for the PostgreSQL node and pool values:

```go
func TestParseGormConfigKeepsPostgresConfigUnderDriver(t *testing.T) {
	got := parseGormConfig(map[string]any{
		"driver": "postgres",
		"postgres": map[string]any{
			"dsn": "host=127.0.0.1 user=nova password=secret dbname=nova port=5432 sslmode=disable",
			"max_open": 20, "max_idle": 10,
			"conn_max_lifetime": 1800, "conn_max_idle_time": 300,
			"prefer_simple_protocol": true,
		},
	})
	if got.Driver != "postgres" || got.Postgres.DSN == "" || !got.Postgres.PreferSimpleProtocol {
		t.Fatalf("postgres config = %#v", got)
	}
	if got.Postgres.MaxOpen != 20 || got.Postgres.MaxIdle != 10 || got.Postgres.ConnMaxLifetime != 1800 || got.Postgres.ConnMaxIdleTime != 300 {
		t.Fatalf("postgres pool config = %#v", got.Postgres)
	}
}
```

Add `TestBuildDefinitionsSupportsNamedMySQLAndPostgres` with `default: main`, a PostgreSQL `main`, and a MySQL `legacy`; assert both builders exist and `main` is selected. Add `TestConfiguredConnectionRejectsPostgresqlAlias` and `TestPostgresMissingDSNDoesNotLeakConfiguration`; the latter must invoke the named builder, assert the error names both the instance and `postgres.dsn`, and excludes a sentinel password. Add `TestApplyPostgresPoolConfig` using `sqlmock`; assert `sqlDB.Stats().MaxOpenConnections` becomes the configured positive value and non-positive values preserve the default.

- [ ] **Step 3: Verify RED**

Run:

```bash
go test ./starter/gorm/novagorm -run 'Test(ParseGormConfigKeepsPostgres|BuildDefinitionsSupportsNamedMySQLAndPostgres|ConfiguredConnectionRejectsPostgresqlAlias|PostgresMissingDSN)' -count=1
```

Expected: compile/test failure because PostgreSQL configuration and dispatch do not exist.

- [ ] **Step 4: Implement minimal configuration and dispatch**

Add:

```go
type postgresConfig struct {
	DSN                  string
	PreferSimpleProtocol bool
	MaxOpen              int
	MaxIdle              int
	ConnMaxLifetime      int
	ConnMaxIdleTime      int
}
```

Add `Postgres postgresConfig` to `gormConfig`; parse the `postgres` node with the same numeric aliases accepted for MySQL; reserve `postgres`; dispatch only `case "postgres"`. Honor `gorm.default` when it names an existing configured instance; otherwise preserve the current one-instance/no-selection behavior.

Open with:

```go
gormpostgres.New(gormpostgres.Config{
	DSN: cfg.Postgres.DSN,
	PreferSimpleProtocol: cfg.Postgres.PreferSimpleProtocol,
})
```

Reject an empty DSN before `gorm.Open`. Wrap builder failures with instance and driver context, but never include the DSN. Extract a driver-neutral pool helper or add an equivalent PostgreSQL helper that applies only positive values.

- [ ] **Step 5: Verify GREEN and commit**

Run `go test ./starter/gorm/novagorm -count=1`, then:

```bash
git add go.mod go.sum starter/gorm/novagorm/gorm.go starter/gorm/novagorm/gorm_dependency_test.go
git commit -m "feat(novagorm): configure PostgreSQL connections"
```

### Task 2: SQL bridge and Model First guard compatibility

**Files:**
- Modify: `starter/gorm/novagorm/gorm.go`
- Modify: `starter/gorm/novagorm/gorm_magic_guard_test.go`
- Modify: `starter/gorm/novagorm/gorm_close_test.go`

**Interfaces:**
- Consumes: `installAutoMigrateGuard(*gorm.DB)` and `autoMigrateGuard.BuildIndexOptions`.
- Produces: `OpenPostgresFromSQLDB(sqlDB *sql.DB) (*gorm.DB, error)`.

- [ ] **Step 1: Write failing bridge and guard tests**

Add nil-input coverage and:

```go
func TestOpenPostgresFromSQLDBInstallsAutoMigrateGuard(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := OpenPostgresFromSQLDB(sqlDB)
	if err != nil { t.Fatalf("open postgres bridge: %v", err) }
	if err := db.AutoMigrate(&implicitTableNameModel{}); !errors.Is(err, ErrORMMagic) {
		t.Fatalf("AutoMigrate error = %v, want ErrORMMagic", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}
```

Add a recording migrator implementing `migrator.BuildIndexOptionsInterface`; assert the guard returns its sentinel options. Add a PostgreSQL bridge close test using `sqlmock.ExpectClose` plus `CloseAll`.

- [ ] **Step 2: Verify RED**

Run:

```bash
go test ./starter/gorm/novagorm -run 'TestOpenPostgresFromSQLDB|TestAutoMigrateGuardPreservesBuildIndexOptions|TestPostgresBridgeClose' -count=1
```

Expected: compile failure because `OpenPostgresFromSQLDB` does not exist.

- [ ] **Step 3: Implement the guarded bridge**

```go
func OpenPostgresFromSQLDB(sqlDB *sql.DB) (*gorm.DB, error) {
	if sqlDB == nil { return nil, fmt.Errorf("sql db is nil") }
	db, err := gorm.Open(gormpostgres.New(gormpostgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil { return nil, fmt.Errorf("failed to open postgres gorm from sql db: %w", err) }
	installAutoMigrateGuard(db)
	return db, nil
}
```

Keep the current `BuildIndexOptions` forwarding unless the failing test exposes a v1.6.0 interface mismatch; forward the exact extension rather than bypassing the guard.

- [ ] **Step 4: Verify GREEN and commit**

Run `go test ./starter/gorm/novagorm -count=1`, then:

```bash
git add starter/gorm/novagorm/gorm.go starter/gorm/novagorm/gorm_magic_guard_test.go starter/gorm/novagorm/gorm_close_test.go
git commit -m "feat(novagorm): add guarded PostgreSQL SQL bridge"
```

### Task 3: PostgreSQL 18 real-server integration

**Files:**
- Create: `starter/gorm/novagorm/postgres_integration_test.go`

**Interfaces:**
- Consumes: configured PostgreSQL opening, guarded `AutoMigrate`, transactions, and lifecycle.
- Produces: build-tagged `postgres_integration` verification driven by `POSTGRES_TEST_DSN`.

- [ ] **Step 1: Write the build-tagged integration test**

Create `//go:build postgres_integration` coverage that requires `POSTGRES_TEST_DSN`, connects through Nova, and asserts `SHOW server_version_num` starts with `180`.

Use dedicated `nova_pg18_*` test table names, drop them before the test, and defer dropping them afterward; the integration database must be disposable and dedicated to this test. Define explicit Models with `TableName()`, `column` tags, primary keys, a unique index, and `type:jsonb;not null`; do not declare associations or foreign-key constraints. Verify guarded `AutoMigrate`, table/column/index creation, JSONB round-trip, duplicate unique-key failure, transaction rollback, `CloseAll`, and failure of a subsequent pool ping.

- [ ] **Step 2: Start PostgreSQL 18**

```bash
docker run --name nova-postgres18-test --rm -d \
  -e POSTGRES_USER=nova -e POSTGRES_PASSWORD=nova -e POSTGRES_DB=nova \
  -p 55432:5432 postgres:18
docker exec nova-postgres18-test pg_isready -U nova -d nova
```

Expected: `accepting connections`.

- [ ] **Step 3: Verify RED on the real server**

```bash
POSTGRES_TEST_DSN='postgres://nova:nova@127.0.0.1:55432/nova?sslmode=disable' \
go test -tags=postgres_integration ./starter/gorm/novagorm -run TestPostgres18Integration -count=1 -v
```

Expected: fail on the first missing or incorrect PostgreSQL behavior after establishing a real PostgreSQL 18 connection.

- [ ] **Step 4: Make only integration-discovered corrections**

Correct the owning dialector, guard-forwarding, Model-tag, or lifecycle code. Do not add business schema APIs, associations, foreign-key generation, or custom migration tags.

- [ ] **Step 5: Verify GREEN, stop the database, and commit**

Re-run the integration command and `go test ./starter/gorm/novagorm -count=1`. Then:

```bash
docker stop nova-postgres18-test
git add starter/gorm/novagorm/postgres_integration_test.go starter/gorm/novagorm/gorm.go
git commit -m "test(novagorm): verify PostgreSQL 18 behavior"
```

### Task 4: Starter index and documentation

**Files:**
- Create: `starter-index.yaml`
- Create: `starter_index_test.go`
- Modify: `docs/starters/novagorm.md`
- Modify: `README.md`
- Modify: `docs/starter_conventions.md`
- Modify: `docs/starter_composition_matrix.md`
- Modify: `docs/ai_coding_guide.md`

**Interfaces:**
- Consumes: shipped `postgres` configuration and public bridge.
- Produces: a machine-readable starter inventory and consistent user documentation.

- [ ] **Step 1: Write the failing index-integrity test**

At repository root, parse `starter-index.yaml` with `gopkg.in/yaml.v3`. Assert schema version 1; unique IDs/package paths; existing package directories/docs; non-empty required fields; every public README starter represented; and stable `mysql` plus `postgres` under `novagorm`, with `verified_server: PostgreSQL 18`.

- [ ] **Step 2: Verify RED**

Run `go test . -run TestStarterIndex -count=1`. Expected: FAIL because `starter-index.yaml` does not exist.

- [ ] **Step 3: Create the complete index**

Add entries for `novaconfig`, `novagin`, `novaredis`, `novaoss`, `novaqwen`, `novawebsocket`, and `novagorm`. Declare:

```yaml
drivers:
  - id: mysql
    status: stable
  - id: postgres
    status: stable
    verified_server: PostgreSQL 18
```

- [ ] **Step 4: Update all affected documentation**

Document single/named PostgreSQL configuration, `prefer_simple_protocol`, all pool fields, `OpenPostgresFromSQLDB`, the `postgres`-only driver ID, PostgreSQL 18 verification, DSN secret handling, and the absence of association/foreign-key generation and data migration. Replace MySQL-only headings where the contract now covers both drivers.

- [ ] **Step 5: Verify GREEN and commit**

```bash
go test . -run TestStarterIndex -count=1
go test ./... -count=1
go vet ./...
git add starter-index.yaml starter_index_test.go README.md docs/starters/novagorm.md docs/starter_conventions.md docs/starter_composition_matrix.md docs/ai_coding_guide.md
git commit -m "docs: publish PostgreSQL starter support"
```

### Task 5: Release verification and上线 handoff

**Files:**
- Modify only if verification exposes a defect in its owning file.

**Interfaces:**
- Consumes: all prior task outputs.
- Produces: fresh release evidence, a pushed review branch, and only the release action authorized by the user.

- [ ] **Step 1: Inspect the complete change**

```bash
git diff main...HEAD --check
git diff --stat main...HEAD
git status --short
```

Map spec sections 7 and 9 to a test, document, or diff hunk. Confirm no DSN, password, generated database data, or unrelated refactor is present.

- [ ] **Step 2: Run fresh full verification**

With a fresh `postgres:18` container:

```bash
go test ./... -count=1
go vet ./...
POSTGRES_TEST_DSN='postgres://nova:nova@127.0.0.1:55432/nova?sslmode=disable' \
go test -tags=postgres_integration ./starter/gorm/novagorm -run TestPostgres18Integration -count=1 -v
```

All commands must exit 0 and the integration log must prove major version 18.

- [ ] **Step 3: Push the verified branch**

```bash
git push -u origin codex/nova-postgresql18-driver-design
```

- [ ] **Step 4: Create and attach the pull request**

Target `main`; title `feat(novagorm): support PostgreSQL 18`; include exact unit, vet, and PostgreSQL 18 evidence; attach the PR to this Codex task.

- [ ] **Step 5: Perform only the explicitly selected release action**

Treat a reviewed pull request as the default上线 boundary. Merge only when explicitly requested and required checks pass. Create a version tag/release only when explicitly requested; first inspect existing tag conventions, agree the next semantic version, tag the merged commit, push it, and verify the release/module proxy state.
