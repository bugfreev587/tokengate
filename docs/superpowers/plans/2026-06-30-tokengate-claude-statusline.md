# TokenGate Claude Statusline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the official TokenGate Claude Code statusline tool with default TokenGate mode, Claude compatibility mode, CLI/settings mode selection, and public docs.

**Architecture:** Implement a portable shell script under `statusline/` with separate TokenGate and Claude render paths. Add focused shell smoke tests, a dedicated Vue docs page, route/sidebar entries, and tests that verify discoverability from CLI Setup.

**Tech Stack:** Bash, `jq`, `curl`, Vue 3, Vue Router, Vitest, Tailwind utility classes.

---

## File Structure

- Create `statusline/tokengate-statusline.sh`: official Claude Code status line command.
- Create `statusline/test_tokengate_statusline.sh`: shell behavior tests with cache fixtures and no network dependency.
- Modify `docs/superpowers/specs/2026-06-30-tokengate-claude-statusline-design.md`: add `--mode` and mode precedence.
- Modify `frontend/src/router/index.ts`: add `/docs/cli/statusline`.
- Modify `frontend/src/config/apiReference.ts`: add sidebar guide entry.
- Modify `frontend/src/config/__tests__/apiReference.spec.ts`: assert sidebar entry.
- Modify `frontend/src/router/__tests__/cli-docs-route.spec.ts`: assert route entry.
- Create `frontend/src/views/public/PublicClaudeStatuslineView.vue`: public docs page.
- Create `frontend/src/views/public/__tests__/PublicClaudeStatuslineView.spec.ts`: docs page tests.
- Modify `frontend/src/views/public/PublicCliDocsView.vue`: link CLI setup to the statusline page.
- Modify `frontend/src/views/public/__tests__/PublicCliDocsView.spec.ts`: assert the link and statusline mention.

## Task 1: Shell Statusline Behavior

**Files:**
- Create: `statusline/test_tokengate_statusline.sh`
- Create: `statusline/tokengate-statusline.sh`

- [ ] **Step 1: Write failing shell tests**

