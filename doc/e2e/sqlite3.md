# atago Behavior Specs
## Summary
2 suites · 10 scenarios
## Contents
- [sqlite3 + changes (workdir-delta of a database write)](#sqlite3--changes-workdir-delta-of-a-database-write) — 2 scenarios
  - [default rollback-journal mode creates exactly the db file](#scenario-default-rollback-journal-mode-creates-exactly-the-db-file)
  - [WAL mode leaves no -wal/-shm behind after a clean close](#scenario-wal-mode-leaves-no--wal-shm-behind-after-a-clean-close)
- [sqlite3 (third-party CLI, no build required)](#sqlite3-third-party-cli-no-build-required) — 8 scenarios
  - [one-shot SQL creates, inserts, and counts in a single invocation](#scenario-one-shot-sql-creates-inserts-and-counts-in-a-single-invocation)
  - [the database file is durable across invocations](#scenario-the-database-file-is-durable-across-invocations)
  - [-json output mode is valid JSON with typed values](#scenario--json-output-mode-is-valid-json-with-typed-values)
  - [-csv output mode emits plain rows](#scenario--csv-output-mode-emits-plain-rows)
  - [.dump emits SQL that rebuilds an identical database](#scenario-dump-emits-sql-that-rebuilds-an-identical-database)
  - [.import loads a CSV fixture into a table](#scenario-import-loads-a-csv-fixture-into-a-table)
  - [bad SQL exits 1 with the error position on stderr](#scenario-bad-sql-exits-1-with-the-error-position-on-stderr)
  - [querying a missing table names it in the diagnostics](#scenario-querying-a-missing-table-names-it-in-the-diagnostics)
## sqlite3 + changes (workdir-delta of a database write)
Source: `test/e2e/thirdparty/sqlite3/changes.atago.yaml`
### Scenario: default rollback-journal mode creates exactly the db file
#### When
```shell
sqlite3 t.db "create table x(a); insert into x values(1)"
```
#### Then
- exit code is `0`
- the step changed exactly created `t.db`, modified nothing, deleted nothing
### Scenario: WAL mode leaves no -wal/-shm behind after a clean close
#### When
```shell
sqlite3 t.db "PRAGMA journal_mode=WAL; create table x(a); insert into x values(1)"
```
#### Then
- exit code is `0`
- the step changed exactly created `t.db`, modified nothing, deleted nothing
## sqlite3 (third-party CLI, no build required)
Source: `test/e2e/thirdparty/sqlite3/sqlite3.atago.yaml`
### Scenario: one-shot SQL creates, inserts, and counts in a single invocation
#### When
```shell
sqlite3 app.db "CREATE TABLE u(id INTEGER, name TEXT); INSERT INTO u VALUES(1,'alice'),(2,'bob'); SELECT count(*) FROM u;"
```
#### Then
- exit code is `0`
- stdout equals an exact value
- file `app.db` exists
#### Generated artifacts
- `app.db`
### Scenario: the database file is durable across invocations
#### When
```shell
sqlite3 app.db "CREATE TABLE kv(k TEXT, v TEXT); INSERT INTO kv VALUES('answer','42');"
sqlite3 app.db "SELECT v FROM kv WHERE k='answer';"
```
#### Then
- after `sqlite3 app.db "SELECT v FROM kv WHERE k='answer';"`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: -json output mode is valid JSON with typed values
#### When
```shell
sqlite3 app.db "CREATE TABLE u(id INTEGER, name TEXT); INSERT INTO u VALUES(1,'alice');"
sqlite3 -json app.db "SELECT * FROM u;"
```
#### Then
- after `sqlite3 -json app.db "SELECT * FROM u;"`:
  - exit code is `0`
  - stdout at `$[0].id` equals `1`
  - stdout at `$[0].name` equals `alice`
### Scenario: -csv output mode emits plain rows
#### When
```shell
sqlite3 app.db "CREATE TABLE u(id INTEGER, name TEXT); INSERT INTO u VALUES(1,'alice'),(2,'bob');"
sqlite3 -csv app.db "SELECT * FROM u ORDER BY id;"
```
#### Then
- after `sqlite3 -csv app.db "SELECT * FROM u ORDER BY id;"`:
  - exit code is `0`
  - stdout equals an exact value
  - stdout equals an exact value
### Scenario: .dump emits SQL that rebuilds an identical database
#### When
```shell
sqlite3 app.db "CREATE TABLE u(id INTEGER, name TEXT); INSERT INTO u VALUES(1,'alice');"
sqlite3 app.db .dump
sqlite3 copy.db ".read dump.sql"
sqlite3 copy.db "SELECT name FROM u WHERE id=1;"
```
#### Then
- after `sqlite3 app.db .dump`:
  - exit code is `0`
  - stdout contains `CREATE TABLE u`, `INSERT INTO u`
- after `sqlite3 copy.db "SELECT name FROM u WHERE id=1;"`:
  - exit code is `0`
  - stdout equals an exact value
#### Generated artifacts
- `dump.sql`
### Scenario: .import loads a CSV fixture into a table
#### Given
- Fixture file `people.csv` is created.
#### Inputs
_Fixture `people.csv`:_
```text
id,name
1,carol
2,dave
```
#### When
```shell
sqlite3 -csv app.db ".import people.csv people"
sqlite3 app.db "SELECT count(*) FROM people;"
```
#### Then
- after `sqlite3 -csv app.db ".import people.csv people"`:
  - exit code is `0`
- after `sqlite3 app.db "SELECT count(*) FROM people;"`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: bad SQL exits 1 with the error position on stderr
#### When
```shell
sqlite3 app.db "SELEC broken;"
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `syntax error`
### Scenario: querying a missing table names it in the diagnostics
#### When
```shell
sqlite3 app.db "SELECT * FROM no_such_table;"
```
#### Then
- exit code is `1`
- stderr contains `no_such_table`
