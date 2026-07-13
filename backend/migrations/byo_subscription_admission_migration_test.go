package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBYOSubscriptionAdmissionMigrationBackfillsExistingAccounts(t *testing.T) {
	sqlBytes, err := FS.ReadFile("140_byo_subscription_admission.sql")
	require.NoError(t, err)
	sql := string(sqlBytes)

	require.Contains(t, sql, "a.owner_user_id IS NOT NULL")
	require.NotContains(t, sql, "JOIN account_groups")
	require.Contains(t, sql, "stripe_subscription_id")
	require.Contains(t, sql, "status = 'active'")
	require.Contains(t, sql, "expires_at > NOW()")
	require.Contains(t, sql, "'byo_disabled_reason'")
	require.Contains(t, sql, "'subscription_inactive'")
	require.Contains(t, sql, "'byo_operational_schedulable'")
	require.Contains(t, sql, "schedulable = FALSE")
	require.Contains(t, sql, "schedulable = COALESCE")
	require.Contains(t, sql, "LOCK TABLE accounts IN SHARE ROW EXCLUSIVE MODE")
	require.Contains(t, sql, "CREATE OR REPLACE FUNCTION public.enforce_byo_subscription_compatibility_lock()")
	require.Contains(t, sql, "BEFORE INSERT OR UPDATE OF schedulable, owner_user_id ON accounts")
	require.Contains(t, sql, "NEW.schedulable := FALSE")
}
