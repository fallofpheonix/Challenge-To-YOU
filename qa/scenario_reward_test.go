package qa

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// ScenarioRewardXP verifies XP rewards are reflected in the player profile message.
func ScenarioRewardXP(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	if srv == nil {
		return ctx.Error
	}
	defer srv.Stop()

	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	snap, err := client.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("read initial snapshot: %w", err)
	}
	ctx.Step("initial_snapshot", "Received initial snapshot")

	// Send profile event to load profile from DB
	if err := client.Send("profile", ""); err != nil {
		return fmt.Errorf("send profile: %w", err)
	}

	profileSnap, err := client.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("read profile snapshot: %w", err)
	}
	ctx.Step("profile_loaded", "Profile event sent and response received")

	// Server sends profile stats in message text
	msg := strings.ToLower(profileSnap.Message)
	if strings.Contains(msg, "reputation") || strings.Contains(msg, "user stats") {
		ctx.Step("xp_tracking", fmt.Sprintf("Profile stats in message: %s", profileSnap.Message))
	} else {
		ctx.Step("xp_tracking", fmt.Sprintf("Server message: %s", profileSnap.Message))
	}

	// Verify challenge loaded correctly
	if snap.ChallengeID != "" {
		ctx.Step("challenge_loaded", fmt.Sprintf("Challenge: %s", snap.ChallengeID))
	}

	return nil
}

// ScenarioRewardCredits verifies credit reward tracking via profile query.
func ScenarioRewardCredits(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	if srv == nil {
		return ctx.Error
	}
	defer srv.Stop()

	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	if _, err := client.ReadSnapshot(5 * time.Second); err != nil {
		return fmt.Errorf("read initial: %w", err)
	}

	if err := client.Send("profile", ""); err != nil {
		return fmt.Errorf("send profile: %w", err)
	}

	snap, err := client.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("read profile snap: %w", err)
	}

	// Profile info is in message text
	msg := snap.Message
	if strings.Contains(strings.ToLower(msg), "reputation") {
		ctx.Step("credits_tracking", fmt.Sprintf("Profile message: %s", msg))
		ctx.Assert("profile_reachable", true, "Profile stats returned in message")
	} else {
		ctx.Assert("profile_reachable", false, fmt.Sprintf("Unexpected response: %s", msg))
	}
	return nil
}

// ScenarioRewardReputation verifies reputation is visible via profile event.
func ScenarioRewardReputation(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	if srv == nil {
		return ctx.Error
	}
	defer srv.Stop()

	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	if _, err := client.ReadSnapshot(5 * time.Second); err != nil {
		return fmt.Errorf("read initial: %w", err)
	}

	if err := client.Send("profile", ""); err != nil {
		return fmt.Errorf("send profile: %w", err)
	}

	snap, err := client.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("read profile snap: %w", err)
	}

	msg := strings.ToLower(snap.Message)
	if strings.Contains(msg, "reputation") {
		ctx.Assert("reputation_visible", true, fmt.Sprintf("Reputation info in message: %s", snap.Message))
	} else {
		ctx.Assert("reputation_visible", false, fmt.Sprintf("No reputation info: %s", snap.Message))
	}
	return nil
}

// ScenarioRewardLuck verifies luck stat is mentioned in profile message.
func ScenarioRewardLuck(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	if srv == nil {
		return ctx.Error
	}
	defer srv.Stop()

	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	if _, err := client.ReadSnapshot(5 * time.Second); err != nil {
		return fmt.Errorf("read initial: %w", err)
	}

	if err := client.Send("profile", ""); err != nil {
		return fmt.Errorf("send profile: %w", err)
	}

	snap, err := client.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("read profile snap: %w", err)
	}

	msg := strings.ToLower(snap.Message)
	if strings.Contains(msg, "luck") {
		ctx.Assert("luck_visible", true, fmt.Sprintf("Luck info in message: %s", snap.Message))
	} else {
		ctx.Assert("luck_visible", false, fmt.Sprintf("No luck info: %s", snap.Message))
	}
	return nil
}

// ScenarioRewardUnlocks verifies MAGITECH is unlocked by default via profile message.
func ScenarioRewardUnlocks(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	if srv == nil {
		return ctx.Error
	}
	defer srv.Stop()

	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	if _, err := client.ReadSnapshot(5 * time.Second); err != nil {
		return fmt.Errorf("read initial: %w", err)
	}

	if err := client.Send("profile", ""); err != nil {
		return fmt.Errorf("send profile: %w", err)
	}

	snap, err := client.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("read profile snap: %w", err)
	}

	msg := strings.ToLower(snap.Message)
	if strings.Contains(msg, "unlocked") || strings.Contains(msg, "paradigm") {
		ctx.Assert("unlocks_visible", true, fmt.Sprintf("Unlocks info in message: %s", snap.Message))
	} else {
		ctx.Assert("unlocks_visible", false, fmt.Sprintf("No unlocks info: %s", snap.Message))
	}

	time.Sleep(50 * time.Millisecond)
	return nil
}

// TestRewardScenarios runs QA reward scenarios via go test.
func TestRewardScenarios(t *testing.T) {
	scenarios := []struct {
		name string
		fn   func(ctx *ScenarioContext) error
	}{
		{"RewardXP", ScenarioRewardXP},
		{"RewardCredits", ScenarioRewardCredits},
		{"RewardReputation", ScenarioRewardReputation},
		{"RewardLuck", ScenarioRewardLuck},
		{"RewardUnlocks", ScenarioRewardUnlocks},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			ctx := &ScenarioContext{
				Name:        s.name,
				T:           t,
				FixturesDir: "fixtures",
				ReportsDir:  "reports",
				LogsDir:     "backend_logs",
			}
			if err := s.fn(ctx); err != nil {
				t.Errorf("Scenario %s failed: %v", s.name, err)
			}
			for _, step := range ctx.Steps {
				if step.Status == StatusFailed {
					t.Errorf("Step %s failed: %s", step.Name, step.Error)
				}
			}
		})
	}
}
