ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_test_user BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE payment_provider_instances
    ADD COLUMN IF NOT EXISTS environment VARCHAR(20) NOT NULL DEFAULT 'live';

UPDATE payment_provider_instances
SET environment = 'live'
WHERE environment IS NULL OR environment = '';

CREATE INDEX IF NOT EXISTS idx_payment_provider_instances_environment
    ON payment_provider_instances(environment);

ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS stripe_sandbox_price_id VARCHAR(128);

CREATE INDEX IF NOT EXISTS idx_subscription_plans_stripe_sandbox_price_id
    ON subscription_plans(stripe_sandbox_price_id)
    WHERE stripe_sandbox_price_id IS NOT NULL AND stripe_sandbox_price_id <> '';

ALTER TABLE user_subscriptions
    ADD COLUMN IF NOT EXISTS stripe_environment VARCHAR(20) NOT NULL DEFAULT 'live',
    ADD COLUMN IF NOT EXISTS stripe_provider_instance_id VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_user_subscriptions_stripe_environment
    ON user_subscriptions(stripe_environment)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_user_subscriptions_stripe_provider_instance_id
    ON user_subscriptions(stripe_provider_instance_id)
    WHERE stripe_provider_instance_id IS NOT NULL AND stripe_provider_instance_id <> '' AND deleted_at IS NULL;
