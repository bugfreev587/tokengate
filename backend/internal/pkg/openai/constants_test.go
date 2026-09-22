package openai

import (
	"slices"
	"testing"
)

func TestDefaultModelsIncludeCurrentCodexModels(t *testing.T) {
	ids := DefaultModelIDs()

	for _, id := range []string{
		"gpt-6-astra",
		"gpt-5.6",
		"gpt-5.6-sol",
		"gpt-5.6-terra",
		"gpt-5.6-luna",
		"gpt-5.5",
	} {
		if !slices.Contains(ids, id) {
			t.Fatalf("DefaultModels missing %s", id)
		}
	}
}

func TestDefaultTestModelUsesCurrentChatGPTCodexModel(t *testing.T) {
	if DefaultTestModel != "gpt-5.6-terra" {
		t.Fatalf("DefaultTestModel = %q, want gpt-5.6-terra", DefaultTestModel)
	}
}
