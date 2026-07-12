package intelligence

import (
	"strings"
)

type CodebaseReport struct {
	TotalFiles           int     `json:"total_files"`
	AverageComplexity    float64 `json:"average_complexity"`
	CouplingRatio        float64 `json:"coupling_ratio"`
	TechnicalDebtScore   int     `json:"technical_debt_score"`
	CleanCodeHealthScore int     `json:"clean_code_health_score"`
}

type HealthDashboard struct {
	Graph *RepositoryGraph
}

func NewHealthDashboard(g *RepositoryGraph) *HealthDashboard {
	return &HealthDashboard{Graph: g}
}

func (hd *HealthDashboard) GenerateReport() *CodebaseReport {
	hd.Graph.Mu.RLock()
	defer hd.Graph.Mu.RUnlock()

	totalFiles := len(hd.Graph.Files)
	if totalFiles == 0 {
		return &CodebaseReport{
			CleanCodeHealthScore: 100,
		}
	}

	totalComplexity := 0
	internalImports := 0
	externalImports := 0

	for _, fileNode := range hd.Graph.Files {
		// Calculate cyclomatic complexity markers
		for _, sym := range fileNode.Symbols {
			if sym.Type == "function" {
				// Base complexity is 1
				totalComplexity += 1
			}
		}

		for _, imp := range fileNode.Imports {
			if strings.Contains(imp, "challenge-to-you") {
				internalImports++
			} else {
				externalImports++
			}
		}
	}

	coupling := 0.0
	if externalImports > 0 {
		coupling = float64(internalImports) / float64(externalImports)
	}

	avgComplexity := float64(totalComplexity) / float64(totalFiles)
	if avgComplexity == 0 {
		avgComplexity = 1.0
	}

	// Calculate a health score out of 100
	health := 100 - int(avgComplexity*2.5) - int(coupling*5.0)
	if health < 20 {
		health = 20
	}

	return &CodebaseReport{
		TotalFiles:           totalFiles,
		AverageComplexity:    avgComplexity,
		CouplingRatio:        coupling,
		TechnicalDebtScore:   int(avgComplexity * 10),
		CleanCodeHealthScore: health,
	}
}
