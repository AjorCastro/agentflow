package protocol_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AjorCastro/agentflow/internal/protocol"
)

func TestDefaultLocalPath(t *testing.T) {
	got := protocol.DefaultLocalPath("/repo", "my-feature")
	want := filepath.Join("/repo", ".agentflow", "local", "my-feature")
	if got != want {
		t.Errorf("DefaultLocalPath = %q, want %q", got, want)
	}
}

func TestValidateLocal_Valid(t *testing.T) {
	dir := t.TempDir()

	for _, sub := range protocol.LocalRequiredDirs {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range protocol.LocalRequiredFiles {
		if err := os.WriteFile(filepath.Join(dir, f), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	result := protocol.ValidateLocal(dir)
	if !result.Valid {
		t.Errorf("expected valid, got missing: %v", result.Missing)
	}
}

func TestValidateLocal_MissingEverything(t *testing.T) {
	dir := t.TempDir()

	result := protocol.ValidateLocal(dir)
	if result.Valid {
		t.Error("expected invalid, got valid")
	}
	if len(result.Missing) != len(protocol.LocalRequiredDirs)+len(protocol.LocalRequiredFiles) {
		t.Errorf("expected all dirs/files missing, got %v", result.Missing)
	}
}

func TestFindLocalFolders(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".agentflow", "local", "feature-a"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".agentflow", "local", "feature-b"), 0755); err != nil {
		t.Fatal(err)
	}

	dirs, err := protocol.FindLocalFolders(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 2 {
		t.Errorf("expected 2 local folders, got %d: %v", len(dirs), dirs)
	}
}

func TestPolicyMD_ContainsFeatureMetadata(t *testing.T) {
	content := protocol.PolicyMD("my-feature", "My Feature", "develop", "feature/my-feature", ".worktrees/my-feature")
	for _, want := range []string{"my-feature", "My Feature", "develop", "feature/my-feature", ".worktrees/my-feature"} {
		if !strings.Contains(content, want) {
			t.Errorf("PolicyMD output missing %q", want)
		}
	}
}
