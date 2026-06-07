# TokenGate P0 Compatibility PRD

## Purpose

TokenGate P0 compatibility is the public contract that lets an existing
OpenAI Chat Completions or Responses API integration migrate by changing
configuration instead of rewriting business logic.

P0 is intentionally smaller than the full OpenAI API platform. The goal is a
stable, testable gateway surface for chat, responses, model discovery, usage
settlement, and operational launch gates.

## Target Users

- Developer teams already using the OpenAI SDK for chat requests.
- Indie SaaS builders who need one TokenGate API key, usage metering, and
  upstream account routing.
- TokenGate operators who need a release gate before opening public signup.

## P0 Contract

TokenGate P0 compatibility covers:

- `GET /v1/models`
- `POST /v1/chat/completions`
- `POST /v1/responses`
- non-streaming Chat Completions responses
- streaming Chat Completions SSE responses
- `stream_options.include_usage=true`
- basic Responses API text input
- model visibility for the requested OpenAI-compatible model
- TokenGate bearer API key authentication
- usage records and balance settlement after successful requests
- clear operator diagnostics for `401`, `403`, `404`, `405`, `429`, and `5xx`

P0 does not cover:

- full OpenAI API parity
- Assistants, Threads, Runs, Vector Stores, Files, Batches, Fine-tuning, Audio,
  Embeddings, or Moderations
- every OpenAI request parameter
- every third-party OpenAI-compatible upstream behavior

## Product Positioning

Public wording should say:

> TokenGate provides OpenAI-compatible Chat Completions and Responses gateway
> endpoints. Existing OpenAI SDK chat integrations usually migrate by changing
> `baseURL`, API key, and model ID. Integrations that depend on other OpenAI
> APIs require compatibility review.

Avoid saying:

> TokenGate is fully OpenAI API compatible.

## Acceptance Criteria

Before TokenGate is declared P0-compatible:

- A real TokenGate API key can list OpenAI-compatible models.
- The configured P0 model appears in `/v1/models`.
- A non-streaming `/v1/chat/completions` request returns `2xx`.
- A streaming `/v1/chat/completions` request returns SSE data and `[DONE]`.
- Streaming output includes a usage chunk when usage is required.
- A `/v1/responses` request returns `2xx`.
- Usage and balance changes are visible in the product after successful calls.
- Optional OpenAI SDK smoke passes in any environment that installs the SDK.
- Launch readiness can run this P0 gate from one command.

## Release Gate

Operators should run:

```bash
TOKENGATE_BASE_URL="https://<backend-domain>" \
TOKENGATE_API_KEY="sk-..." \
TOKENGATE_OPENAI_MODEL="gpt-5.4" \
tools/tokengate_p0_compatibility_suite.sh
```

For a strict SDK gate, install the OpenAI Node SDK in the runner and add:

```bash
TOKENGATE_RUN_P0_SDK=1
```

## Metrics

Track these signals for every P0 smoke and canary:

- model list HTTP status
- chat completion HTTP status
- streaming completion HTTP status
- responses HTTP status
- time to first byte or first token where available
- upstream account ID and model
- usage record creation
- billing settlement result
- categorized failure reason

## Implementation Plan

Phase 1 adds a P0 compatibility suite and wires it into launch readiness.
Phase 2 adds an external canary runner that reuses the P0 suite, writes the
latest status artifact, and can notify an operator webhook.
Phase 3 can add an in-product health dashboard or Embeddings only after a real
customer migration requires it.
