package intelligence

import (
	"challenge-to-you/phoenix/git"
	"challenge-to-you/phoenix/validators"
	"os"
	"testing"
)

func TestPatchComparator(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "phoenix_comparator_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	gitClient := git.NewGitClient(tempDir)
	valEngine := validators.NewValidationEngine(tempDir)
	comparator := NewPatchComparator(gitClient, valEngine)

	candidates := []PatchCandidate{
		{
			Name:         "PatchA",
			PatchContent: "diff --git a/mock_code.go b/mock_code.go\n...",
		},
	}

	_, errEval := comparator.EvaluateCandidates(candidates, "mock_code.go")
	// Since tempDir is not a Git repo, this should fail to apply or skip
	if errEval == nil {
		t.Error("expected candidate evaluation to fail outside git repository")
	}
}
