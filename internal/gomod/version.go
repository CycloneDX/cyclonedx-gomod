package gomod

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GetPseudoVersion constructs a pseudo version for a Go module at the given path.
// Note that this is only possible when path points to a Git repository and the
// git binary is available in the system's PATH.
// See https://golang.org/ref/mod#pseudo-versions
func GetPseudoVersion(path string) (string, error) {
	cmd := exec.Command("git", "show", "-s", "--format=%H %cI")
	cmd.Dir = path

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("executing command '%s' failed: %w", cmd.String(), err)
	}

	parts := strings.Fields(string(output))
	if len(parts) != 2 {
		return "", fmt.Errorf("expected two fields in git output, but got %d: %s", len(parts), output)
	}

	commitHash := parts[0][:12]
	commitDate, err := time.Parse(time.RFC3339, parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to parse commit timestamp: %w", err)
	}

	return fmt.Sprintf("v0.0.0-%s-%s", commitDate.Format("20060102150405"), commitHash), nil
}

// GetVersionFromTag checks if the current commit is annotated with a tag and if yes, returns that tag's name.
// Note that this is only possible when path points to a Git repository and the
// git binary is available in the system's PATH.
func GetVersionFromTag(path string) (string, error) {
	cmd := exec.Command("git", "describe", "--exact-match", "--tags", "HEAD")
	cmd.Dir = path

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("executing command '%s' failed: %w", cmd.String(), err)
	}

	return strings.TrimSpace(string(output)), nil
}
