# scripts

Developer helper scripts. On Windows, run the `.sh` files from Git Bash.

| Script | What it does |
|---|---|
| `install-golangci-lint.sh` | Installs the pinned golangci-lint into `./bin`. |
| `coverage.sh` | Runs the tests with coverage and fails below 80% for `internal/...` (`docs/testing.md`). |
| `check-licenses.sh` | Checks every dependency's license against `allowed-licenses.txt` (`docs/licensing.md`). |
| `check-api-docs-offline.sh [browser]` | Checks that the served API documentation renders with no internet access, in a headless Edge or Chrome (not part of CI). |
| `check-web-licenses.mjs` | Checks the licenses of the web interface's npm packages: shipped ones against `allowed-licenses.txt`, tools also against a short list of permissive tool licenses (`docs/licensing.md`). Run with `node`. |
| `perf-baseline.sh [size]` | Measures listing and transfer speed against NFR-003 (`docs/perf/`). |
| `demo.sh [URL]`, `demo.ps1 [-BaseUrl URL]` | The stage 1 API demo against a running server, every answer checked (`docs/api/usage.md`). `demo.ps1` runs in Windows PowerShell 5.1 and PowerShell 7: `powershell -ExecutionPolicy Bypass -File scripts\demo.ps1`. |

The per-platform deployment setup scripts are not here. They live in [`deploy/`](../deploy/) (S11.2).
