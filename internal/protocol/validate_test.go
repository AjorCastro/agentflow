package protocol_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflow/agentflow/internal/protocol"
)

func TestFeatureIDFromBranch(t *testing.T) {
	cases := []struct {
		branch string
		want   string
	}{
		{"feature/test-agentflow", "feature-test-agentflow"},
		{"feature/my-feature", "feature-my-feature"},
		{"main", "main"},
		{"release/v1.0.0", "release-v1.0.0"},
		{"fix/some/nested/path", "fix-some-nested-path"},
	}

	for _, tc := range cases {
		got := protocol.FeatureIDFromBranch(tc.branch)
		if got != tc.want {
			t.Errorf("FeatureIDFromBranch(%q) = %q, want %q", tc.branch, got, tc.want)
		}
	}
}

func TestDefaultExchangePath(t *testing.T) {
	got := protocol.DefaultExchangePath("/repo", "feature/my-feature")
	want := filepath.Join("/repo", ".agentflow", "features", "feature-my-feature")
	if got != want {
		t.Errorf("DefaultExchangePath = %q, want %q", got, want)
	}
}

func TestValidateExchange_Valid(t *testing.T) {
	dir := t.TempDir()

	for _, sub := range protocol.RequiredDirs {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range protocol.RequiredFiles {
		if err := os.WriteFile(filepath.Join(dir, f), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	result := protocol.ValidateExchange(dir)
	if !result.Valid {
		t.Errorf("expected valid, got missing: %v", result.Missing)
	}
}

func TestValidateExchange_MissingFiles(t *testing.T) {
	dir := t.TempDir()
	// No files or dirs created.

	result := protocol.ValidateExchange(dir)
	if result.Valid {
		t.Error("expected invalid, got valid")
	}
	if len(result.Missing) == 0 {
		t.Error("expected missing items, got none")
	}
}

func TestValidateExchange_MissingOneDir(t *testing.T) {
	dir := t.TempDir()

	// Create all required dirs except "specs".
	for _, sub := range protocol.RequiredDirs {
		if sub == "specs" {
			continue
		}
		_ = os.MkdirAll(filepath.Join(dir, sub), 0755)
	}
	for _, f := range protocol.RequiredFiles {
		_ = os.WriteFile(filepath.Join(dir, f), []byte(""), 0644)
	}

	result := protocol.ValidateExchange(dir)
	if result.Valid {
		t.Error("expected invalid due to missing specs/")
	}
	found := false
	for _, m := range result.Missing {
		if m == "specs/" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'specs/' in missing, got %v", result.Missing)
	}
}
