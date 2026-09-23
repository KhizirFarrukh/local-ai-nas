# ADR-0018: AI models: CLIP-family classification, YuNet, SFace, HDBSCAN; license exclusions

| Field | Value |
|---|---|
| Number | ADR-0018 |
| Status | **Accepted** (direction). Exact model variants are **deferred** to S12, chosen by benchmarking on the S12.11 evaluation set on CPU-only hardware |
| Date proposed | 2026-09-24 (session S003) |
| Date of last status change | 2026-09-24 (session S003) |
| Supersedes | none |
| Superseded by | none |

## Context

S12 needs:
- Photo classification with an extendable taxonomy (FR-033, FR-034, FR-138).
- Face detection (FR-039, FR-137) and face recognition and grouping (FR-040), with user corrections (FR-044).
- Optional OCR (FR-141).

P003 licensing policy: every model must allow **anyone to deploy and use** this project, with no non-commercial or research-only restrictions (NFR-029).

## Options considered

| Task | Options | Chosen |
|---|---|---|
| Classification | CLIP-family zero-shot (SigLIP, OpenCLIP); fixed-label CNN; captioning VLM | **CLIP-family zero-shot, e.g. SigLIP**, exported to ONNX |
| Face detection | YuNet; SCRFD (InsightFace); RetinaFace | **YuNet** (OpenCV Zoo) |
| Face recognition | SFace; ArcFace/InsightFace | **SFace** (OpenCV Zoo) |
| Grouping | HDBSCAN; DBSCAN; Chinese Whispers | **HDBSCAN (scikit-learn)** + nearest-group assignment |
| OCR (if approved) | RapidOCR (ONNX PaddleOCR); Tesseract | **RapidOCR** |

## Decision (direction Accepted)

- **Photo classification:** a **CLIP-family zero-shot model such as SigLIP**, exported to ONNX.
  - Categories are **text prompts** (e.g. "a photo of a receipt"), so adding a category means adding a line of text, with no retraining.
  - **Image embeddings are kept in internal app data** so semantic search (S12.10) is possible later.
  - Verified licenses: `google/siglip-base-patch16-224` and `google/siglip2-base-patch16-224` are **apache-2.0** (Hugging Face model cards).
- **Face detection:** **YuNet** (OpenCV Zoo). **MIT** (verified: "All files in this directory are licensed under MIT License").
- **Face recognition:** **SFace** embeddings (OpenCV Zoo). **Apache-2.0** (verified, same wording).
- **Runtime for YuNet/SFace:** ONNX Runtime directly, or OpenCV's `FaceDetectorYN`/`FaceRecognizerSF` via **opencv-python-headless 5.0.0.93** (Apache-2.0, verified) if its pre- and post-processing saves meaningful effort. S12.1 decides.
- **Face grouping:** **HDBSCAN** (scikit-learn **1.9.1**, BSD-3-Clause, verified) for the initial clustering.
  - New faces are assigned to the nearest existing group within a similarity threshold.
  - **Periodic re-clustering must preserve user corrections** (S12.6).
- **OCR (only if S12.10 is approved):** **RapidOCR** (rapidocr **3.9.2**, Apache-2.0, verified), an ONNX version of PaddleOCR. The PaddleOCR model weights' license is to be verified in S12.10.
- **Embedding storage:** the agent decides at S12. Likely SQLite blobs with brute-force cosine similarity, which is fast enough at household scale; `sqlite-vec` only if needed.
- **Excluded:** **InsightFace pretrained models** (weights licensed for non-commercial research only) and **any other weights with similar restrictions**.
- **Invariant note:** **semantic search needs a small text model to run at query time.** That is an **exception to invariant I4**, so it **requires explicit user approval before it is built**.

## Consequences

- **Easier:** open-vocabulary categories; permissive licenses; CPU-friendly models.
- **Harder:** zero-shot accuracy tuning (thresholds and prompt ensembles); SFace accuracy is below that of excluded models, so more manual merges are expected.
- **Required:**
  - Exact variants chosen in S12 by benchmark. The evaluation set is license-clean or consented and kept outside the repository.
  - A model registry recording name, version, license, and checksum.
  - **For the user's attention (a factual note, not a legal judgment):** models released under permissive licenses may have been trained on datasets that carry their own research-only terms. YuNet was trained on WIDER FACE (per its upstream documentation), and face recognition models are commonly trained on public celebrity-face datasets. The **weight licenses** themselves (MIT / Apache-2.0) meet the P003 policy. The training-data question is recorded here for transparency and should be re-checked for the exact variants in S12.

## Approval record

> Decision made by the user in plan change request #3 (`code-agent-docs/prompts/P003-technology-stack.json`, `technology_stack.ai_models`: "The direction is Accepted now. Exact model variants are 'deferred' until S12")
> (2026-09-24, session S003). Licenses verified in S003 log E005.

## Links

- **Related requirements:** FR-033, FR-034, FR-039, FR-040, FR-044, FR-054, FR-137, FR-138, FR-141, FR-142, NFR-011, NFR-028, NFR-029
- **Related ADRs:** ADR-0017; invariant I4 (exception needs approval)
- **Related stages:** S12
- **Plan version:** 0.3.0
