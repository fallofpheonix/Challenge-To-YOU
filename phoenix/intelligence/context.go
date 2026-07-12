package intelligence

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type RepairContext struct {
	FailingFile        string   `json:"failing_file"`
	FileContents       string   `json:"file_contents"`
	Imports            []string `json:"imports"`
	RelatedSymbols     []Symbol `json:"related_symbols"`
	RelatedCommitLogs  string   `json:"related_commit_logs"`
	DocumentationBrief string   `json:"documentation_brief"`
}

type ContextBuilder struct {
	Graph *RepositoryGraph
}

func NewContextBuilder(g *RepositoryGraph) *ContextBuilder {
	return &ContextBuilder{Graph: g}
}

func (cb *ContextBuilder) BuildContext(failingFile string) (*RepairContext, error) {
	cb.Graph.Mu.RLock()
	defer cb.Graph.Mu.RUnlock()

	node, exists := cb.Graph.Files[failingFile]
	var imports []string
	var symbols []Symbol
	if exists {
		imports = node.Imports
		symbols = node.Symbols
	}

	data, err := os.ReadFile(failingFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read failing file: %w", err)
	}

	// Retrieve documentation brief
	docBrief := ""
	docPath := filepath.Join(cb.Graph.WorkspaceRoot, "docs/DECISIONS")
	files, errRead := os.ReadDir(docPath)
	if errRead == nil && len(files) > 0 {
		docFile := filepath.Join(docPath, files[0].Name())
		docData, errDoc := os.ReadFile(docFile)
		if errDoc == nil {
			lines := strings.Split(string(docData), "\n")
			limit := len(lines)
			if limit > 20 {
				limit = 20
			}
			docBrief = strings.Join(lines[:limit], "\n")
		}
	}

	return &RepairContext{
		FailingFile:        failingFile,
		FileContents:       string(data),
		Imports:            imports,
		RelatedSymbols:     symbols,
		RelatedCommitLogs:  "phoenix: auto tracking workspace logs",
		DocumentationBrief: docBrief,
	}, nil
}
