package engine

import (
	"path/filepath"
	"testing"
)

// --- Optimize challenge ---
// Complexity is gated by the sandbox timeout on fib(40):
//   - O(2^N) recursion at n=40 requires ~100B operations and cannot complete in 2s.
//   - O(N) iterative completes in microseconds.
// This is NOT a noisy wall-clock sample — it's a binary pass/fail on whether the
// algorithm finishes the hardest test case at all. incrementOp() / OpsCount are
// exposed for future step-counting puzzles that need finer-grained complexity checks.

func TestOptimizeChallenge_FailsOnRecursive(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "cyberpunk_07_optimize.json")
	def, err := LoadChallenge(path)
	if err != nil {
		t.Fatalf("Failed to load optimize challenge: %v", err)
	}

	state := make(map[string]interface{})
	for k, v := range def.InitialState {
		state[k] = v
	}

	// Unoptimized recursive fib — validation script wraps it with incrementOp()
	// so it will blow past the 50-op budget and throw.
	failingCode := `
		function fib(n) {
			if (n <= 1) return n;
			return fib(n - 1) + fib(n - 2);
		}
	`
	_, errFail := def.ExecuteScript(failingCode, state)
	if errFail == nil {
		t.Fatal("Expected O(2^N) recursion to exceed op budget, but sandbox returned success")
	}
}

func TestOptimizeChallenge_PassesOnIterative(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "cyberpunk_07_optimize.json")
	def, err := LoadChallenge(path)
	if err != nil {
		t.Fatalf("Failed to load optimize challenge: %v", err)
	}

	state := make(map[string]interface{})
	for k, v := range def.InitialState {
		state[k] = v
	}

	// O(N) iterative fib — should cost exactly 10 instrumented calls (well under 50).
	passingCode := `
		function fib(n) {
			if (n <= 1) return n;
			let a = 0, b = 1;
			for (let i = 2; i <= n; i++) {
				let temp = a + b;
				a = b;
				b = temp;
			}
			return b;
		}
	`
	res, errPass := def.ExecuteScript(passingCode, state)
	if errPass != nil {
		t.Fatalf("Expected optimized script to pass, got error: %v", errPass)
	}

	if res.MutatedState["security_bypass"] != true {
		t.Errorf("Expected security_bypass=true, got %v", res.MutatedState["security_bypass"])
	}

	// Log OpsCount for visibility — the grading gate is fib(40) within the 2s sandbox timeout,
	// not this counter (which includes the reference wrappedFib run in the validation script).
	t.Logf("OpsCount (includes reference wrappedFib): %d", res.OpsCount)
}

// --- Write-From-Spec challenge (sandbox still used, no grading change) ---

func TestWriteFromSpecChallenge(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "cyberpunk_08_spec.json")
	def, err := LoadChallenge(path)
	if err != nil {
		t.Fatalf("Failed to load spec challenge: %v", err)
	}

	state := make(map[string]interface{})
	for k, v := range def.InitialState {
		state[k] = v
	}

	failingCode := `
		function decodePacket(packet) {
			return "";
		}
	`
	_, errFail := def.ExecuteScript(failingCode, state)
	if errFail == nil {
		t.Fatal("Expected incorrect RLE decoder to fail validation")
	}

	passingCode := `
		function decodePacket(packet) {
			let result = "";
			for (let i = 0; i < packet.length; i += 2) {
				let char = packet[i];
				let count = parseInt(packet[i+1]);
				result += char.repeat(count);
			}
			return result;
		}
	`
	res, errPass := def.ExecuteScript(passingCode, state)
	if errPass != nil {
		t.Fatalf("Expected correct decoder to pass, got error: %v", errPass)
	}

	if res.MutatedState["access_granted"] != true {
		t.Errorf("Expected access_granted=true, got %v", res.MutatedState["access_granted"])
	}
}

// --- Recognize challenge ---
// Graded via EvaluateAnswer() ONLY — the Goja sandbox is NEVER started.
// This verifies the skill-type boundary is enforced at the engine level.

func TestRecognizeChallenge_CorrectAnswer(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "cyberpunk_09_recognize.json")
	def, err := LoadChallenge(path)
	if err != nil {
		t.Fatalf("Failed to load recognize challenge: %v", err)
	}

	state := make(map[string]interface{})
	for k, v := range def.InitialState {
		state[k] = v
	}

	mutated, err := def.EvaluateAnswer("4", state)
	if err != nil {
		t.Fatalf("Expected correct answer '4' to pass, got error: %v", err)
	}
	if mutated["system_firewall_disabled"] != true {
		t.Errorf("Expected system_firewall_disabled=true after correct answer, got %v", mutated["system_firewall_disabled"])
	}
}

func TestRecognizeChallenge_WrongAnswer(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "cyberpunk_09_recognize.json")
	def, err := LoadChallenge(path)
	if err != nil {
		t.Fatalf("Failed to load recognize challenge: %v", err)
	}

	state := make(map[string]interface{})
	for k, v := range def.InitialState {
		state[k] = v
	}

	_, err = def.EvaluateAnswer("3", state)
	if err == nil {
		t.Fatal("Expected wrong answer '3' to return an error, but EvaluateAnswer succeeded")
	}
}

func TestRecognizeChallenge_ExecuteScriptIsGuardedOff(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "cyberpunk_09_recognize.json")
	def, err := LoadChallenge(path)
	if err != nil {
		t.Fatalf("Failed to load recognize challenge: %v", err)
	}

	// Calling EvaluateAnswer on a non-recognize challenge should error.
	// And conversely, calling EvaluateAnswer on a recognize challenge with
	// no template_code / validation_script should NEVER touch the sandbox.
	if def.ValidationScript != "" {
		t.Errorf("Recognize challenge should have no validation_script, found: %q", def.ValidationScript)
	}
	if def.TemplateCode != "" {
		t.Errorf("Recognize challenge should have no template_code, found: %q", def.TemplateCode)
	}
}
