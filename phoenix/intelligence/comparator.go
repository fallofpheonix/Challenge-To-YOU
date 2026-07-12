package intelligence

import (
	"challenge-to-you/phoenix/git"
	"challenge-to-you/phoenix/validators"
	"fmt"
	"log"
)

type PatchCandidate struct {
	Name         string
	PatchContent string
	Score        int
	Builds       bool
	TestsPassed  bool
}

type PatchComparator struct {
	Git        *git.GitClient
	Validators *validators.ValidationEngine
}

func NewPatchComparator(g *git.GitClient, v *validators.ValidationEngine) *PatchComparator {
	return &PatchComparator{
		Git:        g,
		Validators: v,
	}
}

func (pc *PatchComparator) EvaluateCandidates(candidates []PatchCandidate, failedFile string) (PatchCandidate, error) {
	log.Printf("Evaluating %d patch candidates...", len(candidates))
	var bestCandidate PatchCandidate
	bestScore := -1

	for _, cand := range candidates {
		log.Printf("Testing candidate %s...", cand.Name)

		// Apply candidate patch
		_, errApply := pc.Git.Apply(cand.PatchContent)
		if errApply != nil {
			log.Printf("Candidate %s: apply failed: %v", cand.Name, errApply)
			continue
		}

		cand.Builds = true
		cand.TestsPassed = true
		score := 100

		// Compile check
		if logs, errBuild := pc.Validators.CompileGo(); errBuild != nil {
			log.Printf("Candidate %s: compile failed: %s", cand.Name, logs)
			cand.Builds = false
			score -= 50
		}

		// Run tests check
		if logs, errTest := pc.Validators.RunTests(); errTest != nil {
			log.Printf("Candidate %s: tests failed: %s", cand.Name, logs)
			cand.TestsPassed = false
			score -= 40
		}

		// Run linters check
		if _, errLint := pc.Validators.RunLinter(); errLint != nil {
			score -= 10
		}

		cand.Score = score

		// Revert candidate patch to keep branch clean for next candidate
		_, _ = pc.Git.Checkout(failedFile)

		log.Printf("Candidate %s score: %d", cand.Name, score)
		if score > bestScore {
			bestScore = score
			bestCandidate = cand
		}
	}

	if bestScore == -1 {
		return PatchCandidate{}, fmt.Errorf("all patch candidates failed to apply")
	}

	return bestCandidate, nil
}
