# TokenGate Claude Code Connect PRD

Date: 2026-07-02

Status: Draft after PRD review

## 1. Executive Summary

TokenGate should provide a first-class Claude Code setup flow for API keys.
When a user clicks **Use Key**, Claude Code should appear as a visible client
option with guided, copyable setup.

The recommended product is **Claude Code Connect**:

- TokenGate adds a Claude Code tab to the Use Key modal.
- TokenGate generates a Claude Code settings payload for each API key.
- The default payload configures Claude Code to use TokenGate through
  `ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN`, gateway model discovery, and
  model-family pins.
- TokenGate relies on Claude Code gateway discovery as the standard way to add
  TokenGate-served model IDs to `/model`.
- TokenGate may optionally generate an allowlist policy mode with
  `availableModels`, but this is not the default model-sync mechanism.
- The local installer previews, backs up, and merges the settings on the user's
  machine.

This is a local-device configuration problem as much as a gateway problem.
TokenGate can serve the correct model catalog and configuration payload, but it
cannot silently modify a user's local Claude Code configuration from the server.

## 2. Problem

Today, API key onboarding is optimized for Codex CLI, Codex WebSocket, and
OpenCode. Claude Code is not visible in the Use Key modal for some key and group
combinations, which makes TokenGate look incomplete even when the backend can
serve Anthropic-compatible Claude Code traffic.

Model availability is also confusing:

- TokenGate connection tests can confirm that a model such as
  `claude-fable-5` works.
- TokenGate `/v1/models` can return that model.
- Claude Code's `/model` picker may still omit it when local Claude Code is not
  configured for gateway discovery, when discovery is disabled by local policy,
  or when the installed Claude Code version is too old.

The wrong product response is to tell every user to add
`ANTHROPIC_CUSTOM_MODEL_OPTION`. That creates a local one-off shortcut, not a
repeatable TokenGate onboarding experience. It also does not scale when model
catalogs change.

## 3. Product Goals

- Make Claude Code a clear, first-class Use Key destination.
- Give ordinary users one safe path to configure Claude Code against TokenGate.
- Keep TokenGate `/v1/models` as the source of truth for models discoverable by
  Claude Code.
- Support newly available Claude models without requiring users to hand-edit
  local settings.
- Provide clear diagnostics when a key, group, Claude Code version, local flag,
  or model catalog cannot support Claude Code.
- Avoid shipping `ANTHROPIC_CUSTOM_MODEL_OPTION` as the standard setup path.

## 4. Non-Goals

- Silently modifying a user's local `~/.claude/settings.json` from the
  TokenGate web app.
- Depending on `ANTHROPIC_CUSTOM_MODEL_OPTION` as the normal model sync path.
- Making Claude Code work through OpenAI-only groups that cannot serve the
  Anthropic Messages API.
- Full MDM, server-managed settings, or enterprise fleet management in the
  first release.
- Replacing Claude Code's native model picker.

## 5. Target Users

Primary users:

- Developers who create a TokenGate API key and want to use Claude Code locally.
- Users who need a new Claude model soon after TokenGate accounts gain access.
- Existing TokenGate users who discover clients through the API Keys page.

Secondary users:

- TokenGate operators who need a supportable, repeatable onboarding path.
- Admins who need to validate that model lists and API key routing are healthy.

## 6. Product Principles

- Be explicit about local control. The user owns local Claude Code settings.
- Prefer official Claude Code settings over custom picker workarounds.
- Make unsupported states visible. Do not hide Claude Code without explaining
  why.
- Separate "copy config" from "install config". Copyable files are transparent;
  install commands must be idempotent and reversible.
- Avoid silent side effects. Every generated env var must be visible in the UI
  with a short reason.
- Keep secrets out of logs and screenshots. Generated commands must avoid
  echoing API keys after installation.

## 7. Core Technical Strategy

### 7.1 Standard Path

The standard TokenGate Claude Code setup writes only the settings needed for a
gateway-backed Claude Code CLI session:

- `ANTHROPIC_BASE_URL`: points Claude Code at TokenGate.
- `ANTHROPIC_AUTH_TOKEN`: sends the selected TokenGate API key as a bearer
  token.
- `CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY=1`: tells Claude Code to query
  TokenGate `/v1/models` at startup and add discovered Claude model IDs to the
  `/model` picker.
