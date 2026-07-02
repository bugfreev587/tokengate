package claude

import (
	"slices"
	"testing"
)

func TestDefaultModelsIncludeLatestGenerallyAvailableClaudeModels(t *testing.T) {
	ids := DefaultModelIDs()

	for _, id := range []string{
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
	if slices.Contains(DefaultModelIDs(), "claude-mythos-5") {
		t.Fatal("limited availability models should only appear when returned by upstream")
	}
}
