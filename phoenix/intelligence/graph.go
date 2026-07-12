package intelligence

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Symbol struct {
	Name      string `json:"name"`
	Type      string `json:"type"` // "struct", "interface", "function"
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type FileNode struct {
	Path         string    `json:"path"`
	LastModified time.Time `json:"last_modified"`
	Size         int64     `json:"size"`
	Imports      []string  `json:"imports"`
	Symbols      []Symbol  `json:"symbols"`
}

type RepositoryGraph struct {
	Mu            sync.RWMutex
	WorkspaceRoot string
	Files         map[string]*FileNode
}

func NewRepositoryGraph(root string) *RepositoryGraph {
	return &RepositoryGraph{
		WorkspaceRoot: root,
		Files:         make(map[string]*FileNode),
	}
}

func (g *RepositoryGraph) Index() error {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	return filepath.Walk(g.WorkspaceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "docs" || info.Name() == "brain" || info.Name() == "phoenix" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		// Cache check
		cached, exists := g.Files[path]
		if exists && cached.LastModified.Equal(info.ModTime()) && cached.Size == info.Size() {
			return nil
		}

		node, errParse := parseGoFile(path, info)
		if errParse != nil {
			// Skip files with parse errors to prevent index failures
			return nil
		}

		g.Files[path] = node
		return nil
	})
}

func parseGoFile(path string, info os.FileInfo) (*FileNode, error) {
	fset := token.NewFileSet()
	fileAST, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	node := &FileNode{
		Path:         path,
		LastModified: info.ModTime(),
		Size:         info.Size(),
		Imports:      make([]string, 0),
		Symbols:      make([]Symbol, 0),
	}

	for _, imp := range fileAST.Imports {
		if imp.Path != nil {
			node.Imports = append(node.Imports, strings.Trim(imp.Path.Value, `"`))
		}
	}

	ast.Inspect(fileAST, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			sym := Symbol{
				Name:      x.Name.Name,
				StartLine: fset.Position(x.Pos()).Line,
				EndLine:   fset.Position(x.End()).Line,
			}
			switch x.Type.(type) {
			case *ast.StructType:
				sym.Type = "struct"
			case *ast.InterfaceType:
				sym.Type = "interface"
			default:
				sym.Type = "type"
			}
			node.Symbols = append(node.Symbols, sym)

		case *ast.FuncDecl:
			sym := Symbol{
				Name:      x.Name.Name,
				Type:      "function",
				StartLine: fset.Position(x.Pos()).Line,
				EndLine:   fset.Position(x.End()).Line,
			}
			node.Symbols = append(node.Symbols, sym)
		}
		return true
	})

	return node, nil
}
