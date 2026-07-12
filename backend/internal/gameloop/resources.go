package gameloop

import (
	"fmt"
	"sync"
)

type ResourceOperation struct {
	PlayerID     string       `json:"player_id"`
	Type         string       `json:"type"`
	ResourceType ResourceType `json:"resource_type"`
	Amount       int          `json:"amount"`
	Target       string       `json:"target,omitempty"`
}

type ResourceResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Removed int    `json:"removed,omitempty"`
	Added   int    `json:"added,omitempty"`
}

type ResourceManager struct {
	mu          sync.RWMutex
	inventories map[string]*Inventory
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		inventories: make(map[string]*Inventory),
	}
}

func (rm *ResourceManager) GetInventory(playerID string) *Inventory {
	rm.mu.RLock()
	inv, ok := rm.inventories[playerID]
	rm.mu.RUnlock()
	if ok {
		return inv
	}
	inv = NewInventory()
	rm.mu.Lock()
	rm.inventories[playerID] = inv
	rm.mu.Unlock()
	return inv
}

func (rm *ResourceManager) EnsureInventory(playerID string) *Inventory {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	inv, ok := rm.inventories[playerID]
	if !ok {
		inv = NewInventory()
		rm.inventories[playerID] = inv
	}
	return inv
}

func (rm *ResourceManager) Gather(playerID string, r ResourceType, amount int) ResourceResult {
	inv := rm.EnsureInventory(playerID)
	inv.Add(r, amount)
	return ResourceResult{
		Success: true,
		Message: fmt.Sprintf("Gathered %d %s", amount, string(r)),
		Added:   amount,
	}
}

func (rm *ResourceManager) Consume(playerID string, r ResourceType, amount int) ResourceResult {
	inv := rm.GetInventory(playerID)
	if !inv.Has(r, amount) {
		return ResourceResult{
			Success: false,
			Message: fmt.Sprintf("Not enough %s: have %d, need %d", string(r), inv.Count(r), amount),
		}
	}
	removed := inv.Remove(r, amount)
	return ResourceResult{
		Success: true,
		Message: fmt.Sprintf("Consumed %d %s", removed, string(r)),
		Removed: removed,
	}
}

func (rm *ResourceManager) Deliver(playerID string, r ResourceType, amount int, target string) ResourceResult {
	inv := rm.GetInventory(playerID)
	if !inv.Has(r, amount) {
		return ResourceResult{
			Success: false,
			Message: fmt.Sprintf("Not enough %s to deliver to %s", string(r), target),
		}
	}
	removed := inv.Remove(r, amount)
	return ResourceResult{
		Success: true,
		Message: fmt.Sprintf("Delivered %d %s to %s", removed, string(r), target),
		Removed: removed,
	}
}

func (rm *ResourceManager) ProcessOperation(op ResourceOperation) ResourceResult {
	switch op.Type {
	case "gather":
		if op.Amount <= 0 {
			op.Amount = 1
		}
		return rm.Gather(op.PlayerID, op.ResourceType, op.Amount)
	case "consume":
		return rm.Consume(op.PlayerID, op.ResourceType, op.Amount)
	case "deliver":
		return rm.Deliver(op.PlayerID, op.ResourceType, op.Amount, op.Target)
	default:
		return ResourceResult{
			Success: false,
			Message: fmt.Sprintf("Unknown resource operation: %s", op.Type),
		}
	}
}

func (rm *ResourceManager) Reset(playerID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.inventories, playerID)
}
