package simulation

import "chrysalis-engine/core/crysmath"

// emitSpawn records a drone coming into existence (initial seeding or fabrication).
func (e *Engine) emitSpawn(droneID, x, y, swarmSize int) {
	e.Bus.Emit(Event{
		TickNum: e.Tick,
		Type:    EventDroneSpawned,
		DroneID: int32(droneID),
		X:       int32(x),
		Y:       int32(y),
		Data:    SpawnedData{SwarmSize: swarmSize},
	})
}

// emitFabricated records a drone created by the fabrication pool.
func (e *Engine) emitFabricated(droneID, x, y, swarmSize int) {
	e.Bus.Emit(Event{
		TickNum: e.Tick,
		Type:    EventFabricated,
		DroneID: int32(droneID),
		X:       int32(x),
		Y:       int32(y),
		Data:    SpawnedData{SwarmSize: swarmSize},
	})
}

// emitDeath records a drone becoming inert.
// cause must be "battery" or "hazard".
func (e *Engine) emitDeath(droneID int, cause string) {
	x := int32(e.Registry.PositionX[droneID].V / crysmath.Precision)
	y := int32(e.Registry.PositionY[droneID].V / crysmath.Precision)
	e.Bus.Emit(Event{
		TickNum: e.Tick,
		Type:    EventDroneDied,
		DroneID: int32(droneID),
		X:       x,
		Y:       y,
		Data:    DiedData{Cause: cause},
	})
}

// emitHarvest records a successful resource collection.
func (e *Engine) emitHarvest(droneID, x, y int, resourcesRemaining int32) {
	e.Bus.Emit(Event{
		TickNum: e.Tick,
		Type:    EventHarvested,
		DroneID: int32(droneID),
		X:       int32(x),
		Y:       int32(y),
		Data:    HarvestData{ResourcesRemaining: resourcesRemaining},
	})
}

// emitDeposit records a resource being dropped at the colony base.
func (e *Engine) emitDeposit(droneID, x, y int, amount, colonyTotal int32) {
	e.Bus.Emit(Event{
		TickNum: e.Tick,
		Type:    EventDeposited,
		DroneID: int32(droneID),
		X:       int32(x),
		Y:       int32(y),
		Data:    DepositData{Amount: amount, ColonyTotal: colonyTotal},
	})
}

// emitInfection records a drone becoming compromised and the accompanying trust drop.
// vector must be "alien_node" or "peer_spread".
func (e *Engine) emitInfection(droneID int, x, y int32, vector string, oldTrust int32) {
	e.Bus.Emit(Event{
		TickNum: e.Tick,
		Type:    EventDroneInfected,
		DroneID: int32(droneID),
		X:       x,
		Y:       y,
		Data:    InfectedData{Vector: vector},
	})
	e.Bus.Emit(Event{
		TickNum: e.Tick,
		Type:    EventTrustChanged,
		DroneID: int32(droneID),
		X:       x,
		Y:       y,
		Data:    TrustData{OldTrust: oldTrust, NewTrust: 50},
	})
}

// emitHazardDamage records battery drain from a hazard field.
func (e *Engine) emitHazardDamage(droneID int, damage, batteryRemaining int64) {
	x := int32(e.Registry.PositionX[droneID].V / crysmath.Precision)
	y := int32(e.Registry.PositionY[droneID].V / crysmath.Precision)
	e.Bus.Emit(Event{
		TickNum: e.Tick,
		Type:    EventHazardDamage,
		DroneID: int32(droneID),
		X:       x,
		Y:       y,
		Data:    HazardData{Damage: damage, BatteryRemaining: batteryRemaining},
	})
}

// emitMissionChanged records a mission state transition.
func (e *Engine) emitMissionChanged(status, reason string) {
	e.Bus.Emit(Event{
		TickNum: e.Tick,
		Type:    EventMissionChanged,
		DroneID: -1,
		Data:    MissionData{Status: status, Reason: reason},
	})
}
