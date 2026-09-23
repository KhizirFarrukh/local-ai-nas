# ADR-0009: Web UI: SvelteKit static SPA embedded in the Go binary

| Field | Value |
|---|---|
| Number | ADR-0009 |
| Status | **Accepted** |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

S02 builds the GUI (FR-078) that lets people use the NAS without touching the API. It must work from any device on the LAN, including phones (NFR-015), make no remote requests (I6), and use only the public API (ADR-0002). P003 resolves open question Q25: a web UI served by the NAS.

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **SvelteKit (adapter-static SPA) + TypeScript + Tailwind CSS** (chosen) | Small, fast bundles. Simple component model. A static build that is easy to embed | Smaller ecosystem than React |
| React + Vite (the S002 plan recommendation) | Largest ecosystem | Heavier, more boilerplate |
| Desktop app (Electron or Tauri) | Native feel | Not needed: a web UI works from any device on the network. Per-OS builds |

## Decision

- **SvelteKit** in static single-page mode with **`@sveltejs/adapter-static`** (SPA fallback page), **TypeScript**, and **Tailwind CSS**.
- **Verified versions (2026-09-24, all MIT unless noted):**
  - svelte **5.57.1**, @sveltejs/kit **2.70.3**, @sveltejs/adapter-static **3.0.10**, vite **8.3.0**.
  - tailwindcss **4.3.3** with @tailwindcss/vite **4.3.3**.
  - typescript **7.0.2** (Apache-2.0).
- **Serving:** the built UI (`web/build`) is **embedded into the Go binary with `go:embed`** (`web/embed.go`, ADR-0004) and served by the core at `/` with SPA fallback. The API stays under `/api/v1`, on the same origin, so no CORS is needed.
- **Development:** the Vite dev server proxies `/api` to the core server.
- **Uploads:** **Uppy** (ADR-0008).
- **API client:** generated from `api/openapi.yaml` (ADR-0002: openapi-typescript + openapi-fetch).
- **Package manager:** **pnpm 12.6.0** (MIT), with `pnpm-lock.yaml` committed.
- **Checks:** svelte-check, ESLint, Prettier. **Tests:** Vitest and Playwright (ADR-0005).

### Implementation details chosen by agent
| Detail | Choice | Reason |
|---|---|---|
| Virtualized lists and timeline grid | **`@tanstack/svelte-virtual` 3.13.39** (MIT, maintained, verified) as the virtualization engine, with a **custom date-grouped / justified row layout** on top for the photo timeline | A maintained engine for variable-height rows, and custom layout where photo timelines need it. **S02.3 includes a prototype task** to confirm Svelte 5 compatibility; if it fails, fall back to a custom windowing component (recorded in the stage document) |
| Package manager | pnpm (as P003 suggests) | Strict, fast, with lockfile support; Dependabot supports it |
| TypeScript version | Pin the version that svelte-check supports; S02.1 confirms TypeScript 7 compatibility | TypeScript 7 is a major compiler change, and tool support needs confirming (unverified today) |

> **Verification note (audit A001, 2026-09-24, F-021):** svelte-check 4.7.6 declares the peer range `typescript: ^5.0.0 || ^6.0.0`, so TypeScript 7 is **not** supported yet. Following the decision above ("pin the version that svelte-check supports"), the register pins **TypeScript 6.0.3**. S02.1 re-checks this and can move to 7.x once svelte-check supports it. The decision is unchanged.
| Fonts and icons | Bundled locally (no Google Fonts or CDN) | I6 |

## Consequences

- **Easier:** a single artifact (one binary includes the UI); same-origin security (no CORS); small bundles on phones.
- **Harder:** Svelte 5 ecosystem maturity for some components. Virtualization is prototyped in S02.3. PDF preview is decided in S02.6 between the browser's built-in viewer and the candidate **pdfjs-dist 6.3.289** (Apache-2.0, verified).
- **Required:** the S02.3 virtualization prototype task; a CSP compatible with the build (no inline scripts, which adapter-static supports, S03.5); an About page with third-party attributions (GeoNames, ADR-0013).

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.web_ui`). The virtualization approach is an "agent decides" item
> (2026-09-24, session S003). Versions and licenses verified in S003 log E005.

## Links

- **Related requirements:** FR-078–FR-083, FR-110, NFR-001, NFR-015, NFR-027
- **Related ADRs:** ADR-0002, ADR-0004, ADR-0005, ADR-0008, ADR-0010
- **Related stages:** S02 onward
- **Plan version:** 0.3.0
