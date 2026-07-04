package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripeBillingMigrationAddsRequiredColumns(t *testing.T) {
	sqlBytes, err := FS.ReadFile("138_stripe_billing.sql")
	require.NoError(t, err)
	sql := string(sqlBytes)

	require.Contains(t, sql, "ALTER TABLE subscription_plans")
	require.Contains(t, sql, "stripe_price_id")
	require.Contains(t, sql, "stripe_trial_days")
	require.Contains(t, sql, "billing_provider")
	require.Contains(t, sql, "billing_mode")

	require.Contains(t, sql, "ALTER TABLE user_subscriptions")
	require.Contains(t, sql, "stripe_customer_id")
	require.Contains(t, sql, "stripe_subscription_id")
	require.Contains(t, sql, "stripe_price_id")
	require.Contains(t, sql, "stripe_status")
	require.Contains(t, sql, "current_period_start")
	require.Contains(t, sql, "current_period_end")
	require.Contains(t, sql, "trial_start")
	require.Contains(t, sql, "trial_end")
	require.Contains(t, sql, "cancel_at_period_end")
	require.Contains(t, sql, "past_due_since")
	require.Contains(t, sql, "trial_used")

	require.Contains(t, sql, "idx_user_subscriptions_stripe_subscription_id")
	require.Contains(t, sql, "WHERE stripe_subscription_id IS NOT NULL")
}
