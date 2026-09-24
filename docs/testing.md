# Testing

How tests are written and run (ADR-0005, S01.1-T04). On Windows, run the `scripts/*.sh` commands from Git Bash.

## Commands

| What | Command |
|---|---|
| All tests | `go test ./...` |
| With the race detector (Linux/macOS; needs cgo) | `go test -race ./...` |
| Coverage gate (80% of `internal/...`) | `scripts/coverage.sh` (extra flags are passed to `go test`, e.g. `scripts/coverage.sh -race`) |
| One fuzz target, searching for new inputs | `go test -run '^$' -fuzz '^FuzzWriteFiles$' -fuzztime 30s ./internal/testutil` |
| Lint and format | `./bin/golangci-lint run ./...` and `./bin/golangci-lint fmt --diff` (install: `scripts/install-golangci-lint.sh`) |
| Dependency licenses | `scripts/check-licenses.sh` |

CI runs the tests on Linux (with `-race`) and Windows, plus the coverage gate (S01.1-T05).

## Conventions

- **Standard library `testing`** plus [`go-cmp`](https://github.com/google/go-cmp) for comparisons: `if diff := cmp.Diff(want, got); diff != "" { t.Errorf("... (-want +got):\n%s", diff) }`. No assertion framework.
- **Table-driven tests** with `t.Run` subtests for rule sets such as validation, precedence, and error mapping.
- **Every task ships its tests** (RULES R6). A bug fix starts with a failing test that reproduces it.
- **No shared state:** each test makes its own storage root and server, so tests can run in parallel.
- **Loopback only:** test servers listen on `127.0.0.1:0` (NFR-020).
- **Both platforms:** anything touching paths or files must pass on Linux and Windows. Windows-only cases go behind `filepath.Separator == '\\'` or `runtime.GOOS`.

## Helpers (`internal/testutil`)

Imported only from `_test.go` files.

| Helper | Use |
|---|---|
| `StorageRoot(t)` | A new, empty storage-root directory (symlinks resolved), removed after the test. S01.2 extends it with the storage layout. |
| `NewServer(t, handler)` | An `httptest.Server` on `127.0.0.1` with a free port, closed after the test. |
| `WriteFiles(dir, map[name]content)` | Creates a fixture tree. Names are slash-separated. Writes go through `os.Root`, so a name can never write outside `dir`. |
| `ReadFiles(dir)` | Reads a tree back as `map[name]content` for comparison with `cmp.Diff`. |

## Integration tests

Integration tests use a real temporary storage root and, from S01.3 on, the real HTTP stack through `NewServer`. They live next to the code (`*_test.go`) and run with `go test ./...`. No external services are needed.

## Fuzz tests

- A fuzz target is a `FuzzXxx(f *testing.F)` function. Its `f.Add(...)` seeds run with every `go test`, so the seed corpus is always part of the normal test run.
- Searching for new failing inputs is manual or scheduled (`-fuzz`, as in the table above). When the fuzzer finds a failure it writes the input to `testdata/fuzz/FuzzXxx/` in the package. Commit that file together with the fix, so the case stays a regression test.
- `FuzzWriteFiles` is the sample target. The important targets are the path resolver and the name rules (S01.6), whose attack corpus also seeds the fuzzer.

## Coverage

`scripts/coverage.sh` writes `coverage.out` (git-ignored), prints the per-function report, and fails when the total for `internal/...` is below 80%. Generated code (`internal/api/gen`, from `api/openapi.yaml`) is left out of the figure; its behavior is tested through `internal/api`. Set `COVERAGE_MIN` to try another threshold. `cmd/` is covered by the smoke test (S01.1-T11), not by the threshold.

## Test data

Files under `testdata/` are listed in [`testdata/SOURCES.md`](../testdata/SOURCES.md) with their origin and license. Generate fixtures in test code when possible.
