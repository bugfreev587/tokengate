# TokenGate P0 Compatibility Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a release gate that proves TokenGate's OpenAI-compatible P0 surface works before launch.

**Architecture:** Add one self-contained shell suite that exercises `/v1/models`, `/v1/chat/completions`, streaming Chat Completions, `/v1/responses`, and optional OpenAI SDK checks. Wire the suite into launch readiness and document the P0 product contract.

**Tech Stack:** Bash, curl, optional Node.js with the `openai` package, existing TokenGate gateway routes.

---

## File Structure

- `docs/TOKENGATE_P0_COMPATIBILITY_PRD.md`: product contract, non-goals, acceptance criteria, release gate, and metrics.
- `tools/tokengate_p0_compatibility_suite.sh`: executable P0 gate.
- `tools/tokengate_launch_readiness.sh`: calls the P0 gate when API smoke is enabled.
- `docs/TOKENGATE_OPERATIONS_RUNBOOK.md`: operator instructions for P0 checks.
- `docs/TOKENGATE_QUICKSTART.md`: quick command for first production verification.
- `docs/TOKENGATE_PRODUCTION_ENV_CHECKLIST.md`: launch checklist wiring.
- `Makefile`: convenience `test-p0-compatibility` target.

## Tasks

### Task 1: Write the P0 PRD

- [x] **Step 1: Define the P0 contract**

Create `docs/TOKENGATE_P0_COMPATIBILITY_PRD.md` with the endpoint list, non-goals, acceptance criteria, release gate command, and metrics.

- [x] **Step 2: Review the PRD for scope creep**

Confirm that Embeddings, Files, Assistants, Audio, Batches, and Fine-tuning are explicitly outside P0.

### Task 2: Add the P0 Suite

- [x] **Step 1: Verify the missing script fails syntax validation**

Run:

```bash
bash -n tools/tokengate_p0_compatibility_suite.sh
```

Expected before implementation: failure because the file does not exist.

- [x] **Step 2: Create the suite script**

Create `tools/tokengate_p0_compatibility_suite.sh` with checks for model visibility, non-streaming chat, streaming chat, Responses API, and optional OpenAI SDK smoke.

- [x] **Step 3: Verify script syntax**

Run:

```bash
bash -n tools/tokengate_p0_compatibility_suite.sh
```

Expected: no output and exit code `0`.

- [x] **Step 4: Verify usage output**

Run:

```bash
tools/tokengate_p0_compatibility_suite.sh --help
```

Expected: usage text includes `TOKENGATE_BASE_URL`, `TOKENGATE_API_KEY`, and `TOKENGATE_OPENAI_MODEL`.

### Task 3: Wire Launch Readiness

- [x] **Step 1: Replace API smoke call**

Modify `tools/tokengate_launch_readiness.sh` so `TOKENGATE_RUN_API_SMOKE=1` runs `tools/tokengate_p0_compatibility_suite.sh`.

- [x] **Step 2: Validate launch readiness syntax**

Run:

```bash
bash -n tools/tokengate_launch_readiness.sh
```

Expected: no output and exit code `0`.

### Task 4: Update Operator Docs

- [x] **Step 1: Update the operations runbook**

Add the P0 suite command and explain that the older smoke script remains useful for narrower provider checks.

- [x] **Step 2: Update the quickstart**

Add the P0 suite command after the first production smoke section.

### Task 5: Final Verification

- [x] **Step 1: Run syntax checks**

Run:

```bash
bash -n tools/tokengate_p0_compatibility_suite.sh
bash -n tools/tokengate_launch_readiness.sh
bash -n tools/tokengate_smoke_test.sh
bash -n tools/tokengate_model_visibility.sh
```

Expected: all commands exit `0`.

- [x] **Step 2: Check changed files**

Run:

```bash
git status --short --untracked-files=all
```

Expected: only the PRD, plan, scripts, docs, and Makefile touched by this task are listed.
