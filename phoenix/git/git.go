package git

import (
	"bytes"
	"os/exec"
)

type GitClient struct {
	WorkspaceRoot string
}

func NewGitClient(root string) *GitClient {
	return &GitClient{WorkspaceRoot: root}
}

func (g *GitClient) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.WorkspaceRoot
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stdout.String() + "\n" + stderr.String(), err
	}
	return stdout.String(), nil
}

func (g *GitClient) Diff() (string, error) {
	return g.run("diff")
}

func (g *GitClient) Apply(patchContent string) (string, error) {
	cmd := exec.Command("git", "apply", "-")
	cmd.Dir = g.WorkspaceRoot
	cmd.Stdin = bytes.NewBufferString(patchContent)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stdout.String() + "\n" + stderr.String(), err
	}
	return stdout.String(), nil
}

func (g *GitClient) Stash() (string, error) {
	return g.run("stash")
}

func (g *GitClient) Checkout(file string) (string, error) {
	return g.run("checkout", "--", file)
}

func (g *GitClient) CheckoutBranch(branch string) (string, error) {
	return g.run("checkout", branch)
}

func (g *GitClient) Commit(message string) (string, error) {
	_, errAdd := g.run("add", ".")
	if errAdd != nil {
		return "", errAdd
	}
	return g.run("commit", "-m", message)
}

func (g *GitClient) ActiveBranch() (string, error) {
	out, err := g.run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return bytes.NewBufferString(out).String(), nil
}
