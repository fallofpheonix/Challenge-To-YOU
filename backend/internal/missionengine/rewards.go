package missionengine

import (
	"challenge-to-you/backend/internal/eventbus"
)

// RewardDispatcher handles distributing rewards to players
type RewardDispatcher struct {
	bus *eventbus.EventBus
}

// NewRewardDispatcher creates a new reward dispatcher
func NewRewardDispatcher(bus *eventbus.EventBus) *RewardDispatcher {
	return &RewardDispatcher{bus: bus}
}

// DispatchReward sends rewards to a player
func (rd *RewardDispatcher) DispatchReward(playerID string, reward *MissionReward) {
	if reward == nil {
		return
	}

	// Dispatch XP
	if reward.XP > 0 {
		rd.bus.Publish(eventbus.Event{
			Type: eventbus.EventXPEarned,
			Payload: map[string]interface{}{
				"player_id": playerID,
				"xp":        reward.XP,
				"source":    "mission_reward",
			},
			Source: "reward_dispatcher",
		})
	}

	// Dispatch achievements
	for _, achievement := range reward.Achievements {
		rd.bus.Publish(eventbus.Event{
			Type: eventbus.EventAchievementUnlocked,
			Payload: map[string]interface{}{
				"player_id":   playerID,
				"achievement": achievement,
				"source":      "mission_reward",
			},
			Source: "reward_dispatcher",
		})
	}

	// Dispatch unlocks
	for _, unlock := range reward.Unlocks {
		rd.bus.Publish(eventbus.Event{
			Type: "unlock_granted",
			Payload: map[string]interface{}{
				"player_id": playerID,
				"unlock":    unlock,
				"source":    "mission_reward",
			},
			Source: "reward_dispatcher",
		})
	}
}

// UnlockPropagation manages unlocking new content based on completion
type UnlockPropagation struct {
	bus *eventbus.EventBus
}

// NewUnlockPropagation creates a new unlock propagation system
func NewUnlockPropagation(bus *eventbus.EventBus) *UnlockPropagation {
	return &UnlockPropagation{bus: bus}
}

// PropagateUnlocks checks if completing a mission unlocks new content
func (up *UnlockPropagation) PropagateUnlocks(playerID string, completedMissionID string, registry *MissionRegistry) {
	// Find missions that are now unlockable
	for _, mission := range registry.GetAll() {
		// Check if this mission requires the completed one
		for _, req := range mission.Requirements {
			if req == completedMissionID {
				// Check if all other requirements are met
				// (This is simplified - in production, check against player's full completion history)
				up.bus.Publish(eventbus.Event{
					Type: "mission_unlocked",
					Payload: map[string]interface{}{
						"player_id":   playerID,
						"mission_id":  mission.ID,
						"unlocked_by": completedMissionID,
					},
					Source: "unlock_propagation",
				})
			}
		}
	}
}
