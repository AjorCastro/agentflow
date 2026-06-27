package protocol_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AjorCastro/agentflow/internal/protocol"
)

func TestNewInitialState(t *testing.T) {
	s := protocol.NewInitialState(
		"feature-test", "Test Feature",
		"develop", "feature/test",
		".worktrees/test", ".agentflow/features/feature-test",
	)

	if s.Protocol != "agentflow" {
		t.Errorf("Protocol = %q, want agentflow", s.Protocol)
	}
	if s.Version != "0.1" {
		t.Errorf("Version = %q, want 0.1", s.Version)
	}
	if s.CurrentPhase != "intake" {
		t.Errorf("CurrentPhase = %q, want intake", s.CurrentPhase)
	}
	if s.CurrentTurn != "human" {
		t.Errorf("CurrentTurn = %q, want human", s.CurrentTurn)
	}
	if s.Status != "initialized" {
		t.Errorf("Status = %q, want initialized", s.Status)
	}
	if !s.AttentionRequired {
		t.Error("AttentionRequired should be true")
	}
	if s.LastResult != nil {
		t.Error("LastResult should be nil")
	}
	if s.NextAction != "run_initial_cli_skill" {
		t.Errorf("NextAction = %q, want run_initial_cli_skill", s.NextAction)
	}
	if s.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}
	if s.UpdatedAt == "" {
		t.Error("UpdatedAt should not be empty")
	}
}

func TestWriteAndReadStateFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	original := protocol.NewInitialState(
		"feature-write-test", "Write Test",
		"develop", "feature/write-test",
		".worktrees/write-test", ".agentflow/features/feature-write-test",
	)

	if err := protocol.WriteStateFile(path, original); err != nil {
		t.Fatalf("WriteStateFile: %v", err)
	}

	// Verify the file is valid JSON.
	data, _ := os.ReadFile(path)
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("state.json is not valid JSON: %v", err)
	}

	// Round-trip read.
	read, err := protocol.ReadStateFile(path)
	if err != nil {
		t.Fatalf("ReadStateFile: %v", err)
	}

	if read.FeatureID != original.FeatureID {
		t.Errorf("FeatureID mismatch: got %q, want %q", read.FeatureID, original.FeatureID)
	}
	if read.Protocol != original.Protocol {
		t.Errorf("Protocol mismatch: got %q, want %q", read.Protocol, original.Protocol)
	}
}
