package main

import (
	"fmt"
	"strings"
)

type Agent string

const (
	AgentClaude Agent = "claude"
	AgentCodex  Agent = "codex"
)

func DefaultAgent() Agent {
	return AgentCodex
}

func ParseAgent(s string) (Agent, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "":
		return DefaultAgent(), nil
	case string(AgentClaude):
		return AgentClaude, nil
	case string(AgentCodex):
		return AgentCodex, nil
	default:
		return "", fmt.Errorf("unknown agent %q (expected claude or codex)", s)
	}
}

func (a Agent) String() string {
	if a == "" {
		return string(DefaultAgent())
	}
	return string(a)
}
