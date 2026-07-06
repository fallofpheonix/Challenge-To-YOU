// Package budget defines the execution limits enforced identically by every
// P-Script backend, so the bytecode VM and the tree-walk interpreter cannot
// diverge on well-behaved programs.
//
// The interpreter is a temporary verification oracle for the VM (see ADR-006).
// Both backends bound TOTAL work per drone per tick by MaxExecutionSteps — an
// aggregate budget, not a per-loop cap. Parity is guaranteed for any program
// that terminates within this budget (all realistic agent scripts). Behavior at
// the runaway safety cutoff may differ slightly because the VM meters bytecode
// instructions while the interpreter meters AST evaluations; that boundary
// difference is acceptable for an anti-infinite-loop safety net and is one
// reason the interpreter is slated for removal once P-Script 2.0 stabilizes.
//
// Exit criterion for the interpreter (ADR-006): remove it once P-Script 2.0 is
// feature-complete AND backend parity has held across releases AND fuzz/property
// tests provide equivalent confidence. Until then it guards the VM.
package budget

// MaxExecutionSteps caps the total execution steps one drone's program may take
// per tick, across all backends. Well-behaved scripts terminate far below it.
const MaxExecutionSteps = 1000
