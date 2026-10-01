# Security documents

**Designed in alignment with ISO/IEC 27001:2022 Annex A controls (self-assessed; not certified).** No software can be "ISO certified" by itself: certification is an audit of an organization by an accredited auditor. These documents are the project's own, honest account. Most technical controls are still **Planned**: the security stage (S03) is in progress and nothing has been released yet.

**Found a vulnerability?** Report it privately; see [`SECURITY.md`](../../SECURITY.md).

| Document | What it covers |
|---|---|
| [Security policy](security-policy.md) | Scope, principles, honesty rules, the standards used, the owner's four decisions, the tier policy, roles |
| [Statement of Applicability](statement-of-applicability.md) | The ISO/IEC 27001:2022 alignment matrix: all 93 Annex A controls, who each applies to, how it is met, and its status |
| [Threat model overview](threat-model.md) | What is protected, from whom, and the main defenses |
| [Storage security guide](storage-security-guide.md) | How stored data is protected (aligned with ISO/IEC 27040) and what operators must do |
| [Hardening guide](hardening-guide.md) | Checklists for operators on Linux, Raspberry Pi, Windows, and Docker, plus physical security |
| [Crypto policy](crypto-policy.md) | Algorithms, TLS, keys, rotation |
| [Security requirements](security-requirements.md) | OWASP ASVS 5.0 and MASVS 2.1 in scope, mapped to the plan and to ISO controls |
| [Secure coding standard](secure-coding-standard.md) | Rules per language, and the linters that enforce them |
| [Vulnerability management](vulnerability-management.md) | Reporting, handling, fix times, advisories |
| [Incident response](incident-response.md) | What to do when something goes wrong, for operators and for the project |
| [Supply chain](supply-chain.md) | Dependencies, CI, releases, signing, and the owner's repository checklist |

The developer's working records (the full threat model, the risk register, the security backlog) are in [`code-agent-docs/security/`](../../code-agent-docs/security/). The whole repository is public.