Create `statusline/test_tokengate_statusline.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT="$ROOT/statusline/tokengate-statusline.sh"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

strip_ansi() {
  sed -E $'s/\x1B\\[[0-9;]*[mK]//g'
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"
  if [[ "$haystack" != *"$needle"* ]]; then
    printf 'FAIL: %s\nExpected to contain: %s\nActual: %s\n' "$label" "$needle" "$haystack" >&2
    exit 1
  fi
}

write_input() {
  cat > "$TMP/input.json" <<JSON
{
  "model": { "display_name": "Claude Sonnet" },
  "context_window": {
    "context_window_size": 200000,
    "current_usage": {
      "input_tokens": 40000,
      "cache_creation_input_tokens": 1000,
      "cache_read_input_tokens": 1000
    }
  },
  "cost": 0.1234,
  "cwd": "$ROOT",
  "version": "2.1.0"
}
JSON
}

write_statusline_cache() {
  mkdir -p "$TMP/cache"
  cat > "$TMP/cache/tokengate-statusline-cache.json" <<'JSON'
{
  "ok": true,
  "billing_mode": "API_USAGE",
  "cost": { "today": "1.28" },
  "budgets": {
    "monthly": { "used": "38", "limit": "100", "percent": 38 },
    "daily": { "used": "4", "limit": "20", "percent": 20 }
  }
}
JSON
}

write_input
write_statusline_cache

default_output="$(
  TOKENGATE_STATUSLINE_CACHE_DIR="$TMP/cache" \
  TOKENGATE_API_KEY="tg_test" \
  ANTHROPIC_BASE_URL="https://api.tokengate.to" \
  COLUMNS=240 \
  sh "$SCRIPT" < "$TMP/input.json" | strip_ansi
)"
assert_contains "$default_output" "Claude Sonnet" "default mode includes model"
assert_contains "$default_output" "token-gate@" "default mode includes project and branch"
assert_contains "$default_output" "ctx 42k/200k 21%" "default mode includes context"
assert_contains "$default_output" "$1.28 today" "default mode includes today cost"
assert_contains "$default_output" "month" "default mode includes month budget"
assert_contains "$default_output" "$38/$100 38%" "default mode includes month budget amount"
assert_contains "$default_output" "day" "default mode includes day budget"
assert_contains "$default_output" "$4/$20 20%" "default mode includes day budget amount"

env_claude_output="$(
  TOKENGATE_STATUSLINE_MODE="claude" \
  TOKENGATE_STATUSLINE_CACHE_DIR="$TMP/cache" \
  TOKENGATE_API_KEY="tg_test" \
  ANTHROPIC_BASE_URL="https://api.tokengate.to" \
  COLUMNS=240 \
  sh "$SCRIPT" < "$TMP/input.json" | strip_ansi
)"
assert_contains "$env_claude_output" "21% used" "env mode selects Claude renderer"
assert_contains "$env_claude_output" "79% remain" "env mode keeps Claude remain segment"

arg_claude_output="$(
  TOKENGATE_STATUSLINE_MODE="tokengate" \
  TOKENGATE_STATUSLINE_CACHE_DIR="$TMP/cache" \
  TOKENGATE_API_KEY="tg_test" \
  ANTHROPIC_BASE_URL="https://api.tokengate.to" \
  COLUMNS=240 \
  sh "$SCRIPT" --mode claude < "$TMP/input.json" | strip_ansi
)"
assert_contains "$arg_claude_output" "21% used" "argument mode overrides env mode"

unavailable_output="$(
  TOKENGATE_STATUSLINE_CACHE_DIR="$TMP/empty-cache" \
  TOKENGATE_API_KEY="tg_test" \
  ANTHROPIC_BASE_URL="http://127.0.0.1:9" \
  COLUMNS=240 \
  sh "$SCRIPT" < "$TMP/input.json" | strip_ansi
)"
assert_contains "$unavailable_output" "TokenGate unavailable" "tokengate mode degrades"

empty_output="$(sh "$SCRIPT" </dev/null | strip_ansi)"
assert_contains "$empty_output" "Claude" "empty stdin fallback"

printf 'statusline tests passed\n'
```

- [ ] **Step 2: Run tests and verify RED**

Run:

```bash
bash statusline/test_tokengate_statusline.sh
```

Expected: FAIL because `statusline/tokengate-statusline.sh` does not exist.

- [ ] **Step 3: Implement shell script**

Create `statusline/tokengate-statusline.sh` by adapting `~/burnrate-ai/statusline/tokengate-statusline.sh` and adding:

```bash
statusline_mode="${TOKENGATE_STATUSLINE_MODE:-tokengate}"
while [ "$#" -gt 0 ]; do
  case "$1" in
    --mode)
      statusline_mode="${2:-tokengate}"
      shift 2
      ;;
    --mode=*)
      statusline_mode="${1#--mode=}"
      shift
      ;;
    *)
      shift
      ;;
  esac
done
case "$statusline_mode" in
  claude|tokengate) ;;
  *) statusline_mode="tokengate" ;;
esac
```

Add a TokenGate renderer with this shape:

```bash
build_tokengate_output() {
  local level="$1"
  local s=" ${dim}|${reset} "
  [ "$level" -gt 1 ] && s="${dim}|${reset}"
  local o="${blue}${model_name}${reset}"
  [ -n "$display_dir" ] && o+="${s}${cyan}${display_dir}${reset}"
  o+="${s}${orange}ctx ${used_tokens}/${total_tokens} ${ctx_pct}%${reset}"
  if [ "${#tokengate_parts[@]}" -gt 0 ]; then
    for part in "${tokengate_parts[@]}"; do o+="${s}${part}"; done
  else
    o+="${s}${dim}TokenGate unavailable${reset}"
  fi
  printf "%b" "$o"
}
```

