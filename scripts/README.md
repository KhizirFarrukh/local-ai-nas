# scripts

Developer helper scripts. On Windows, run the `.sh` files from Git Bash.

| Script | What it does |
|---|---|
| `install-golangci-lint.sh` | Installs the pinned golangci-lint into `./bin`. |
| `coverage.sh` | Runs the tests with coverage and fails below 80% for `internal/...` (`docs/testing.md`). |
| `check-licenses.sh` | Checks every dependency's license against `allowed-licenses.txt` (`docs/licensing.md`). |
| `check-api-docs-offline.sh [browser]` | Checks that the served API documentation renders with no internet access, in a headless Edge or Chrome (not part of CI). |
| `perf-baseline.sh [size]` | Measures listing and transfer speed against NFR-003 (`docs/perf/`). |

The S01.7 demo scripts (`demo.sh`, `demo.ps1`) follow in S01.7-T06.

The per-platform deployment setup scripts are not here. They live in [`deploy/`](../deploy/) (S11.2).