- `ANTHROPIC_DEFAULT_OPUS_MODEL`, `ANTHROPIC_DEFAULT_SONNET_MODEL`,
  `ANTHROPIC_DEFAULT_HAIKU_MODEL`, and `ANTHROPIC_DEFAULT_FABLE_MODEL`: pin
  family aliases to concrete model IDs from TokenGate's live model catalog when
  that family is present.

Discovery is responsible for making gateway model IDs appear. Family pins are
responsible for making aliases such as `opus`, `sonnet`, `haiku`, and `fable`
resolve to TokenGate-supported concrete models.

### 7.2 Optional Allowlist Mode

`availableModels` and `enforceAvailableModels` are model-selection policy
controls. They can hide or reject models outside the effective allowlist, and
`enforceAvailableModels` requires Claude Code v2.1.175 or later.

TokenGate must not use them as the default sync mechanism for personal installs.
They may be offered only as an explicit advanced option:

```text
Pin Claude Code's model picker to this TokenGate key
```

When this option is enabled, the UI must explain the tradeoff:

```text
This limits Claude Code to models returned for this TokenGate key. It can be
useful on a dedicated machine, but surprising if you use multiple gateways or
Claude accounts.
```

This keeps the normal setup aligned with gateway discovery while still allowing
operators to opt into a stricter picker experience.

### 7.3 Explicitly Not Used

TokenGate must not generate `ANTHROPIC_CUSTOM_MODEL_OPTION` in normal product
flows. It remains a manual emergency workaround, not the product path.

## 8. UX Requirements

### 8.1 Use Key Modal Client Tabs

The Use Key modal should always reserve a place for Claude Code in the client
tab row, with one of three states:

- **Available**: key's group can serve Anthropic-compatible `/v1/messages`.
- **Needs compatible group**: key is assigned to an OpenAI/Gemini-only group.
- **Needs admin enablement**: group is OpenAI-shaped but can support Claude Code
  only when Anthropic Messages dispatch is enabled.

For unavailable states, the tab should be visible but disabled, with a tooltip or
inline note:

```text
Claude Code needs an Anthropic-compatible TokenGate group. Select a Claude group
or ask an admin to enable Claude Code dispatch for this group.
```

This prevents the "where is Claude Code?" moment from the current UI.

### 8.2 Claude Code Tab Layout

When available, the Claude Code tab should show:

1. **Recommended install command**
   - macOS/Linux
   - Windows PowerShell
   - Windows CMD only if support is practical; otherwise link to PowerShell.
2. **What this command changes**
   - `~/.claude/settings.json`
   - `ANTHROPIC_BASE_URL`
   - `ANTHROPIC_AUTH_TOKEN`
   - `CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY`
   - model family pins such as `ANTHROPIC_DEFAULT_FABLE_MODEL`
   - optional advanced vars only when enabled, for example
     `CLAUDE_CODE_ATTRIBUTION_HEADER`
3. **Model sync preview**
   - Show the current models TokenGate will expose through `/v1/models`.
   - Show family alias pins derived from that list.
   - Highlight new or premium models, for example Fable.
4. **Local compatibility checks**
   - Minimum Claude Code version for gateway discovery.
   - Minimum Claude Code version for Fable in the picker.
   - Warning if the user's local config disables nonessential traffic, because
     that prevents gateway discovery from running.
5. **Verification commands**
   - `claude --version`
   - `claude --model claude-fable-5 -p "reply ok"` when Fable is available.
   - `/model` instruction for interactive verification.
   - `claude --debug` troubleshooting instruction when discovery does not run.
6. **Manual files**
   - Keep copyable JSON for users who do not run install scripts.

### 8.3 Copy And Install UX

The modal should provide two primary actions:

- **Copy install command**
- **Copy settings JSON**

The install command must be described as local and reversible:

```text
This command updates your local Claude Code settings. It creates a timestamped
backup before writing. You can inspect the script URL before running it.
```

Avoid presenting `curl | bash` as the only path. Provide an inspectable two-step
option:

```bash
curl -fsSL "https://api.tokengate.to/install/claude-code?...token=..." -o /tmp/tokengate-claude-code.sh
sh /tmp/tokengate-claude-code.sh --preview
sh /tmp/tokengate-claude-code.sh --apply
```

### 8.4 Model Sync UX

The user should not need to know which Claude Code environment variable maps to
which model family. TokenGate should display:

- Default model recommendation
- Opus model pin
- Sonnet model pin
- Haiku model pin
- Fable model pin, when available
- Full discoverable model list from `/v1/models`

Recommended copy:

