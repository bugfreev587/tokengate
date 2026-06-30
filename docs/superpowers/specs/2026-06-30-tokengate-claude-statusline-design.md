# TokenGate Claude Code Statusline Design

## Context

TokenGate needs an official Claude Code status line tool that lives in this repository and is documented from the public CLI docs. A working prototype already exists in `~/burnrate-ai/statusline/tokengate-statusline.sh` and is currently installed through `~/.claude/statusline-command.sh`.

The prototype is valuable because it already reads Claude Code status line JSON from stdin, renders model/context/project information, polls TokenGate for API usage data, fetches Claude OAuth usage windows for subscription users, caches network calls, and degrades when remote data is unavailable.

The TokenGate product version should preserve that compatibility while making TokenGate cost and budget visibility the default experience.

## Goals

- Ship one official script in the TokenGate repo for Claude Code status line integration.
- Support two user-selectable modes:
  - `tokengate`: the default TokenGate-first budget and cost display.
  - `claude`: a compatibility mode that preserves the current burnrate-ai behavior.
- Add a public docs page under the CLI Setup area with install, configuration, mode switching, examples, and troubleshooting.
- Keep the tool safe for daily terminal use: fast, cached, quiet on errors, and non-blocking.

## Non-Goals

- No dashboard one-click installer in this scope.
- No automatic modification of a user's `~/.claude/settings.json` in this scope.
- No dependency on a successful TokenGate network call to render a usable Claude Code status line.
- No hidden collection of prompt or response content. The script only reads Claude Code status metadata and TokenGate usage summaries.

## User Experience

### Default: TokenGate Mode

`TOKENGATE_STATUSLINE_MODE` defaults to `tokengate`.

Example:

```text
Claude Sonnet | token-gate@main | ctx 42k/200k 21% | $1.28 today | month ●●●○○○ $38/$100 38% | day ●○○○○○ $4/$20 20%
```

This mode is for users who route Claude Code through TokenGate and want immediate cost and budget awareness. It prioritizes:

- Active Claude model.
- Current directory and git branch.
- Context-window usage.
- Today's TokenGate spend.
- Configured TokenGate budget periods, especially monthly and daily budgets.

If TokenGate data cannot be fetched, the status line still renders local Claude Code context:

```text
Claude Sonnet | token-gate@main | ctx 42k/200k 21% | TokenGate unavailable
```

### Optional: Claude Mode

Users can switch to the compatibility display:

```bash
export TOKENGATE_STATUSLINE_MODE=claude
```

Claude mode copies the current burnrate-ai behavior. It is useful for users who care more about Claude Code session/context information and Claude OAuth 5h/7d rate-limit windows than the TokenGate-first budget view.

Example:

```text
Claude Sonnet | token-gate@main | 42k/200k | 21% used | 79% remain | 5h ●●○○○○ 34% @3:20pm | 7d ●○○○○○ 12% @Jul 2
```

## Configuration

The script is installed by copying it to the Claude config directory and making it executable:

```bash
mkdir -p ~/.claude
curl -o ~/.claude/tokengate-statusline.sh \
  https://raw.githubusercontent.com/bugfreev587/TokenGate.to/main/statusline/tokengate-statusline.sh
chmod +x ~/.claude/tokengate-statusline.sh
```

Claude Code configuration:

```json
{
  "statusLine": {
    "type": "command",
    "command": "sh ~/.claude/tokengate-statusline.sh"
  }
}
```

TokenGate mode uses the same environment users already configure for Claude Code routing:

- `ANTHROPIC_BASE_URL`: TokenGate gateway root, such as `https://api.tokengate.to`.
- `TOKENGATE_API_KEY`: preferred explicit TokenGate API key.
- `ANTHROPIC_AUTH_TOKEN` or `ANTHROPIC_API_KEY`: fallback key locations for users who already configured Claude Code this way.
- `ANTHROPIC_CUSTOM_HEADERS`: fallback source for `X-TokenGate-Key`.

Optional settings:

- `TOKENGATE_STATUSLINE_MODE`: `tokengate` or `claude`; default `tokengate`.
- `TOKENGATE_STATUSLINE_POLL`: cache TTL in seconds; default `60`.
- `TOKENGATE_STATUSLINE_BARS`: progress bar block count; default `6`.
- `TOKENGATE_STATUSLINE_ENDPOINT`: optional override for the TokenGate status endpoint.
- `COLUMNS`: terminal width override for adaptive density.

## Data Flow

Claude Code invokes the script as a command status line provider and sends status JSON on stdin. The script parses:

- `.model.display_name`
- `.context_window.context_window_size`
- `.context_window.current_usage.input_tokens`
- `.context_window.current_usage.cache_creation_input_tokens`
- `.context_window.current_usage.cache_read_input_tokens`
- `.cost`
- `.cwd`
- `.version`

In `tokengate` mode, the script:

1. Builds the local Claude Code segment from stdin.
2. Resolves the TokenGate API key.
3. Resolves the TokenGate base URL.
4. Reads cached TokenGate data if it is younger than `TOKENGATE_STATUSLINE_POLL`.
5. Otherwise fetches TokenGate usage data.
6. Renders the TokenGate-first status line.

In `claude` mode, the script uses the current burnrate-ai behavior:

1. Builds the local Claude Code segment.
2. Auto-detects billing mode.
3. For monthly subscription mode, fetches Claude OAuth usage windows.
4. For API usage mode, fetches TokenGate budget data.
5. Renders the adaptive Claude-first status line.

## TokenGate Data Contract

The preferred TokenGate endpoint is:

