# P0 Regression Model Update Design

## Goal

Restore the scheduled TokenGate P0 regression monitor by replacing the unsupported
`gpt-5.4` canary model with a model that the production ChatGPT-backed upstream can
actually serve.

## Selected approach

Use `gpt-5.5`, subject to a production probe through the existing GitHub Actions
workflow. The probe must exercise model listing, Chat Completions, streaming Chat
Completions, and Responses with the existing regression API key.

Because the regression API key exists only in GitHub Secrets, temporarily change
the repository variable `TOKENGATE_REGRESSION_OPENAI_MODEL` from `gpt-5.4` to
`gpt-5.5`, dispatch the workflow, and inspect its uploaded log. If the probe fails,
restore `gpt-5.4` immediately and do not update code defaults.

## Durable configuration

If the probe succeeds:

- keep `TOKENGATE_REGRESSION_OPENAI_MODEL=gpt-5.5` in the GitHub repository;
- change the workflow fallback model to `gpt-5.5`;
- change the canary and compatibility-suite defaults and usage examples to
  `gpt-5.5`;
- update the workflow contract test to require the new fallback.

Explicit `TOKENGATE_OPENAI_MODEL` values used by tests remain unchanged because
they verify model forwarding and are not defaults.

## Verification and rollback

The support probe and final verification must both complete with zero P0 failures.
The final workflow artifact must show successful model visibility, non-streaming
Chat Completions, streaming Chat Completions including usage, and Responses API.

If `gpt-5.5` is visible but any generation endpoint rejects it, the model is not
considered supported. Restore the previous repository variable and report the
upstream error before evaluating another model.