```text
TokenGate will let Claude Code discover models from this key's /v1/models list.
When TokenGate sees a new model for this key, restart Claude Code or rerun the
sync command.
```

Future enhancement: the TokenGate CLI wrapper can run this sync automatically
before launching Claude Code.

## 9. Functional Requirements

### 9.1 Gateway Model Contract

TokenGate `/v1/models` must return all Claude models a key can actually use
through its assigned group. For Anthropic-compatible groups, that list must
include:

- model mapping keys
- refreshed upstream account model catalogs
- TokenGate default Claude models when no narrower catalog exists

The list should be stable, sorted, and deduplicated. It should not include
models that routing will reject.

The endpoint must support Claude Code gateway discovery:

- Serve `GET /v1/models?limit=1000`.
- Return a `data` array with `id` and optional `display_name`.
- Ensure IDs begin with `claude` or `anthropic` if they should appear in the
  Claude Code picker.
- Authenticate both bearer-token discovery and future helper/x-api-key
  discovery paths.
- Avoid redirects and slow responses because Claude Code discovery treats these
  as failures.

### 9.2 Claude Code Settings Payload API

Add an authenticated API endpoint that returns a Claude Code settings payload
for a specific API key.

Proposed endpoint:

```text
GET /api/v1/keys/:id/claude-code/connect
```

Response shape:

```json
{
  "supported": true,
  "reason": null,
  "base_url": "https://api.tokengate.to",
  "settings": {
    "env": {
      "ANTHROPIC_BASE_URL": "https://api.tokengate.to",
      "ANTHROPIC_AUTH_TOKEN": "sk-...",
      "CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY": "1",
      "ANTHROPIC_DEFAULT_OPUS_MODEL": "claude-opus-4-8",
      "ANTHROPIC_DEFAULT_SONNET_MODEL": "claude-sonnet-4-6",
      "ANTHROPIC_DEFAULT_HAIKU_MODEL": "claude-haiku-4-5-20251001",
      "ANTHROPIC_DEFAULT_FABLE_MODEL": "claude-fable-5"
    }
  },
  "optional_policy_settings": {
    "availableModels": [
      "claude-opus-4-8",
      "claude-fable-5",
      "claude-sonnet-4-6",
      "claude-haiku-4-5-20251001"
    ],
    "enforceAvailableModels": false
  },
  "optional_env": {
    "CLAUDE_CODE_ATTRIBUTION_HEADER": {
      "value": "0",
      "default_enabled": false,
      "reason": "Omit Claude Code's system-prompt attribution block only when TokenGate explicitly wants that gateway behavior."
    }
  },
  "models": {
    "default": "claude-opus-4-8",
    "opus": "claude-opus-4-8",
    "sonnet": "claude-sonnet-4-6",
    "haiku": "claude-haiku-4-5-20251001",
    "fable": "claude-fable-5",
    "available": [
      "claude-opus-4-8",
      "claude-fable-5",
      "claude-sonnet-4-6",
      "claude-haiku-4-5-20251001"
    ]
  },
  "minimum_versions": {
    "gateway_discovery": "2.1.129",
    "fable_picker": "2.1.170",
    "enforce_available_models": "2.1.175",
    "sonnet_5_builtin_alias": "2.1.197"
  },
  "recommended_claude_code_version": "latest"
}
```

The Sonnet ID above is an example, not a constant. TokenGate must derive all
family pins from the selected key's live `/v1/models` result. Do not hard-code a
specific Sonnet generation in frontend examples or tests.

If unsupported:

```json
{
  "supported": false,
  "reason": "GROUP_NOT_ANTHROPIC_COMPATIBLE",
  "message": "This API key is assigned to an OpenAI-only group. Select an Anthropic-compatible group to use Claude Code."
}
```

### 9.3 Settings Generation Rules

Default generated settings:

- Include `ANTHROPIC_BASE_URL`.
- Include either `ANTHROPIC_AUTH_TOKEN` or, in a later release,
  `apiKeyHelper`.
- Include `CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY=1`.
- Include family pins for families present in TokenGate `/v1/models`.
- Omit a family pin when that family is absent.
- Omit `model` unless TokenGate intentionally wants to set the initial model.
- Omit `availableModels` and `enforceAvailableModels` unless the user enables
  optional allowlist mode.
- Omit `CLAUDE_CODE_ATTRIBUTION_HEADER` unless the UI explicitly shows and
  enables that advanced behavior.
