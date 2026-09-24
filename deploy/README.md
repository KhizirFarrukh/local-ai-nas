# deploy

Deployment files.

| File | Purpose |
|---|---|
| `Dockerfile.dev` | Development image (S01.1-T06). Stages: `build` (Go toolchain, builds the binary), `dev` (runs `go run` on the mounted source), `runtime` (debian:trixie-slim with only the binary, non-root; the default target, scanned by Trivy in CI). |
| `compose.dev.yaml` | Development environment: `docker compose -f deploy/compose.dev.yaml up --build`. Serves `http://127.0.0.1:8080` to this computer only; data in a Docker volume. |
| `config.example.toml` | Every setting with its default and an explanation. A test keeps it in step with the code. |

Later: S11 adds the release image, the production Docker Compose setup, and a **separate setup script for each platform** (Linux x86-64, Raspberry Pi, Windows 11; FR-149). The scripts install the prerequisites listed in [`code-agent-docs/dependencies.md`](../code-agent-docs/dependencies.md) section 12.
