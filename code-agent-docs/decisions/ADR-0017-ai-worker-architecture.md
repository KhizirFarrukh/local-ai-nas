# ADR-0017: AI worker architecture: Python + ONNX Runtime in a separate optional container; core as single writer

| Field | Value |
|---|---|
| Number | ADR-0017 |
| Status | **Accepted** |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

S17 adds optional, fully local AI (I6, I7, I8): classification and face grouping. Results go into sidecars and the index, and search never runs a model (I4). Constraints:
- The NAS must be fully functional with AI off.
- CPU by default on modest hardware (NFR-004).
- No internet at runtime.
- New invariant **I9** (P003): only the core writes sidecars and the index.

## Options considered

| Option | Pros | Cons |
|---|---|---|
| **Separate Python worker container with ONNX Runtime** (chosen) | The AI ecosystem lives in Python. Isolation: a crash or memory spike cannot hurt the core. Truly optional (Compose profile) | A second language and a process boundary |
| AI inside the Go core (onnxruntime-go via cgo) | One process | cgo; heavier core; violates "optional" isolation |
| PyTorch-based worker | All models available | Multi-GB images; slower on CPU |

## Decision

- **Language/runtime:** **Python**, current stable (**3.14.7** at verification, EOL 2030-10-31), with **ONNX Runtime** (**onnxruntime 1.30.0**, MIT, verified), in a **separate, optional container** enabled via a Compose profile (`docker compose --profile ai up`, ADR-0006).
- **Acceleration:** CPU by default. Optional GPU acceleration through ONNX Runtime execution providers (e.g. CUDA, OpenVINO). **Which to support is evaluated in S17.1 (deferred).**
- **Tooling:** dependencies managed with **uv 0.12.18** (MIT OR Apache-2.0) and a lockfile (`uv.lock`). Checks: **Ruff 0.16.8** and **pytest 9.1.1** (ADR-0005).
- **Supporting libraries (candidates, confirmed in S17.1):** **numpy 2.5.3** (BSD-3-Clause and other permissive licenses) and **pillow 12.3.0** (MIT-CMU) for image decoding. All verified 2026-09-24.
- **Job flow:** the worker **pulls jobs from a local-only internal API** on the core server, authenticated with a **token generated at setup**, using the job lease protocol (ADR-0011).
  - The internal API is bound to the internal container network or localhost, never exposed on the LAN.
- **Storage access:** media is mounted **read-only** into the AI container. The worker **cannot write to storage**.
- **Results:** returned as **JSON**. The core **validates** them (schema, ranges, model@version) and is the **only writer** of sidecars and the index (**invariant I9**).
- **Offline:** no internet at runtime. Models are either bundled in the AI image or downloaded once at opt-in with explicit user consent, and are **checksum-verified**. The container runs with no outbound network in the Compose file (e.g. an internal-only network).

## Consequences

- **Easier:** AI failures are contained; the core stays pure Go; the AI can be removed entirely.
- **Harder:** a versioned, contract-tested JSON job protocol between Go and Python; model management and consent flow.
- **Required:**
  - The contract schema (S17.1); contract tests with a fake worker (plan 12).
  - The token rotation procedure; the Compose network isolation.
  - **Unverified:** that ONNX Runtime publishes wheels for Python 3.14. Confirm in S17.1, and pin the newest Python minor that ONNX Runtime supports.

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.ai_worker`)
> (2026-09-24, session S003). Versions and licenses verified in S003 log E005.

## Links

- **Related requirements:** FR-031, FR-032, FR-035, FR-036, FR-136, NFR-002, NFR-004, NFR-005, NFR-018
- **Related ADRs:** ADR-0006, ADR-0011, ADR-0018, invariant I9 (RULES.md)
- **Related stages:** S17
- **Plan version:** 0.3.0

## Amendment 1 (2026-09-30, plan 1.9.0): hardware-aware execution, query encoder

- **Execution providers** (FR-361; verified 2026-09-30 on onnxruntime.ai and PyPI, all MIT): CPU (default; ARM64 for the Raspberry Pi); CUDA and TensorRT (`onnxruntime-gpu` 1.30.0); OpenVINO for Intel CPUs, GPUs, and NPUs (`onnxruntime-openvino` 1.24.1); AMD through MIGraphX (`onnxruntime-migraphx` 1.27.1; the ROCm provider is marked deprecated); DirectML on Windows (`onnxruntime-directml` 1.24.4; "in sustained engineering", new development in Windows ML); CoreML only if macOS is supported. The worker detects the hardware, picks the best available provider, benchmarks on first start, falls back to CPU, and records the provider with each result.
- **Images per accelerator** as compose profiles (CPU, NVIDIA, Intel, AMD), each with its ONNX Runtime build (S14.1).
- **FastEmbed** (0.8.1, Apache-2.0) may be evaluated only as a convenience layer running strictly from local files (verified: `specific_model_path` returns the local folder without any download; `local_files_only` or `HF_HUB_OFFLINE` blocks downloads), so I6 holds.
- **Query-text encoder** (the user's decision D3, I4 reworded): the worker serves one more internal endpoint that embeds typed search text, with a strict time limit and a small cache (ADR-0043).
