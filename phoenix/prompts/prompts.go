package prompts

import (
	"os"
	"path/filepath"
	"strings"
)

type PromptLibrary struct {
	PlannerPrompt string
	RepairPrompt  string
	ReviewPrompt  string
}

func LoadPrompts(workspaceRoot string) *PromptLibrary {
	defaultLib := &PromptLibrary{
		PlannerPrompt: "You are the Lead Project Planner. Analyze evidence and produce a list of modification steps.",
		RepairPrompt:  "You are the Self-Repair Specialist. Generate a minimal unified git diff patch block to fix the error.",
		ReviewPrompt:  "You are the Code Quality Auditor. Review the proposed patch and output APPROVED or REJECTED.",
	}

	promptsPath := filepath.Join(workspaceRoot, "docs/phoenix/prompts.md")
	data, err := os.ReadFile(promptsPath)
	if err != nil {
		return defaultLib
	}

	content := string(data)
	sections := strings.Split(content, "## ")
	for _, sec := range sections {
		lines := strings.Split(sec, "\n")
		if len(lines) == 0 {
			continue
		}
		title := strings.TrimSpace(lines[0])
		body := strings.Join(lines[1:], "\n")

		// Extract content within block quotes or code blocks
		codeBlock := extractCodeBlock(body)
		if codeBlock == "" {
			continue
		}

		if strings.Contains(title, "Planner") {
			defaultLib.PlannerPrompt = codeBlock
		} else if strings.Contains(title, "Repairer") {
			defaultLib.RepairPrompt = codeBlock
		} else if strings.Contains(title, "Reviewer") {
			defaultLib.ReviewPrompt = codeBlock
		}
	}

	return defaultLib
}

func extractCodeBlock(text string) string {
	start := strings.Index(text, "```")
	if start == -1 {
		return ""
	}
	text = text[start+3:]
	// Skip language specifier if any
	firstNewLine := strings.Index(text, "\n")
	if firstNewLine != -1 {
		text = text[firstNewLine+1:]
	}
	end := strings.Index(text, "```")
	if end == -1 {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(text[:end])
}
