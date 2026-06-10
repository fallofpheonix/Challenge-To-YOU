// Package state: Defines the shared game state used by the core and the interpreter.
// This structure is the "Source of Truth" for the simulation.
package state

type Status string

const (
	StatusActive   Status = "active"
	StatusComplete Status = "complete"
)

type Vector2 struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type GameState struct {
	Tick   int64   `json:"tick"`
	Agent  Vector2 `json:"agent"`
	Goal   Vector2 `json:"goal"`
	Status Status  `json:"status"`
}
