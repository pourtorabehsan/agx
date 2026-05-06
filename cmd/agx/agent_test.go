package main

import "testing"

func TestParseAgent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Agent
		wantErr bool
	}{
		{name: "empty defaults to codex", input: "", want: AgentCodex},
		{name: "claude", input: "claude", want: AgentClaude},
		{name: "codex", input: "codex", want: AgentCodex},
		{name: "case insensitive", input: "Codex", want: AgentCodex},
		{name: "trim space", input: " codex ", want: AgentCodex},
		{name: "invalid", input: "aider", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAgent(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAgentStringDefaultsEmpty(t *testing.T) {
	var a Agent
	if got := a.String(); got != "codex" {
		t.Fatalf("got %q, want codex", got)
	}
}
