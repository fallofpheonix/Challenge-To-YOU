package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"challenge-to-you/backend/internal/engine"
)

func (s *Server) handleListLanguages(w http.ResponseWriter, r *http.Request) {
	langs := s.compilerManager.ListLanguages()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(langs)
}

func (s *Server) handleListChallenges(w http.ResponseWriter, r *http.Request) {
	dataDir := "challenges"
	ids, err := listChallengeIDs(dataDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"total":      len(ids),
		"challenges": ids,
	})
}

func listChallengeIDs(dataDir string) ([]string, error) {
	var ids []string
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subDir := filepath.Join(dataDir, entry.Name())
		subEntries, err := os.ReadDir(subDir)
		if err != nil {
			continue
		}
		for _, subEntry := range subEntries {
			if subEntry.IsDir() || !strings.HasSuffix(subEntry.Name(), ".json") || subEntry.Name() == "pack.json" {
				continue
			}
			path := filepath.Join(subDir, subEntry.Name())
			def, err := engine.LoadChallenge(path)
			if err != nil {
				continue
			}
			ids = append(ids, def.ID)
		}
	}
	return ids, nil
}
