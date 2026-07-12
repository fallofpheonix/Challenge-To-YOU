package intelligence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRepositoryGraph(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "phoenix_graph_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockFile := filepath.Join(tempDir, "mock_code.go")
	code := `package mock

import "fmt"

type Config struct {
	OllamaHost string
}

func GetConfig() *Config {
	return &Config{}
}
`
	err = os.WriteFile(mockFile, []byte(code), 0644)
	if err != nil {
		t.Fatalf("failed to write mock file: %v", err)
	}

	graph := NewRepositoryGraph(tempDir)
	err = graph.Index()
	if err != nil {
		t.Fatalf("failed to index graph: %v", err)
	}

	node, exists := graph.Files[mockFile]
	if !exists {
		t.Fatalf("expected node to exist for %s", mockFile)
	}

	if len(node.Imports) != 1 || node.Imports[0] != "fmt" {
		t.Errorf("expected import [fmt], got %v", node.Imports)
	}

	var foundStruct, foundFunc bool
	for _, sym := range node.Symbols {
		if sym.Name == "Config" && sym.Type == "struct" {
			foundStruct = true
		}
		if sym.Name == "GetConfig" && sym.Type == "function" {
			foundFunc = true
		}
	}

	if !foundStruct {
		t.Error("expected to find struct symbol Config")
	}
	if !foundFunc {
		t.Error("expected to find function symbol GetConfig")
	}
}