- Omit `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`; it prevents gateway discovery
  and belongs to a user's privacy/policy choice, not TokenGate's connect
  payload.
- Never include `ANTHROPIC_CUSTOM_MODEL_OPTION`.

Optional allowlist mode:

- Add `availableModels` from the same `/v1/models` list.
- Set `enforceAvailableModels` only when the user chooses strict picker
  enforcement and the detected Claude Code version is v2.1.175 or newer.
- Warn that this mode can conflict with multi-gateway or direct-Anthropic
  workflows.

### 9.4 Installer Script

Add a small installer script that can be served by TokenGate or bundled in the
repo.

Required behavior:

- Works on macOS and Linux.
- Supports `--preview`, `--apply`, `--backup-dir`, and `--settings-path`.
- Reads a signed or short-lived connect payload.
- Creates a timestamped backup before writing.
- Deep-merges `env` and top-level Claude Code settings without deleting
  unrelated user settings.
- Shows a diff or summary before apply.
- Redacts API keys from terminal output.
- Detects and warns if `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`, because
  this prevents gateway discovery.
- Exits with clear codes for invalid JSON, unsupported group, network failure,
  incompatible Claude Code version, disabled discovery, and permission errors.

Windows support can start with PowerShell script generation in the modal, then
move into the installer after macOS/Linux is stable.

### 9.5 TokenGate CLI Wrapper

Future V2:

```bash
tokengate claude
tokengate claude --sync
tokengate claude --model fable
```

The wrapper should:

- Fetch the latest connect payload for the key.
- Update local Claude Code settings if the model catalog changed.
- Warn when local flags prevent discovery.
- Exec the user's installed `claude` binary.
- Never hide the underlying `claude` command from advanced users.

This is the closest user experience to "TokenGate automatically syncs Claude
Code", because it runs on the user's machine with explicit consent.

## 10. Model Selection Details

TokenGate should choose family pins from the available model list:

- Fable: prefer `claude-fable-5`.
- Opus: prefer latest `claude-opus-*` by TokenGate model ranking.
- Sonnet: prefer latest `claude-sonnet-*` by TokenGate model ranking.
- Haiku: prefer latest `claude-haiku-*` by TokenGate model ranking.

If a family is absent, omit that family env variable. Do not invent a model.

If `claude-fable-5` is absent:

- Do not show Fable as a promised model.
- The Use Key modal can still explain how to refresh upstream models.
- Verification commands should target an available model instead.

## 11. UX Edge Cases

- **Key has no group**: show existing warning and disable all client config.
- **OpenAI-only group**: show Codex tabs normally; show disabled Claude Code tab
  with reason.
- **OpenAI group with Messages dispatch**: show Claude Code tab, but label it as
  "Claude Code via OpenAI group" and include a warning that model support
  depends on dispatch configuration.
- **Antigravity group**: show Claude Code and Gemini tabs; Claude Code base URL
  must use the Antigravity Anthropic-compatible path.
- **Claude Code version too old for discovery**: show upgrade command and block
  discovery-based setup.
- **Claude Code version supports discovery but is too old for Fable**: allow
  basic setup, but do not promise Fable picker support.
- **Nonessential traffic disabled**: warn that discovery will not run until the
  user removes or changes that local setting.
- **Model list empty**: use safe default Claude models only if routing supports
  them; otherwise block setup and show diagnostics.
- **Settings write failure**: installer should leave original settings untouched
  and point to the backup location when one was created.

## 12. Security Requirements

- The web app must not display API keys in URLs where browser history, referers,
  or logs can capture them.
- If the installer needs a URL token, it must be short-lived and scoped only to
  reading the connect payload for that key.
- The install script must redact secrets in preview output.
- The settings JSON copy block may contain the key because the user requested
  it, but it should stay behind the authenticated Use Key modal.
- Support logs and analytics must never record the full API key or generated
  settings content.
- Optional attribution-header changes must be visible to the user. They must not
  be silently added to the default payload.

## 13. Analytics And Success Metrics

Track:

- Use Key modal opens.
- Claude Code tab views.
- Unsupported Claude Code tab views by reason.
- Install command copies.
- Settings JSON copies.
- Optional allowlist mode enables.
- Connect payload fetch success/failure.
- Installer preview/apply success/failure.
- Post-install model verification success.
- Discovery failure reasons detected by installer or diagnostics.

Success criteria:

- 90% of users with an Anthropic-compatible key can complete setup from Use Key
  without support.
- Newly available Claude models appear in TokenGate `/v1/models` within one
  successful model refresh cycle.
