ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS stripe_price_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS stripe_trial_days INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS billing_provider VARCHAR(30) NOT NULL DEFAULT 'internal',
    ADD COLUMN IF NOT EXISTS billing_mode VARCHAR(30) NOT NULL DEFAULT 'fixed_period';

CREATE INDEX IF NOT EXISTS idx_subscription_plans_stripe_price_id
    ON subscription_plans(stripe_price_id)
    WHERE stripe_price_id IS NOT NULL AND stripe_price_id <> '';

CREATE INDEX IF NOT EXISTS idx_subscription_plans_billing_provider
    ON subscription_plans(billing_provider);

ALTER TABLE user_subscriptions
    ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS stripe_subscription_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS stripe_price_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS stripe_status VARCHAR(30),
    ADD COLUMN IF NOT EXISTS current_period_start TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS current_period_end TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS trial_start TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS trial_end TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS past_due_since TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS trial_used BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_subscriptions_stripe_subscription_id
    ON user_subscriptions(stripe_subscription_id)
    WHERE stripe_subscription_id IS NOT NULL AND stripe_subscription_id <> '' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_user_subscriptions_stripe_customer_id
    ON user_subscriptions(stripe_customer_id)
    WHERE stripe_customer_id IS NOT NULL AND stripe_customer_id <> '' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_user_subscriptions_stripe_status
    ON user_subscriptions(stripe_status)
    WHERE stripe_status IS NOT NULL AND deleted_at IS NULL;
