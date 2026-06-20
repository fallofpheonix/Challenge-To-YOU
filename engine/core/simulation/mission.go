package simulation

const (
	DefaultMissionTargetResources  int     = 10
	DefaultMissionMaxTicks         int64   = 3000
	DefaultMissionInfectionLoss    float64 = 0.50
	MissionReasonResourceTarget    string  = "resource_target_reached"
	MissionReasonInfectionExceeded string  = "infection_threshold_exceeded"
	MissionReasonTickLimit         string  = "tick_limit_exceeded"
)

type MissionStatus string

const (
	MissionRunning MissionStatus = "running"
	MissionVictory MissionStatus = "victory"
	MissionDefeat  MissionStatus = "defeat"
)

type MissionState struct {
	Status                 MissionStatus `json:"status"`
	Reason                 string        `json:"reason,omitempty"`
	TargetResources        int           `json:"target_resources"`
	MaxTicks               int64         `json:"max_ticks"`
	InfectionLossThreshold float64       `json:"infection_loss_threshold"`
	ResourcesDeposited     int           `json:"resources_deposited"`
	InfectedRatio          float64       `json:"infected_ratio"`
	Tick                   int64         `json:"tick"`
}

func NewDefaultMissionState() MissionState {
	return MissionState{
		Status:                 MissionRunning,
		TargetResources:        DefaultMissionTargetResources,
		MaxTicks:               DefaultMissionMaxTicks,
		InfectionLossThreshold: DefaultMissionInfectionLoss,
	}
}

func (e *Engine) EvaluateMission() {
	if e.Mission.Status != MissionRunning {
		return
	}

	e.Mission.Tick = e.Tick
	e.Mission.ResourcesDeposited = int(e.TotalDeposited)
	e.Mission.InfectedRatio = e.InfectedRatio()

	switch {
	case e.Mission.ResourcesDeposited >= e.Mission.TargetResources:
		e.Mission.Status = MissionVictory
		e.Mission.Reason = MissionReasonResourceTarget
	case e.Tick > 0 && e.Mission.InfectedRatio >= e.Mission.InfectionLossThreshold:
		e.Mission.Status = MissionDefeat
		e.Mission.Reason = MissionReasonInfectionExceeded
	case e.Tick >= e.Mission.MaxTicks:
		e.Mission.Status = MissionDefeat
		e.Mission.Reason = MissionReasonTickLimit
	}
}

func (e *Engine) InfectedRatio() float64 {
	if e.Registry.Count == 0 {
		return 0
	}

	infected := 0
	for i := 0; i < e.Registry.Count; i++ {
		if e.Registry.Compromised[i] {
			infected++
		}
	}
	return float64(infected) / float64(e.Registry.Count)
}
