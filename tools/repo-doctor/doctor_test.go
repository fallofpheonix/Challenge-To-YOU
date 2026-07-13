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

func TestDoctorScannerSkipsExternalReferenceDirs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repo_doctor_skip_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	for _, dir := range []string{"refs/vendor_pkg", "research/vendor_pkg"} {
		fullDir := filepath.Join(tempDir, dir)
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			t.Fatalf("failed to create %s: %v", fullDir, err)
		}

		source := "package vendor\n// TODO: ignored external reference\nfunc Run() {}\n"
		if err := os.WriteFile(filepath.Join(fullDir, "vendor.go"), []byte(source), 0644); err != nil {
			t.Fatalf("failed to write fixture: %v", err)
		}
	}

	scanner := NewDoctorScanner(tempDir)
	report, errScan := scanner.Scan()
	if errScan != nil {
		t.Fatalf("Scan failed: %v", errScan)
	}

	if len(report.MissingTests) != 0 {
		t.Errorf("expected skipped refs/research to produce no missing tests, got %v", report.MissingTests)
	}
	if len(report.MissingDocDirs) != 0 {
		t.Errorf("expected skipped refs/research to produce no missing docs, got %v", report.MissingDocDirs)
	}
	if len(report.Todos) != 0 {
		t.Errorf("expected skipped refs/research to produce no TODOs, got %v", report.Todos)
	}
}

func TestDoctorScannerTreatsTestsAsPackageLevel(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repo_doctor_package_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pkgDir := filepath.Join(tempDir, "pkg")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatalf("failed to create package dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(pkgDir, "alpha.go"), []byte("package pkg\nfunc Alpha() {}\n"), 0644); err != nil {
		t.Fatalf("failed to write alpha.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "beta.go"), []byte("package pkg\nfunc Beta() {}\n"), 0644); err != nil {
		t.Fatalf("failed to write beta.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "pkg_test.go"), []byte("package pkg\nimport \"testing\"\nfunc TestAlpha(t *testing.T) {}\n"), 0644); err != nil {
		t.Fatalf("failed to write package test: %v", err)
	}

	scanner := NewDoctorScanner(tempDir)
	report, errScan := scanner.Scan()
	if errScan != nil {
		t.Fatalf("Scan failed: %v", errScan)
	}

	if len(report.MissingTests) != 0 {
		t.Errorf("expected package-level test to cover package, got %v", report.MissingTests)
	}
}

func TestDoctorScannerSkipsGeneratedArtifacts(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repo_doctor_artifact_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	largeData := strings.Repeat("x", 50001)
	for _, name := range []string{"doctor_report.md", "doctor_report.json", "sandbox", "sandbox_server_test"} {
		if err := os.WriteFile(filepath.Join(tempDir, name), []byte(largeData), 0644); err != nil {
			t.Fatalf("failed to write artifact %s: %v", name, err)
		}
	}

	scanner := NewDoctorScanner(tempDir)
	report, errScan := scanner.Scan()
	if errScan != nil {
		t.Fatalf("Scan failed: %v", errScan)
	}

	if len(report.LargeFiles) != 0 {
		t.Errorf("expected generated artifacts to be skipped, got %v", report.LargeFiles)
	}
}

func TestDoctorScannerDoesNotRequireTestsForDocGo(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repo_doctor_docgo_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pkgDir := filepath.Join(tempDir, "pkg")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatalf("failed to create package dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "doc.go"), []byte("// Package pkg documents a package.\npackage pkg\n"), 0644); err != nil {
		t.Fatalf("failed to write doc.go: %v", err)
	}

	scanner := NewDoctorScanner(tempDir)
	report, errScan := scanner.Scan()
	if errScan != nil {
		t.Fatalf("Scan failed: %v", errScan)
	}

	if len(report.MissingTests) != 0 {
		t.Errorf("expected doc.go to be ignored for missing tests, got %v", report.MissingTests)
	}
	if len(report.MissingDocDirs) != 0 {
		t.Errorf("expected doc.go to satisfy package docs, got %v", report.MissingDocDirs)
	}
}
