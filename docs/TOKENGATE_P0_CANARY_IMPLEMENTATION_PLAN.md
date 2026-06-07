# TokenGate P0 Canary Implementation Plan

**Goal:** Add a scheduled production canary path that continuously proves
TokenGate's P0 OpenAI-compatible gateway remains healthy after launch.

**Architecture:** Keep the canary outside the application runtime for this
phase. A shell runner calls `tools/tokengate_p0_compatibility_suite.sh`, captures
the result, writes a small status artifact, and optionally posts a JSON alert to
an operator webhook. This lets Railway Cron, GitHub Actions, or any external
monitor run the same P0 gate without adding new database tables or UI.

**Tech Stack:** Bash, curl, existing TokenGate P0 compatibility suite, optional
generic webhook receiver.

---

## Scope

This phase adds:

- a canary runner that can run once or loop at a configurable interval
- structured status output for the latest run
- optional webhook notification on failure or every run
- shell self-tests for success, failure, and webhook payload behavior
- docs for smoke-key setup, scheduling, and alert triage

This phase does not add:

- a new backend scheduler
- a new admin UI page
- persistent database records
- provider-specific alert integrations beyond generic JSON webhook posts

## Files

- Create `tools/tokengate_p0_canary.sh`
- Create `tools/tokengate_p0_canary_test.sh`
- Modify `Makefile`
- Modify `docs/TOKENGATE_OPERATIONS_RUNBOOK.md`
- Modify `docs/TOKENGATE_PRODUCTION_ENV_CHECKLIST.md`
- Modify `docs/TOKENGATE_P0_COMPATIBILITY_PRD.md`

## Tasks

### Task 1: Canary Tests

- Write shell tests with fake suite and fake curl binaries.
- Verify success exits `0`, writes status JSON, and skips alerts by default.
- Verify failure exits non-zero and sends a JSON alert when webhook is set.
- Verify `TOKENGATE_P0_CANARY_NOTIFY_ON=always` sends success notifications.

### Task 2: Canary Runner

- Validate required `TOKENGATE_BASE_URL` and `TOKENGATE_API_KEY`.
- Run the P0 suite once by default.
- Support loop mode with `TOKENGATE_P0_CANARY_INTERVAL_SECONDS`.
- Capture output in `TOKENGATE_P0_CANARY_STATE_DIR`.
- Redact API keys from logs and webhook payloads.
- Send generic JSON webhook alerts with status, model, backend URL, exit code,
  started/finished timestamps, and an output excerpt.

### Task 3: Workflow Integration

- Add `make test-p0-canary`.
- Document production smoke-key creation and rotation.
- Document sample Cron/GitHub Actions/Railway job usage.
- Add canary and alerting to the production checklist.

### Task 4: Verification

- Run `bash -n` on all affected shell scripts.
- Run the canary self-test.
- Run the canary once against production with the current smoke key.
- Run `git diff --check`.
