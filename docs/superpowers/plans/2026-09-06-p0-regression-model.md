# P0 Regression Model Update Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Confirm that production supports `gpt-5.5`, then make it the durable default for the TokenGate P0 regression monitor.

**Architecture:** Use the existing GitHub repository variable as a reversible production probe because the regression API key is available only to Actions. After a successful probe, align the workflow fallback and both shell-script defaults, protected by the existing workflow contract self-test.

**Tech Stack:** GitHub Actions, GitHub CLI, Bash

---

### Task 1: Probe production support for `gpt-5.5`

**Files:**
- Inspect: `.github/workflows/p0-regression-monitor.yml`
- Inspect artifact: `latest.log` from the dispatched workflow run

- [ ] **Step 1: Record and temporarily update the repository variable**

Run:

```bash
gh variable get TOKENGATE_REGRESSION_OPENAI_MODEL --repo bugfreev587/tokengate
gh variable set TOKENGATE_REGRESSION_OPENAI_MODEL --repo bugfreev587/tokengate --body gpt-5.5
```

Expected: the first command prints `gpt-5.4`; the update command exits 0.

- [ ] **Step 2: Dispatch the existing workflow**

Run:

```bash
gh workflow run p0-regression-monitor.yml --repo bugfreev587/tokengate --ref main
```

Expected: dispatch exits 0 and creates a `workflow_dispatch` run.

- [ ] **Step 3: Wait for completion and inspect the artifact**

Run `gh run watch <run-id> --repo bugfreev587/tokengate --exit-status`, download
`tokengate-p0-canary-regression`, and inspect `latest.log`.

Expected: the log reports `Model: gpt-5.5`, `Summary: 0 failure(s)`, and successful
Chat Completions, streaming Chat Completions, and Responses checks.

- [ ] **Step 4: Apply the rollback rule if needed**

If Step 3 fails, run:

```bash
gh variable set TOKENGATE_REGRESSION_OPENAI_MODEL --repo bugfreev587/tokengate --body gpt-5.4
```

Expected: the repository variable is restored before reporting the unsupported
upstream response. Do not continue to Task 2 on failure.

### Task 2: Lock the new defaults with a failing contract test

**Files:**
- Modify: `tools/tokengate_regression_workflow_test.sh`
- Test: `tools/tokengate_regression_workflow_test.sh`

- [ ] **Step 1: Update the expected workflow fallback and add script-default assertions**

Define paths for `tools/tokengate_p0_canary.sh` and
`tools/tokengate_p0_compatibility_suite.sh`, then assert that all three defaults
contain `gpt-5.5`:

```bash
CANARY="$ROOT_DIR/tools/tokengate_p0_canary.sh"
SUITE="$ROOT_DIR/tools/tokengate_p0_compatibility_suite.sh"

assert_file_contains "$WORKFLOW" 'TOKENGATE_OPENAI_MODEL: ${{ vars.TOKENGATE_REGRESSION_OPENAI_MODEL || '\''gpt-5.5'\'' }}'
assert_file_contains "$CANARY" 'OPENAI_MODEL="${TOKENGATE_OPENAI_MODEL:-gpt-5.5}"'
assert_file_contains "$SUITE" 'OPENAI_MODEL="${TOKENGATE_OPENAI_MODEL:-gpt-5.5}"'
```

- [ ] **Step 2: Run the test and verify RED**

Run:

```bash
tools/tokengate_regression_workflow_test.sh
```

Expected: FAIL because the production files still contain `gpt-5.4` defaults.

### Task 3: Change the minimal production defaults

**Files:**
- Modify: `.github/workflows/p0-regression-monitor.yml`
- Modify: `tools/tokengate_p0_canary.sh`
- Modify: `tools/tokengate_p0_compatibility_suite.sh`
- Test: `tools/tokengate_regression_workflow_test.sh`

- [ ] **Step 1: Replace only default and usage-example model names**

Change the workflow fallback and each script's `OPENAI_MODEL` default and usage
example from `gpt-5.4` to `gpt-5.5`. Do not alter explicit model values in
functional tests.

- [ ] **Step 2: Run the focused self-tests and verify GREEN**

Run:

```bash
tools/tokengate_regression_workflow_test.sh
tools/tokengate_p0_canary_test.sh
```

Expected: both commands print their `PASS` summary and exit 0.

- [ ] **Step 3: Check the patch**

Run:

```bash
git diff --check -- .github/workflows/p0-regression-monitor.yml tools/tokengate_p0_canary.sh tools/tokengate_p0_compatibility_suite.sh tools/tokengate_regression_workflow_test.sh
git diff -- .github/workflows/p0-regression-monitor.yml tools/tokengate_p0_canary.sh tools/tokengate_p0_compatibility_suite.sh tools/tokengate_regression_workflow_test.sh
```

Expected: no whitespace errors and only the intended model-default assertions and
replacements.

- [ ] **Step 4: Commit the implementation**

```bash
git add .github/workflows/p0-regression-monitor.yml tools/tokengate_p0_canary.sh tools/tokengate_p0_compatibility_suite.sh tools/tokengate_regression_workflow_test.sh
git commit -m "ci: use gpt-5.5 for P0 regression"
```

### Task 4: Verify the live regression configuration

**Files:**
- Inspect artifact: `latest.log` from the final dispatched workflow run

- [ ] **Step 1: Confirm the live repository variable**

Run:

```bash
gh variable get TOKENGATE_REGRESSION_OPENAI_MODEL --repo bugfreev587/tokengate
```

Expected: `gpt-5.5`.

- [ ] **Step 2: Dispatch and wait for a fresh P0 run**

Run the workflow on `main`, wait with `gh run watch`, and require an exit status of
0.

Expected: the final job conclusion is `success`.

- [ ] **Step 3: Inspect the final artifact**

Download `tokengate-p0-canary-regression` and verify `latest.log` contains:

```text
Model:   gpt-5.5
PASS chat_completions HTTP 200
PASS chat_completions_stream HTTP 200
PASS responses HTTP 200
Summary: 0 failure(s)
```
