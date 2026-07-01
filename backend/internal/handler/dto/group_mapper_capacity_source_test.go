package dto

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGroupFromServiceIncludesCapacitySource(t *testing.T) {
	group := &service.Group{
		ID:             42,
		Name:           "byo-openai",
		Platform:       service.PlatformOpenAI,
		RateMultiplier: 1,
		Status:         service.StatusActive,
		CapacitySource: service.CapacitySourceConnectedAccount,
	}

	got := GroupFromService(group)

	require.NotNil(t, got)
	require.Equal(t, service.CapacitySourceConnectedAccount, got.CapacitySource)
}