- Claude Code gateway discovery can add those models to `/model` after a
  supported Claude Code restart.
- Support tickets about missing Fable/Claude Code tab drop after release.

## 14. Acceptance Criteria

- Use Key modal includes a visible Claude Code tab state for every API key.
- Anthropic-compatible API keys produce Claude Code settings with
  `ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN`,
  `CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY=1`, and family pins.
- Generated default settings include `ANTHROPIC_DEFAULT_FABLE_MODEL` when
  `/v1/models` returns `claude-fable-5`.
- Generated default settings do not include `ANTHROPIC_CUSTOM_MODEL_OPTION`.
- Generated default settings do not include `availableModels` or
  `enforceAvailableModels` unless the user enables optional allowlist mode.
- Generated default settings do not include `CLAUDE_CODE_ATTRIBUTION_HEADER`
  unless the user enables the advanced attribution option.
- `/v1/models?limit=1000` returns discoverable Claude IDs for supported keys.
- Applying generated settings and restarting a supported Claude Code version
  makes gateway-discovered models appear in `/model` as "From gateway", unless
  the model is folded into a built-in row by Claude Code.
- OpenAI-only keys explain why Claude Code is unavailable instead of hiding the
  tab.
- Installer preview and apply paths are covered by tests.
- Existing Codex CLI, Codex WebSocket, Gemini CLI, and OpenCode tabs continue to
  render.

## 15. Testing Plan

Backend:

- Unit test model family pin selection.
- Unit test connect payload for Anthropic, OpenAI-only, OpenAI dispatch,
  Gemini, Antigravity, and ungrouped keys.
- Integration test that `/v1/models` and connect payload use the same model
  source.
- Contract test for `GET /v1/models?limit=1000` shape, auth, timeout budget,
  and no redirect.

Frontend:

- UseKeyModal test for visible Claude Code available state.
- UseKeyModal test for disabled Claude Code unsupported state.
- Test generated settings with Fable.
- Test optional allowlist mode copy and warning.
- Regression tests for existing Codex/OpenCode tabs.

Installer:

- Golden tests for JSON merge.
- Backup creation test.
- Redaction test.
- Idempotency test.
- Detection test for `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`.
- Failure mode tests for invalid payload, unsupported group, old Claude Code,
  disabled discovery, and unwritable settings path.

Manual verification:

```bash
claude --version
claude --debug
/model
```

The picker should include a gateway-discovered `claude-fable-5` row when Fable
is in `/v1/models`, the installed Claude Code version supports Fable, and local
settings do not disable discovery.

## 16. Rollout Plan

Phase 1: Documentation and Use Key UX

- Add Claude Code tab visibility states.
- Generate manual settings JSON from frontend or backend.
- Add unsupported-state copy.
- Add discovery and version troubleshooting copy.

Phase 2: Backend Connect Payload

- Add `/api/v1/keys/:id/claude-code/connect`.
- Centralize model family selection and settings generation.
- Add tests.

Phase 3: Installer

- Add macOS/Linux installer with preview/apply.
- Add Use Key install command.
- Add detection for local discovery blockers.

Phase 4: Optional Policy Mode

- Add explicit "pin picker to this key" toggle.
- Generate `availableModels` only when the user opts in.
- Add warning copy and tests.

Phase 5: Auto Sync Wrapper

- Add `tokengate claude --sync`.
- Optionally expose a small packaged CLI or script download.

## 17. Open Questions

- Should TokenGate set `"model": "opus"` or leave the user's current Claude Code
  model untouched?
- Should optional allowlist mode ever set `enforceAvailableModels` by default,
  or should strict enforcement require a second confirmation?
- Should the installer write user-level `~/.claude/settings.json` or project
  local `.claude/settings.local.json` by default?
- How should users rotate TokenGate API keys already installed in Claude Code?
- Do we want a hosted "Connect Claude Code" public doc page separate from the
  authenticated Use Key modal?
- Should TokenGate expose a diagnostics endpoint that compares the selected key,
  group, `/v1/models`, and generated family pins in one response?

## 18. References

- Claude Code model configuration:
  https://code.claude.com/docs/en/model-config
- Claude Code settings:
  https://code.claude.com/docs/en/settings
- Connect Claude Code to an LLM gateway:
  https://code.claude.com/docs/en/llm-gateway-connect
- Claude Code gateway protocol reference:
  https://code.claude.com/docs/en/llm-gateway-protocol