```http
GET /v1/statusline
X-TokenGate-Key: <tokengate-api-key>
Accept: application/json
```

Preferred response shape:

```json
{
  "ok": true,
  "billing_mode": "API_USAGE",
  "cost": {
    "today": "1.28"
  },
  "budgets": {
    "monthly": {
      "used": "38",
      "limit": "100",
      "percent": 38
    },
    "daily": {
      "used": "4",
      "limit": "20",
      "percent": 20
    }
  }
}
```

Current TokenGate routing already exposes `GET /v1/usage`, which can provide API-key-scoped usage, quota, and rate-limit data. The implementation should prefer `/v1/statusline` when present and may fall back to `/v1/usage` to avoid making the client useless while the dedicated statusline endpoint is absent.

The fallback mapping from `/v1/usage` is:

- `usage.today.cost` or `usage.today.actual_cost` maps to today's spend.
- `quota.limit` and `quota.used` map to a total budget-like segment when available.
- `rate_limits[]` can map to day/week/hour quota segments if present.

If neither endpoint returns parseable data, TokenGate mode shows `TokenGate unavailable`.

## Rendering Rules

The output is a single line with ANSI color. It uses conservative labels and avoids long explanatory text because it is always visible during coding.

TokenGate mode order:

1. Model.
2. Project directory and git branch.
3. Context usage as `ctx used/total percent`.
4. Today's spend.
5. Budget segments sorted by product value: month, week, day, total, then other windows.
6. Degraded TokenGate status if no remote data is available.

Colors:

- Model: blue.
- Project: cyan.
- Context token count: orange.
- Spend: magenta.
- Healthy budget: green.
- Warning budget: yellow or orange.
- Critical budget: red.
- Separators and degraded details: dim gray.

Budget thresholds:

- `< 50%`: green.
- `50-69%`: yellow.
- `70-89%`: orange.
- `>= 90%`: red.

The script tries three density levels:

- Full: spaced separators, labels, progress bars.
- Compact: tight separators and shorter labels.
- Minimal: key numbers and degraded state only.

## Architecture

The script remains a portable shell script because Claude Code status line commands should be easy to inspect, copy, and run without a project build step.

Recommended file:

```text
statusline/tokengate-statusline.sh
```

High-level modules inside the script:

- Input parsing: read stdin once and extract Claude Code metadata with `jq`.
- Formatting helpers: token formatting, money formatting, progress bars, colors, width measurement.
- Environment resolution: mode, API key, base URL, cache TTL, bar count.
- TokenGate fetcher: `/v1/statusline` preferred, `/v1/usage` fallback, cached JSON.
- Claude fetcher: current OAuth usage logic from burnrate-ai for `claude` mode.
- Renderers: separate `render_tokengate_mode` and `render_claude_mode` paths.

The split renderers keep product semantics clear: TokenGate mode is cost governance first; Claude mode is Claude usage-window visibility first.

## Documentation

Add a public page:

```text
frontend/src/views/public/PublicClaudeStatuslineView.vue
```

Route:

```text
/docs/cli/statusline
```

Sidebar:

```text
Guides
- CLI setup
- Claude Code statusline
```

The page should include:

- What the statusline does.
- Install command.
- Claude Code `settings.json` configuration.
- TokenGate mode setup and default behavior.
- Claude mode setup.
- Example outputs for normal and degraded states.
- Environment variable reference.
- Troubleshooting for missing `jq`, wrong base URL, missing TokenGate key, unavailable endpoint, and narrow terminals.

The existing `/docs/cli` page should link to the statusline page from its Claude Code section or setup checklist so users naturally discover the enhancement while setting up Claude Code.

## Testing Strategy

Script tests should cover behavior, not implementation details:

- Empty stdin renders a safe fallback.
- TokenGate mode is the default when `TOKENGATE_STATUSLINE_MODE` is unset.
- `TOKENGATE_STATUSLINE_MODE=claude` uses the compatibility renderer.
- TokenGate mode renders model, repo branch, context usage, today cost, and budget segments from a fixture response.
- TokenGate mode degrades to `TokenGate unavailable` when the endpoint fails.
- API key resolution follows the documented precedence.
- Width adaptation removes lower-priority detail when `COLUMNS` is small.

Frontend tests should cover:

- Router registers `/docs/cli/statusline` as a public route.
- Sidebar Guides includes `Claude Code statusline`.
- The docs page shows install, mode switching, default `tokengate`, degraded output, and environment variables.
- Existing `/docs/cli` still renders and links to the new page.

Verification commands:

```bash
cd frontend && pnpm test:run src/router/__tests__/cli-docs-route.spec.ts src/views/public/__tests__/PublicCliDocsView.spec.ts
```

Add any new frontend spec files to that command during implementation.

For shell behavior, add a focused shell test or executable smoke script under `statusline/` or `tools/` and run it directly with fixture JSON.

## Rollout

This can ship as a repository feature before any dashboard installer exists. Users install by copying the raw script URL from docs.

The release should call out:

- Default mode is `tokengate`.
- `claude` mode preserves the old burnrate-ai behavior.
- TokenGate mode fails quiet and never blocks Claude Code.
- Full TokenGate budget display depends on TokenGate usage/statusline endpoints being reachable with the configured API key.

## Open Decisions Resolved

- Default mode: `tokengate`.
- Mode switch variable: `TOKENGATE_STATUSLINE_MODE`.
- `claude` mode behavior: copy the existing burnrate-ai script behavior.
- Docs location: a dedicated `/docs/cli/statusline` page linked under CLI Setup guides.
