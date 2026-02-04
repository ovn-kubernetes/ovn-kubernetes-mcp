package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StripEmptyLines strips empty lines from a slice of strings. It will return a new slice of strings
// with the empty lines removed.
func StripEmptyLines(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}
	result := []string{}
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}

// GetGitRepositoryRoot returns the root directory of the git repository. It will return an error
// if the current directory is not a git repository or the root directory cannot be found.
func GetGitRepositoryRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		gitDirPath := filepath.Join(currentDir, ".git")
		if _, err := os.Stat(gitDirPath); err == nil {
			return currentDir, nil // Found .git directory, this is the root
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir { // Reached the file system root
			return "", fmt.Errorf("failed to find git repository root")
		}
		currentDir = parentDir
	}
}
