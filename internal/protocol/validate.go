package protocol

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RequiredDirs lists subdirectories that must exist under an exchange folder.
var RequiredDirs = []string{
	"specs", "discovery", "plans", "tasks",
	"decisions", "reviews", "handoffs", "prompts",
}

// RequiredFiles lists files that must exist under an exchange folder.
var RequiredFiles = []string{
	"README.md", "STATUS.md", "CONFIG.md", "state.json",
}

// ValidationResult holds the outcome of validating a single exchange folder.
type ValidationResult struct {
	ExchangeFolder string
	Missing        []string
	Valid          bool
}

// ValidateExchange checks that all required files and dirs exist under exchangeDir.
func ValidateExchange(exchangeDir string) ValidationResult {
	result := ValidationResult{ExchangeFolder: exchangeDir, Valid: true}

	for _, f := range RequiredFiles {
		path := filepath.Join(exchangeDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			result.Missing = append(result.Missing, f)
			result.Valid = false
		}
	}

	for _, d := range RequiredDirs {
		path := filepath.Join(exchangeDir, d)
		if info, err := os.Stat(path); os.IsNotExist(err) || (err == nil && !info.IsDir()) {
			result.Missing = append(result.Missing, d+"/")
			result.Valid = false
		}
	}

	return result
}

// FindExchangeFolders returns all directories matching .agentflow/features/*/ under root.
func FindExchangeFolders(root string) ([]string, error) {
	pattern := filepath.Join(root, ".agentflow", "features", "*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob error: %w", err)
	}

	var dirs []string
	for _, m := range matches {
		info, err := os.Stat(m)
		if err == nil && info.IsDir() {
			dirs = append(dirs, m)
		}
	}
	return dirs, nil
}

// FeatureIDFromBranch converts a branch name to a feature ID for use under
// .agentflow/features/<feature-id>. A leading "feature/" prefix is dropped
// before conversion, since the "features" directory already conveys that —
// keeping it would produce a redundant ".../features/feature-<name>" path.
// Any other remaining '/' is replaced with '-'.
func FeatureIDFromBranch(branch string) string {
	branch = strings.TrimPrefix(branch, "feature/")

	result := make([]byte, len(branch))
	for i := range branch {
		if branch[i] == '/' {
			result[i] = '-'
		} else {
			result[i] = branch[i]
		}
	}
	return string(result)
}

// DefaultExchangePath returns the default exchange folder path for a branch.
func DefaultExchangePath(repoDir, branch string) string {
	featureID := FeatureIDFromBranch(branch)
	return filepath.Join(repoDir, ".agentflow", "features", featureID)
}
