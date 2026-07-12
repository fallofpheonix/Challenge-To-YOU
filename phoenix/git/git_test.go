package git

import (
	"testing"
)

func TestNewGitClient(t *testing.T) {
	client := NewGitClient(".")
	if client.WorkspaceRoot != "." {
		t.Errorf("expected WorkspaceRoot to be '.', got %s", client.WorkspaceRoot)
	}
}

func TestGitActiveBranch(t *testing.T) {
	client := NewGitClient(".")
	branch, err := client.ActiveBranch()
	if err != nil {
		t.Skip("skipping test; git command may not be available inside sandbox")
		return
	}
	if len(branch) == 0 {
		t.Error("expected non-empty branch name")
	}
}
