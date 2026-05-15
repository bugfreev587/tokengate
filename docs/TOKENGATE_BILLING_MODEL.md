# TokenGate Billing Model

This document defines the billing language TokenGate should use across product, frontend, backend, and support.

## Core Definitions

### Tokens

Tokens are provider/model usage units.

Use tokens only when the provider natively prices text usage by token count.

Display text model pricing as:

- input price per 1M tokens
- output price per 1M tokens
- cache write/read price per 1M tokens when applicable

### Balance

Balance is the user's billing wallet.

Balance is the unit that payment, included spend, overage, refunds, and manual adjustments settle against.

Display balance in currency-like terms, currently USD.

### Included Usage

Included usage is the spend allowance attached to a subscription plan or group.

Use included usage for:

- daily included spend
- weekly included spend
- monthly included spend

Do not describe included usage as tokens unless the plan is explicitly scoped to one text-only model family.

### Credits

Credits are not the primary public pricing language for TokenGate V1.

If credits remain in internal code or legacy admin features, treat them as platform billing units, not provider tokens.

## Public Pricing Rules

### Text Models

Public pricing should show:

- `$ / 1M input tokens`
- `$ / 1M output tokens`
- optional cache write/read pricing

Example:

```text
Claude Haiku
Input:  $0.80 / 1M tokens
Output: $4.00 / 1M tokens
```

### Image Models

Public pricing should show:

- price per image
- or provider-native output unit if the provider does not sell by image count

Example:

```text
Image generation
$0.04 / image
```

### Video Models

Public pricing should show one provider-native unit:

- per job
- per second
- per generated clip
- per resolution tier

Do not force video into token pricing unless the upstream provider does.

## Deduction Order

TokenGate should explain deductions in this order:

1. Active plan included usage, if the request belongs to a subscribed group and allowance remains.
2. Bonus or promotional balance, if implemented.
3. Prepaid wallet balance.
4. Reject the request with an insufficient balance error.

If the current code path does not yet distinguish bonus and prepaid balance, the UI should only describe a single available balance.

## Subscription Plans

Plans should be described as access plus included spend.

Recommended V1 plan copy:

```text
Starter
Monthly access to selected model families
Includes $X monthly usage
Overages use available balance at published model rates
```

Recommended plan fields:

- plan name
- monthly price
- validity period
- model family or group
- included monthly usage
- rate multiplier, if used
- overage behavior

## User UI Language

Use:

- Available Balance
- Included Usage
- Today Spend
- Model Pricing
- Billing Rate
- Usage Cost
- Top Up

Avoid as primary labels:

- Credits
- Token Balance
- Quota Balance
- Recharge Credits

## Admin UI Language

Admin pages may expose more precise controls, but labels should still map back to public concepts:

- `rate_multiplier`: Billing rate multiplier
- `daily_limit_usd`: Daily included usage
- `weekly_limit_usd`: Weekly included usage
- `monthly_limit_usd`: Monthly included usage
- `balance`: Available balance

## Support Explanation

A support-safe explanation:

```text
TokenGate prices each model using its native usage unit. Text models use per-million-token pricing, media models use per-image or per-video units, and all successful usage is settled against your available balance or plan included usage.
```

## Launch Checklist

Before public launch:

- every public pricing surface uses per-1M token language for text models
- image/video pricing does not mention tokens unless provider-native
- payment top-up pages say balance, not credits
- subscription pages say included usage, not quota
- usage records show both usage units and final cost
- insufficient balance errors point users to top-up or plan upgrade
