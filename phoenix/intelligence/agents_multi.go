package intelligence

import (
	"challenge-to-you/phoenix/agents"
	"fmt"
)

type AgentMessage struct {
	Sender      string
	Receiver    string
	MessageType string // "PLAN_REQ", "PLAN_RESP", "CODE_REQ", "CODE_RESP", "REVIEW_REQ", "REVIEW_RESP"
	Payload     string
}

type AgentCoordinator struct {
	Planner  agents.Agent
	Coder    agents.Agent
	Reviewer agents.Agent
}

func NewAgentCoordinator(p, c, r agents.Agent) *AgentCoordinator {
	return &AgentCoordinator{
		Planner:  p,
		Coder:    c,
		Reviewer: r,
	}
}

func (ac *AgentCoordinator) Dispatch(msg AgentMessage) (AgentMessage, error) {
	switch msg.MessageType {
	case "PLAN_REQ":
		resp, err := ac.Planner.Run(msg.Payload, "qwen2.5:32b-instruct")
		if err != nil {
			return AgentMessage{}, err
		}
		return AgentMessage{
			Sender:      "Planner",
			Receiver:    msg.Sender,
			MessageType: "PLAN_RESP",
			Payload:     resp,
		}, nil

	case "CODE_REQ":
		resp, err := ac.Coder.Run(msg.Payload, "deepseek-coder:16b-lite")
		if err != nil {
			return AgentMessage{}, err
		}
		return AgentMessage{
			Sender:      "Coder",
			Receiver:    msg.Sender,
			MessageType: "CODE_RESP",
			Payload:     resp,
		}, nil

	case "REVIEW_REQ":
		resp, err := ac.Reviewer.Run(msg.Payload, "qwen2.5:32b-instruct")
		if err != nil {
			return AgentMessage{}, err
		}
		return AgentMessage{
			Sender:      "Reviewer",
			Receiver:    msg.Sender,
			MessageType: "REVIEW_RESP",
			Payload:     resp,
		}, nil

	default:
		return AgentMessage{}, fmt.Errorf("unsupported message type: %s", msg.MessageType)
	}
}