Use `TOKENGATE_STATUSLINE_CACHE_DIR` to let tests isolate cache files.

- [ ] **Step 4: Run tests and verify GREEN**

Run:

```bash
bash statusline/test_tokengate_statusline.sh
```

Expected: `statusline tests passed`.

## Task 2: Mode Selection Spec Update

**Files:**
- Modify: `docs/superpowers/specs/2026-06-30-tokengate-claude-statusline-design.md`

- [ ] **Step 1: Update spec**

Add mode precedence:

```text
Mode resolution order:
1. `--mode <tokengate|claude>` or `--mode=<tokengate|claude>` in the command string.
2. `TOKENGATE_STATUSLINE_MODE` from Claude Code settings env or shell env.
3. Default `tokengate`.
```

Add settings examples for both command argument and env configuration.

- [ ] **Step 2: Self-check**

Run:

```bash
rg -n "Mode resolution|--mode|TOKENGATE_STATUSLINE_MODE|default `tokengate`" docs/superpowers/specs/2026-06-30-tokengate-claude-statusline-design.md
```

Expected: matches for argument mode, env mode, and default mode.

## Task 3: Public Docs Route And Sidebar

**Files:**
- Modify: `frontend/src/router/__tests__/cli-docs-route.spec.ts`
- Modify: `frontend/src/router/index.ts`
- Modify: `frontend/src/config/__tests__/apiReference.spec.ts`
- Modify: `frontend/src/config/apiReference.ts`

- [ ] **Step 1: Write failing router/sidebar tests**

Update router test to assert:

```ts
it('registers the public Claude statusline docs page', async () => {
  const { default: router } = await import('@/router')
  const route = router.getRoutes().find((record) => record.name === 'ClaudeStatuslineDocs')

  expect(route?.path).toBe('/docs/cli/statusline')
  expect(route?.meta.requiresAuth).toBe(false)
  expect(route?.meta.title).toBe('Claude Code Statusline')
})
```

Update sidebar test to assert:

```ts
expect(guideItems).toContainEqual({
  title: 'Claude Code statusline',
  href: '/docs/cli/statusline',
})
```

- [ ] **Step 2: Run tests and verify RED**

Run:

```bash
cd frontend && pnpm test:run src/router/__tests__/cli-docs-route.spec.ts src/config/__tests__/apiReference.spec.ts
```

Expected: FAIL because route/sidebar entry is missing.

- [ ] **Step 3: Add route/sidebar entry**

Add route:

```ts
{
  path: '/docs/cli/statusline',
  name: 'ClaudeStatuslineDocs',
  component: () => import('@/views/public/PublicClaudeStatuslineView.vue'),
  meta: {
    requiresAuth: false,
    title: 'Claude Code Statusline'
  }
}
```

Add sidebar item:

```ts
{ title: 'Claude Code statusline', href: '/docs/cli/statusline' }
```

- [ ] **Step 4: Run tests and verify GREEN**

Run:

```bash
cd frontend && pnpm test:run src/router/__tests__/cli-docs-route.spec.ts src/config/__tests__/apiReference.spec.ts
```

Expected: PASS.

## Task 4: Public Statusline Docs Page

**Files:**
- Create: `frontend/src/views/public/__tests__/PublicClaudeStatuslineView.spec.ts`
- Create: `frontend/src/views/public/PublicClaudeStatuslineView.vue`

- [ ] **Step 1: Write failing docs page test**

Create a Vitest mount test that checks the page text contains:

```ts
expect(text).toContain('Claude Code statusline')
expect(text).toContain('TOKENGATE_STATUSLINE_MODE')
expect(text).toContain('--mode tokengate')
expect(text).toContain('--mode claude')
expect(text).toContain('TokenGate unavailable')
expect(text).toContain('https://api.tokengate.to')
expect(text).toContain('statusLine')
```

- [ ] **Step 2: Run test and verify RED**

