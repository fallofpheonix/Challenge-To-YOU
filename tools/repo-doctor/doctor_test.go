package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctorScanner(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repo_doctor_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create sub-package directory and go file
	subPkg := filepath.Join(tempDir, "pkg1")
	_ = os.MkdirAll(subPkg, 0755)

	goCode := `package pkg1
// TODO: implement this logic
func Run() {}
`
	_ = os.WriteFile(filepath.Join(subPkg, "pkg1.go"), []byte(goCode), 0644)

	// Create JSON challenge template containing ID
	chalDir := filepath.Join(tempDir, "challenges")
	_ = os.MkdirAll(chalDir, 0755)
	validChal := `{"id": "M-01"}`
	_ = os.WriteFile(filepath.Join(chalDir, "chal1.json"), []byte(validChal), 0644)

	scanner := NewDoctorScanner(tempDir)
	report, errScan := scanner.Scan()
	if errScan != nil {
		t.Fatalf("Scan failed: %v", errScan)
	}

	// Verify missing test detection
	if len(report.MissingTests) != 1 || !strings.Contains(report.MissingTests[0], "pkg1.go") {
		t.Errorf("expected 1 missing test for pkg1.go, got %v", report.MissingTests)
	}

	// Verify TODO detection
	if len(report.Todos) != 1 || !strings.Contains(report.Todos[0], "TODO: implement this logic") {
		t.Errorf("expected 1 TODO comment, got %v", report.Todos)
	}

	// Verify undocumented package detection
	if len(report.MissingDocDirs) != 1 || !strings.Contains(report.MissingDocDirs[0], "pkg1") {
		t.Errorf("expected missing documentation in pkg1, got %v", report.MissingDocDirs)
	}
}
