package intelligence

import (
	"fmt"
	"strings"
)

type DependencySelector struct {
	Graph *RepositoryGraph
}

func NewDependencySelector(g *RepositoryGraph) *DependencySelector {
	return &DependencySelector{Graph: g}
}

func (ds *DependencySelector) SelectTargetTests(modifiedFile string) ([]string, error) {
	ds.Graph.Mu.RLock()
	defer ds.Graph.Mu.RUnlock()

	targetTests := []string{"go test -run=TestDummy"} // fallback default

	node, exists := ds.Graph.Files[modifiedFile]
	if !exists {
		return []string{"go test ./..."}, nil
	}

	// Simple heuristic: find tests in the same package/directory
	dir := strings.Replace(modifiedFile, "/"+node.Path, "", -1)
	targetTests = append(targetTests, fmt.Sprintf("go test %s", dir))

	// Find structural references
	for path, fileNode := range ds.Graph.Files {
		if strings.HasSuffix(path, "_test.go") {
			for _, imp := range fileNode.Imports {
				if strings.Contains(modifiedFile, imp) {
					targetTests = append(targetTests, fmt.Sprintf("go test %s", path))
				}
			}
		}
	}

	return targetTests, nil
}
