# CURRENT_STATE

**Last updated:** 2026-09-24 06:29 +0500 (session S005)
**Plan version:** 1.1.1 (`code-agent-docs/plan.md`), **Approved baseline** 1.0.0 (S005) + the setup-script requirement (1.1.0, S005 E015)
**Current phase:** Implementation: S01 (Basic NAS implementation) approved; starting S01.1

## Active stage and task
- **Active stage:** S01 (Basic NAS implementation), **In Progress** (`stages/S01-basic-nas.md`); substage S01.1 In Progress.
- **Active task:** none (S01.1-T05 done; S01.1-T07 next, then T08–T11, then T06)

## In progress (write-ahead)
- S01.1-T05: commit the records, push, and merge `feat/S01.1-T05-ci` into `develop`.

## Last completed
- `develop` verified complete (S005): merge `c535cda` brought `cf60f72` (plan 0.4.0, which PR #3 had put on `main` only) and the user's `bda8321` (prompts 3 and 4).
- Approval-stage decisions D-01–D-14 applied (S005 E007–E010). **Plan 1.0.0 baseline and S01 approved** (S005 E012).
- **S01.1-T05 done** (S005 E027): CI green on GitHub (run 35945085750, 7 jobs), red on a deliberate failing test (run 35945229679), green again after the revert (run 35945387694); arm64/amd64/windows artifacts produced.
- **S01.1-T04 done and merged** (`develop` dc9c0a0; S005 E025): `internal/testutil` + tests (unit, integration, fuzz seeds), `scripts/coverage.sh` (83.0%, gate works), `docs/testing.md`; 45 s of fuzzing found nothing.
- **S01.1-T03 done and merged** (`develop` f0cc48c; S005 E023): `.golangci.yml` (v2) + `scripts/install-golangci-lint.sh`; clean on the skeleton; depguard, errcheck, and gofmt violations fail (then removed).
- **S01.1-T02 done and merged** (`develop` b0cc2ec; S005 E021): `LICENSE` (AGPL v3), README License section, `docs/licensing.md`, `scripts/allowed-licenses.txt`, `scripts/check-licenses.sh`; the check passes, and a fixture with an Unlicense license fails it (then reverted).
- **S01.1-T01 done and merged** (`develop` 0100920; S005 E019): `go.mod` (go 1.27, toolchain go1.27.1; oapi-codegen, govulncheck, go-licenses as `tool` directives), skeleton packages, `api/openapi.yaml` stub, `.gitattributes`/`.editorconfig`/`.gitignore`. Build, vet, and test pass on Windows; Linux amd64/arm64 cross-build + vet pass; renormalize makes no changes.
- Plan 1.1.0: the user's requirement to record every dependency and build a setup script per platform (FR-149, NFR-032, S11.2, Q41; `dependencies.md` section 12; RULES 1.5.0) (S005 E015–E016).

## Next steps
1. **S01.1-T07** (configuration, `internal/config`): go-toml/v2 v2.4.3; keys `storage.*`, `server.*`, `log.*`, `uploads.*`; env `LOCALAINAS_*`; flags `--<section>-<key>`; precedence defaults < file < env < flags; validation errors name the key and its source. A draft and its tests exist in the session scratchpad (not in the repo). Branch `feat/S01.1-T07-config`.
2. Then T08 (logging), T09 (errors), T10 (SQLite), T11 (app skeleton), then **T06** (dev environment + image/Trivy CI job + Dependabot docker).
3. Then S01.2 onwards, in the execution order of the S01 document. Local Go: `C:\Program Files\Go\bin` (a new shell has it on PATH).
4. CI runs on every push (`.github/workflows/ci.yml`). Watch it anonymously with the public API (the repository is public; `gh` is not installed).
5. Every finished branch: merge it into `develop` myself and push (RULES User Preferences, S005).

## Blocked or waiting on user
- Q18 (library size) is needed before S01.7 only.

## Open questions (short list; full text in plan.md section 5)
- ★ **For S01:** Q18 library size (before S01.7)
- **Partly open:** Q5 native Windows/macOS installers (S11.2) · Q6 AI speed expectations · Q26 image formats (HEIC/RAW) · Q32 photos exposure over shares
- **Open for later stages:** Q10, Q11, Q13, Q14, Q15, Q19, Q27–Q31, Q33–Q36, Q39, Q40
- **Answered in S005:** Q1 (multi-platform), Q16 (closed), Q22 (AGPL-3.0), Q37 (none), Q38 (S01–S11)

## Pointers
- Latest session log: `code-agent-docs/logs/sessions/2026-09-24_S005.md`
- Active stage document: `code-agent-docs/stages/S01-basic-nas.md` (In Progress)
- Audit report: `code-agent-docs/audits/A001-2026-09-24-documentation-audit.md` (decisions received in section 9)
- ADRs: `code-agent-docs/decisions/ADR-0001` … `ADR-0020` (0019 Proposed; 0012 superseded in part by 0020; the rest Accepted, including 0003 since S005)
- Dependency register: `code-agent-docs/dependencies.md` (section 12: deployment prerequisites per platform, the input for the S11.2 setup scripts)
- Rules: `code-agent-docs/RULES.md` (v1.5.0: targeted plan reading; merge-into-develop; dependency record + per-platform setup scripts) · Prompts: `code-agent-docs/prompts/` (P002–P004)
