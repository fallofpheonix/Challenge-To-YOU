package engine

import (
	"path/filepath"
	"runtime"
	"testing"
)

// challengeFile resolves a challenge JSON path relative to this test file.
func challengeFile(rel string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "challenges", rel)
}

// TestWriteFromSpecChallenges verifies that every authored write_from_spec
// milestone challenge is solvable: its reference solution drives the win
// condition through the sandbox, and an incorrect solution is rejected.
func TestWriteFromSpecChallenges(t *testing.T) {
	cases := []struct {
		name    string
		file    string
		correct string
		wrong   string
	}{
		{
			name:    "magitech_m01_runic_initiation",
			file:    "magitech_tier1/magitech_m01_runic_initiation.json",
			correct: "var mana_current = 100;",
			wrong:   "var mana_current = 42;",
		},
		{
			name:    "cosmic_v81_ast_parser",
			file:    "cosmic_tier1/cosmic_v81_ast_parser.json",
			correct: "function buildAST(a){ if(a.length===0) return null; var m=Math.floor(a.length/2); return {v:a[m], left:buildAST(a.slice(0,m)), right:buildAST(a.slice(m+1))}; }",
			wrong:   "function buildAST(a){ return null; }",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def, err := LoadChallenge(challengeFile(tc.file))
			if err != nil {
				t.Fatalf("LoadChallenge failed: %v", err)
			}
			if def.SkillType != "write_from_spec" {
				t.Fatalf("expected skill_type write_from_spec, got %q", def.SkillType)
			}
			if def.ValidationScript == "" {
				t.Fatal("challenge has no validation_script")
			}

			// Correct solution must satisfy the win condition.
			state := make(map[string]interface{}, len(def.InitialState))
			for k, v := range def.InitialState {
				state[k] = v
			}
			res, err := def.ExecuteScript(tc.correct, state)
			if err != nil {
				t.Fatalf("reference solution rejected: %v", err)
			}
			got := res.MutatedState[def.WinCondition.TargetStateKey]
			if got != def.WinCondition.ExpectedValue {
				t.Errorf("win state %q = %v, want %v", def.WinCondition.TargetStateKey, got, def.WinCondition.ExpectedValue)
			}

			// Wrong solution must be rejected by the validation script.
			stateWrong := make(map[string]interface{}, len(def.InitialState))
			for k, v := range def.InitialState {
				stateWrong[k] = v
			}
			if _, err := def.ExecuteScript(tc.wrong, stateWrong); err == nil {
				t.Error("incorrect solution was accepted; validation_script must reject it")
			}
		})
	}
}
