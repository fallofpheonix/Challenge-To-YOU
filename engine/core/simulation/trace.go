package simulation

// DecisionStep is a single record within a drone's policy execution for one tick.
// Kind is "condition" or "action". Result is always a human-readable string.
type DecisionStep struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Result string `json:"result"`
	Taken  bool   `json:"taken,omitempty"` // only meaningful for "condition" steps
}

// DecisionFrame is a completed, immutable record of every decision a drone made
// during one tick of P-Script execution. Produced by the interpreter after Eval;
// the builder that produced it is not exposed outside the interpreter package.
type DecisionFrame struct {
	DroneID int            `json:"drone_id"`
	Tick    int64          `json:"tick"`
	Steps   []DecisionStep `json:"steps"`
}
