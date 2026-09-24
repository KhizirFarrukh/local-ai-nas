# deploy

Deployment files.

- S01.1-T06 adds the development environment: `Dockerfile.dev`, `compose.dev.yaml`, and `config.example.toml`.
- S11 adds the release image, the Docker Compose setup, and a **separate setup script for each platform** (Linux x86-64, Raspberry Pi, Windows 11; FR-149). The scripts install the prerequisites listed in [`code-agent-docs/dependencies.md`](../code-agent-docs/dependencies.md) section 12.