Run:

```bash
cd frontend && pnpm test:run src/views/public/__tests__/PublicClaudeStatuslineView.spec.ts
```

Expected: FAIL because the component is missing.

- [ ] **Step 3: Build docs page**

Create the page using existing docs layout components. Include sections for install, configure, mode selection, examples, env vars, and troubleshooting. Keep the page functional and direct rather than marketing-heavy.

- [ ] **Step 4: Run test and verify GREEN**

Run:

```bash
cd frontend && pnpm test:run src/views/public/__tests__/PublicClaudeStatuslineView.spec.ts
```

Expected: PASS.

## Task 5: Link CLI Setup To Statusline Page

**Files:**
- Modify: `frontend/src/views/public/__tests__/PublicCliDocsView.spec.ts`
- Modify: `frontend/src/views/public/PublicCliDocsView.vue`

- [ ] **Step 1: Write failing CLI docs test**

Add expectations:

```ts
expect(text).toContain('Claude Code statusline')
expect(viewSource).toContain('/docs/cli/statusline')
```

- [ ] **Step 2: Run test and verify RED**

Run:

```bash
cd frontend && pnpm test:run src/views/public/__tests__/PublicCliDocsView.spec.ts
```

Expected: FAIL because the link is not present.

- [ ] **Step 3: Add statusline link**

Add a compact link near the Claude Code setup section and sidebar checklist:

```vue
<RouterLink to="/docs/cli/statusline" class="...">
  Claude Code statusline
</RouterLink>
```

- [ ] **Step 4: Run test and verify GREEN**

Run:

```bash
cd frontend && pnpm test:run src/views/public/__tests__/PublicCliDocsView.spec.ts
```

Expected: PASS.

## Task 6: Final Verification, Commit, Push

**Files:**
- All files above.

- [ ] **Step 1: Run shell tests**

```bash
bash statusline/test_tokengate_statusline.sh
```

Expected: `statusline tests passed`.

- [ ] **Step 2: Run frontend targeted tests**

```bash
cd frontend && pnpm test:run src/router/__tests__/cli-docs-route.spec.ts src/config/__tests__/apiReference.spec.ts src/views/public/__tests__/PublicCliDocsView.spec.ts src/views/public/__tests__/PublicClaudeStatuslineView.spec.ts
```

Expected: all selected test files pass.

- [ ] **Step 3: Inspect git diff**

```bash
git diff -- statusline docs/superpowers frontend/src/router/index.ts frontend/src/router/__tests__/cli-docs-route.spec.ts frontend/src/config/apiReference.ts frontend/src/config/__tests__/apiReference.spec.ts frontend/src/views/public/PublicCliDocsView.vue frontend/src/views/public/__tests__/PublicCliDocsView.spec.ts frontend/src/views/public/PublicClaudeStatuslineView.vue frontend/src/views/public/__tests__/PublicClaudeStatuslineView.spec.ts
```

Expected: only statusline/docs/frontend docs changes.

- [ ] **Step 4: Commit only this work**

```bash
git add -f statusline/tokengate-statusline.sh statusline/test_tokengate_statusline.sh docs/superpowers/specs/2026-06-30-tokengate-claude-statusline-design.md docs/superpowers/plans/2026-06-30-tokengate-claude-statusline.md
git add frontend/src/router/index.ts frontend/src/router/__tests__/cli-docs-route.spec.ts frontend/src/config/apiReference.ts frontend/src/config/__tests__/apiReference.spec.ts frontend/src/views/public/PublicCliDocsView.vue frontend/src/views/public/__tests__/PublicCliDocsView.spec.ts frontend/src/views/public/PublicClaudeStatuslineView.vue frontend/src/views/public/__tests__/PublicClaudeStatuslineView.spec.ts
git commit -m "Add TokenGate Claude statusline tool"
```

- [ ] **Step 5: Push main**

```bash
git push origin main
```

Expected: push succeeds.
