package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type ScanReport struct {
	MissingTests   []string `json:"missing_tests"`
	LargeFiles     []string `json:"large_files"`
	Todos          []string `json:"todos"`
	DuplicateIDs   []string `json:"duplicate_ids"`
	MissingDocDirs []string `json:"missing_doc_dirs"`
	BrokenImports  []string `json:"broken_imports"`
}

type DoctorScanner struct {
	WorkspaceRoot string
}

func NewDoctorScanner(root string) *DoctorScanner {
	return &DoctorScanner{WorkspaceRoot: root}
}

func (ds *DoctorScanner) Scan() (*ScanReport, error) {
	report := &ScanReport{
		MissingTests:   make([]string, 0),
		LargeFiles:     make([]string, 0),
		Todos:          make([]string, 0),
		DuplicateIDs:   make([]string, 0),
		MissingDocDirs: make([]string, 0),
		BrokenImports:  make([]string, 0),
	}

	seenIDs := make(map[string]string)
	hasTestMap := make(map[string]bool)
	var goFiles []string
	pkgDirs := make(map[string]bool)

	err := filepath.WalkDir(ds.WorkspaceRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "docs" || d.Name() == "brain" || d.Name() == "tools" || d.Name() == "phoenix" {
				return filepath.SkipDir
			}
			return nil
		}

		// Track test files
		name := d.Name()
		if strings.HasSuffix(name, "_test.go") {
			base := strings.TrimSuffix(name, "_test.go")
			hasTestMap[base] = true
		} else if strings.HasSuffix(name, ".go") {
			goFiles = append(goFiles, path)
			pkgDirs[filepath.Dir(path)] = true
		}

		// Check large files (> 50KB or > 1000 lines)
		info, errInfo := d.Info()
		if errInfo == nil && info.Size() > 50000 {
			report.LargeFiles = append(report.LargeFiles, fmt.Sprintf("%s (%d bytes)", path, info.Size()))
		}

		// Check TODOs/FIXMEs and line count
		if strings.HasSuffix(name, ".go") || strings.HasSuffix(name, ".gd") || strings.HasSuffix(name, ".gdshader") {
			ds.scanFileContents(path, report)
		}

		// Validate JSON challenge IDs
		if strings.HasSuffix(name, ".json") && strings.Contains(path, "challenges") {
			ds.scanChallengeID(path, seenIDs, report)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Verify missing test files
	for _, goFile := range goFiles {
		base := strings.TrimSuffix(filepath.Base(goFile), ".go")
		if !hasTestMap[base] {
			report.MissingTests = append(report.MissingTests, goFile)
		}
	}

	// Verify package documentation (must contain README.md or doc.go)
	for dir := range pkgDirs {
		hasDoc := false
		files, errRead := os.ReadDir(dir)
		if errRead == nil {
			for _, f := range files {
				name := strings.ToLower(f.Name())
				if name == "readme.md" || name == "doc.go" {
					hasDoc = true
					break
				}
			}
		}
		if !hasDoc {
			report.MissingDocDirs = append(report.MissingDocDirs, dir)
		}
	}

	return report, nil
}

func (ds *DoctorScanner) scanFileContents(path string, report *ScanReport) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.Contains(line, "TODO") || strings.Contains(line, "FIXME") {
			report.Todos = append(report.Todos, fmt.Sprintf("%s:%d: %s", path, lineNum, strings.TrimSpace(line)))
		}
	}

	if lineNum > 1000 {
		report.LargeFiles = append(report.LargeFiles, fmt.Sprintf("%s (>1000 lines: %d lines)", path, lineNum))
	}
}

func (ds *DoctorScanner) scanChallengeID(path string, seenIDs map[string]string, report *ScanReport) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var chal struct {
		ID string `json:"id"`
	}
	if errUnmarshal := json.Unmarshal(data, &chal); errUnmarshal == nil && chal.ID != "" {
		if originalPath, dup := seenIDs[chal.ID]; dup {
			report.DuplicateIDs = append(report.DuplicateIDs, fmt.Sprintf("ID %s duplicated in %s and %s", chal.ID, originalPath, path))
		} else {
			seenIDs[chal.ID] = path
		}
	}
}
