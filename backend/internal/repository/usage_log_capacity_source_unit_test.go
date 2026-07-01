//go:build unit

package repository

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestPrepareUsageLogInsertPreservesCapacitySource(t *testing.T) {
	groupID := int64(9)
	subscriptionID := int64(10)
	prepared := prepareUsageLogInsert(&service.UsageLog{
		UserID:         1,
		APIKeyID:       2,
		AccountID:      3,
		RequestID:      "req-byo",
		Model:          "gpt-4",
		GroupID:        &groupID,
		SubscriptionID: &subscriptionID,
		CapacitySource: service.CapacitySourceConnectedAccount,
		InputTokens:    10,
		OutputTokens:   20,
	})

	require.GreaterOrEqual(t, len(prepared.args), 10)
	require.Equal(t, service.CapacitySourceConnectedAccount, prepared.args[9])
}
