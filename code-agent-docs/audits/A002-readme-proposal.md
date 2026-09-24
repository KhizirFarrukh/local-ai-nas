# A002: README change proposal

| Field | Value |
|---|---|
| Audit | A002 (R12, the S01 final review, S01.7-T07), session S005, 2026-09-24 |
| Finding | F-005 (the README status line was changed outside the approved sections) |
| Status | **Decided (S005 E126):** R-11 and R-12 **approved and applied** to `README.md` right after the S01 sign-off. User's answer: "Apply both (Recommended)" |

**Basis rule (as in A001):** only approved decisions and completed, verified work are used. The README is the source of vision (plan, header). Tasks change only the sections their stage document names: the License section (S01.1-T02) and the Development and API usage sections (S01.1-T06, S01.7-T05). Every other README change is proposed here and applied only with the user's approval.

IDs continue from A001 (R-01 to R-10).

---

### R-11: Status line

**Current (line 7, the user's original text):**
> ⚠️ **Status: early development.** This project is in its initial stage. Features described below are the planned scope and are not yet implemented.

**Proposed:**
> ⚠️ **Status: early development.** Stage 1, the NAS core, is complete: the Files area can be managed through a REST API on the same computer (see Development). Everything else below is the planned scope and is not implemented yet.

**Reason:**
- The S01 exit criterion (plan 10.2): "Using only the API (e.g. an HTTP client), files and folders in the files area can be managed reliably, including large resumable uploads. All tests pass in CI."
- This now holds. The demo scripts check it in CI on Linux and Windows (S01.7-T06), and the guide `docs/api/usage.md` shows it step by step (S01.7-T05).
- "Features described below … are not yet implemented" no longer holds for the Files part of "NAS Core".

**History:** the agent made a similar change during S01.7-T05, outside that task's sections. Audit A002 restored the original text (F-005), and this proposal replaces that change.

**When:** after the user signs off S01 (S01.7-T08), because "complete" depends on the sign-off.

---

### R-12: Roadmap, stage 1 done

**Current (Roadmap section):**
```
- [ ] Stage 1: Basic NAS (storage service and API; Files and Photos areas)
```

**Proposed:**
```
- [x] Stage 1: Basic NAS (storage service and API; Files and Photos areas)
```

**Reason:** stage S01 is done once the user signs it off (S01.7-T08). The Photos area exists in stage 1 as a reserved, validated folder with no API (S01.2-T06). Its features come in stage 4, which the roadmap line already implies.

**When:** together with R-11, after the S01 sign-off.

---

**Recommendation:** approve R-11 and R-12, to be applied in S01.7-T08 right after the sign-off.
