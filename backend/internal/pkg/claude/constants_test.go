package claude

import (
	"slices"
	"testing"
)

func TestDefaultModelsIncludeLatestGenerallyAvailableClaudeModels(t *testing.T) {
	ids := DefaultModelIDs()

	for _, id := range []string{
		"claude-fable-5-1",
		"claude-opus-5",
		"claude-sonnet-5",
		"claude-opus-4-8",
		"claude-fable-5",
	} {
		if !slices.Contains(ids, id) {
			t.Fatalf("DefaultModels missing %s", id)
		}
	}
}

func TestDefaultModelsDoNotIncludeLimitedAvailabilityModels(t *testing.T) {
	for _, id := range []string{"claude-mythos-5", "claude-mythos-5-1"} {
		if slices.Contains(DefaultModelIDs(), id) {
			t.Fatalf("limited availability model %s should only appear when returned by upstream", id)
		}
	}
}
