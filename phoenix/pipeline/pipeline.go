package pipeline

import (
	"challenge-to-you/phoenix/agents"
	"challenge-to-you/phoenix/config"
	"challenge-to-you/phoenix/git"
	"challenge-to-you/phoenix/memory"
	"challenge-to-you/phoenix/prompts"
	"challenge-to-you/phoenix/validators"
	"fmt"
	"log"
	"strings"
	"time"
)

type RepairPipeline struct {
	Config     *config.Config
	Git        *git.GitClient
	Memory     *memory.MemoryManager
	Prompts    *prompts.PromptLibrary
	Validators *validators.ValidationEngine
	Planner    agents.Agent
	Repairer   agents.Agent
	Reviewer   agents.Agent
}

func NewRepairPipeline(cfg *config.Config) *RepairPipeline {
	gitClient := git.NewGitClient(cfg.WorkspaceRoot)
	memoryMgr := memory.NewMemoryManager(cfg.BrainDirectory)
	promptLib := prompts.LoadPrompts(cfg.WorkspaceRoot)
	valEngine := validators.NewValidationEngine(cfg.WorkspaceRoot)

	plannerAgent := agents.NewLocalAgent("Planner", []string{"planning"}, cfg.OllamaHost)
	repairerAgent := agents.NewLocalAgent("Repairer", []string{"coding", "repair"}, cfg.OllamaHost)
	reviewerAgent := agents.NewLocalAgent("Reviewer", []string{"reviewing"}, cfg.OllamaHost)

	return &RepairPipeline{
		Config:     cfg,
		Git:        gitClient,
		Memory:     memoryMgr,
		Prompts:    promptLib,
		Validators: valEngine,
		Planner:    plannerAgent,
		Repairer:   repairerAgent,
		Reviewer:   reviewerAgent,
	}
}

func (p *RepairPipeline) RunSelfRepair(diagnosticError string, failedFilePath string) error {
	log.Printf("Starting Project Phoenix Self-Repair for diagnostic: %s in file: %s", diagnosticError, failedFilePath)

	// Step 1: Query memory for similar solutions
	similarRepairs, errSearch := p.Memory.SearchSimilar(diagnosticError)
	var fewShots string
	if errSearch == nil && len(similarRepairs) > 0 {
		fewShots = fmt.Sprintf("Similar past successful patch:\n%s", similarRepairs[0].AppliedPatch)
	}

	// Step 2: Invoke Planner Agent
	plannerPrompt := fmt.Sprintf("%s\nFailed File: %s\nError Detail: %s", p.Prompts.PlannerPrompt, failedFilePath, diagnosticError)
	planResponse, errPlan := p.Planner.Run(plannerPrompt, p.Config.ModelRouter["planning"])
	if errPlan != nil {
		log.Printf("Warning: Planner agent invocation failed (falling back to direct repair): %v", errPlan)
	} else {
		log.Printf("Planner Action Plan:\n%s", planResponse)
	}

	// Step 3: Run Self-Repair patch generation loop
	var activePatch string
	success := false
	startTime := time.Now()

	for attempt := 1; attempt <= p.Config.MaxRetries; attempt++ {
		log.Printf("Repair Attempt %d/%d...", attempt, p.Config.MaxRetries)

		repairPrompt := fmt.Sprintf("%s\nTarget File: %s\nDiagnostic: %s\n%s\n%s",
			p.Prompts.RepairPrompt, failedFilePath, diagnosticError, fewShots, planResponse)

		patchResp, errRepair := p.Repairer.Run(repairPrompt, p.Config.ModelRouter["coding"])
		if errRepair != nil {
			log.Printf("Repair attempt %d failed to generate patch: %v", attempt, errRepair)
			continue
		}

		cleanPatch := extractPatchDiff(patchResp)
		if cleanPatch == "" {
			log.Printf("Repair attempt %d generated empty patch diff blocks", attempt)
			continue
		}

		// Apply patch
		log.Printf("Applying patch...")
		_, errApply := p.Git.Apply(cleanPatch)
		if errApply != nil {
			log.Printf("Patch application failed: %v", errApply)
			// Reset repository files
			_, _ = p.Git.Checkout(failedFilePath)
			continue
		}

		// Validate compiling, testing, and lint checks
		log.Printf("Validating patch against repository builds and tests...")
		errVal := p.Validators.VerifyAll()
		if errVal != nil {
			log.Printf("Patch validation failed: %v", errVal)
			// Revert the changes
			_, _ = p.Git.Checkout(failedFilePath)
			// Save rejected repair to history
			_ = p.Memory.SaveRepair(&memory.RepairRecord{
				AssociatedFailure: diagnosticError,
				AppliedPatch:      cleanPatch,
				DurationMs:        int(time.Since(startTime).Milliseconds()),
				TestsPassed:       false,
				Keywords:          extractKeywords(diagnosticError),
			}, false)
			continue
		}

		// Success!
		log.Printf("Patch successfully passed validation!")
		activePatch = cleanPatch
		success = true
		break
	}

	// Log metrics and finalize
	if !success {
		// Save failure history record
		_ = p.Memory.SaveFailure(&memory.FailureRecord{
			Category:   "compilation",
			Diagnostic: diagnosticError,
			StackTrace: "Max retries exceeded without patch verification.",
		})
		return fmt.Errorf("self-repair failed to resolve issue within %d retries", p.Config.MaxRetries)
	}

	// Save successful repair details
	_ = p.Memory.SaveRepair(&memory.RepairRecord{
		AssociatedFailure: diagnosticError,
		AppliedPatch:      activePatch,
		DurationMs:        int(time.Since(startTime).Milliseconds()),
		TestsPassed:       true,
		Keywords:          extractKeywords(diagnosticError),
	}, true)

	// Commit patch to branch
	commitMsg := fmt.Sprintf("phoenix: self-repair fix for issue in %s", failedFilePath)
	_, errCommit := p.Git.Commit(commitMsg)
	if errCommit != nil {
		log.Printf("Commit failed: %v", errCommit)
		return errCommit
	}
	log.Printf("Successfully committed self-repair patch: %s", commitMsg)
	return nil
}

func extractPatchDiff(response string) string {
	start := strings.Index(response, "diff --git")
	if start == -1 {
		return ""
	}
	end := strings.Index(response[start:], "```")
	if end == -1 {
		return strings.TrimSpace(response[start:])
	}
	return strings.TrimSpace(response[start : start+end])
}

func extractKeywords(diagnostic string) []string {
	words := strings.Fields(strings.ToLower(diagnostic))
	var kws []string
	for _, w := range words {
		if len(w) > 4 {
			wClean := strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
					return r
				}
				return -1
			}, w)
			if len(wClean) > 3 {
				kws = append(kws, wClean)
			}
		}
	}
	return kws
}
