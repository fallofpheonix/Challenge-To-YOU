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
	if len(args) == 0 || args[0] != "scan" {
		fmt.Println("Usage: repo-doctor scan")
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Fatal: failed to get working directory: %v", err)
	}

	log.Println("Repository Doctor: Initiating codebase scan...")
	scanner := NewDoctorScanner(cwd)
	report, err := scanner.Scan()
	if err != nil {
		log.Fatalf("Fatal: scan failed: %v", err)
	}

	// Output JSON Report
	jsonData, errJSON := json.MarshalIndent(report, "", "  ")
	if errJSON != nil {
		log.Fatalf("Fatal: failed to marshal JSON report: %v", errJSON)
	}
	err = os.WriteFile("doctor_report.json", jsonData, 0644)
	if err != nil {
		log.Printf("Warning: failed to write JSON report file: %v", err)
	}

	// Output Markdown Report
	mdContent := renderMarkdownReport(report)
	err = os.WriteFile("doctor_report.md", []byte(mdContent), 0644)
	if err != nil {
		log.Printf("Warning: failed to write Markdown report file: %v", err)
	}

	fmt.Println("Scan complete! Generated doctor_report.json and doctor_report.md")
}

func renderMarkdownReport(r *ScanReport) string {
	var sb strings.Builder
	sb.WriteString("# Repository Doctor: Codebase Health Report\n\n")

	sb.WriteString("## 🧪 Test Coverage Check\n")
	if len(r.MissingTests) == 0 {
		sb.WriteString("✓ All source files have corresponding test suites!\n\n")
	} else {
		sb.WriteString("### Missing Test Files:\n")
		for _, f := range r.MissingTests {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## 📦 Package Documentation Check\n")
	if len(r.MissingDocDirs) == 0 {
		sb.WriteString("✓ All package folders contain README.md or doc.go files.\n\n")
	} else {
		sb.WriteString("### Undocumented Package Directories:\n")
		for _, d := range r.MissingDocDirs {
			sb.WriteString(fmt.Sprintf("- %s\n", d))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## ⚠️ TODOs & FIXMEs\n")
	if len(r.Todos) == 0 {
		sb.WriteString("✓ Zero pending TODO/FIXME markers found.\n\n")
	} else {
		sb.WriteString("### Found Comment Markers:\n")
		for _, todo := range r.Todos {
			sb.WriteString(fmt.Sprintf("- %s\n", todo))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## 📂 File Size Check\n")
	if len(r.LargeFiles) == 0 {
		sb.WriteString("✓ No large files detected.\n\n")
	} else {
		sb.WriteString("### Large Files (>50KB or >1000 lines):\n")
		for _, lf := range r.LargeFiles {
			sb.WriteString(fmt.Sprintf("- %s\n", lf))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## 🔑 Duplicate JSON Challenge IDs\n")
	if len(r.DuplicateIDs) == 0 {
		sb.WriteString("✓ No duplicate challenge IDs detected in level templates.\n\n")
	} else {
		sb.WriteString("### Duplicate IDs:\n")
		for _, dup := range r.DuplicateIDs {
			sb.WriteString(fmt.Sprintf("- %s\n", dup))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
