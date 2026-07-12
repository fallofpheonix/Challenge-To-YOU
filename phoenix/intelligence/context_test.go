package intelligence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestContextBuilder(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "phoenix_context_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockFile := filepath.Join(tempDir, "mock_code.go")
	code := `package mock
import "os"
func Run() {}`
	_ = os.WriteFile(mockFile, []byte(code), 0644)

	graph := NewRepositoryGraph(tempDir)
	_ = graph.Index()

	builder := NewContextBuilder(graph)
	ctx, err := builder.BuildContext(mockFile)
	if err != nil {
		t.Fatalf("failed to build context: %v", err)
	}

	if ctx.FailingFile != mockFile {
		t.Errorf("expected failing file %s, got %s", mockFile, ctx.FailingFile)
	}
	if !contains(ctx.Imports, "os") {
		t.Errorf("expected imports to contain 'os', got %v", ctx.Imports)
	}
}

func contains(arr []string, val string) bool {
	for _, a := range arr {
		if a == val {
			return true
		}
	}
	return false
}
