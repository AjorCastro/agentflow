package protocol

import (
	"encoding/json"
	"os"
	"time"
)

// State is the parsed representation of state.json.
type State struct {
	Protocol        string  `json:"protocol"`
	Version         string  `json:"version"`
	FeatureID       string  `json:"feature_id"`
	Title           string  `json:"title"`
	RootBranch      string  `json:"root_branch"`
	FeatureBranch   string  `json:"feature_branch"`
	Worktree        string  `json:"worktree"`
	ExchangeFolder  string  `json:"exchange_folder"`
	CurrentPhase    string  `json:"current_phase"`
	CurrentTurn     string  `json:"current_turn"`
	Status          string  `json:"status"`
	AttentionRequired bool  `json:"attention_required"`
	LastResult      *string `json:"last_result"`
	NextAction      string  `json:"next_action"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// NewInitialState returns a State populated for a fresh AgentFlow workspace.
func NewInitialState(featureID, title, root, branch, worktree, exchange string) State {
	now := time.Now().UTC().Format(time.RFC3339)
	return State{
		Protocol:          "agentflow",
		Version:           "0.1",
		FeatureID:         featureID,
		Title:             title,
		RootBranch:        root,
		FeatureBranch:     branch,
		Worktree:          worktree,
		ExchangeFolder:    exchange,
		CurrentPhase:      "intake",
		CurrentTurn:       "human",
		Status:            "initialized",
		AttentionRequired: true,
		LastResult:        nil,
		NextAction:        "run_initial_cli_skill",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// WriteStateFile marshals state to a JSON file at path.
func WriteStateFile(path string, s State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ReadStateFile reads and parses a state.json file.
func ReadStateFile(path string) (State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return State{}, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}, err
	}
	return s, nil
}
