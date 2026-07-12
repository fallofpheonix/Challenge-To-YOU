package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 || args[0] != "validate" {
		fmt.Println("Usage: challenge-validator validate [content_directory]")
		os.Exit(1)
	}

	targetDir := "challenges"
	if len(args) > 1 {
		targetDir = args[1]
	}

	log.Printf("Challenge Validator: Scanning folder '%s'...", targetDir)
	validator := NewChallengeValidator(targetDir)
	report, err := validator.Validate()
	if err != nil {
		log.Fatalf("Fatal: scan failed to complete: %v", err)
	}

	// Output JSON Report
	jsonData, errJSON := json.MarshalIndent(report, "", "  ")
	if errJSON != nil {
		log.Fatalf("Fatal: failed to marshal JSON report: %v", errJSON)
	}
	_ = os.WriteFile("challenge_validation.json", jsonData, 0644)

	// Output Markdown Report
	mdContent := renderMarkdownReport(report)
	_ = os.WriteFile("challenge_validation.md", []byte(mdContent), 0644)

	// Print summary statistics
	fmt.Printf("Validation Finished! Scanned Files: %d, Errors: %d, Warnings: %d\n",
		report.TotalScanned, len(report.Errors), len(report.Warnings))

	if !report.Valid {
		fmt.Println("Validation status: FAILED.")
		os.Exit(1)
	}
	fmt.Println("Validation status: PASSED.")
}

func renderMarkdownReport(r *ValidationReport) string {
	var sb strings.Builder
	sb.WriteString("# Challenge Verification Report\n\n")

	sb.WriteString(fmt.Sprintf("- **Total Scanned Templates**: %d\n", r.TotalScanned))
	if r.Valid {
		sb.WriteString("- **Overall Status**: :white_check_mark: **PASSED**\n\n")
	} else {
		sb.WriteString("- **Overall Status**: :x: **FAILED**\n\n")
	}

	sb.WriteString("## 🛑 Errors\n")
	if len(r.Errors) == 0 {
		sb.WriteString("✓ Zero schema errors detected!\n\n")
	} else {
		for _, err := range r.Errors {
			sb.WriteString(fmt.Sprintf("- %s\n", err))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## ⚠️ Warnings\n")
	if len(r.Warnings) == 0 {
		sb.WriteString("✓ Zero warnings detected!\n\n")
	} else {
		for _, warn := range r.Warnings {
			sb.WriteString(fmt.Sprintf("- %s\n", warn))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
