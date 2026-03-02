package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// ApplyTimeout returns a context with a timeout if the duration is positive.
// When timeout is zero or negative, the original context is returned with a no-op cancel function.
// Always call the returned cancel function (typically via defer).
func ApplyTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	}
	return ctx, func() {}
}

// WrapTimeoutError checks if an error is a timeout/deadline error and wraps it with a more
// informative message for the LLM, including the timeout duration and suggestions.
func WrapTimeoutError(err error, operation string, timeout time.Duration) error {
	if err == nil {
		return nil
	}

	// Check if this is a timeout/deadline exceeded error
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%s: timed out after %v. Possible causes: remote system slow/unresponsive, large dataset, or operation inherently slow. "+
			"Solutions: use filters to reduce data size. Error: %w", operation, timeout, err)
	}

	// Check if this is a canceled context (user/system initiated cancellation)
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("%s: operation canceled: %w",
			operation, err)
	}

	return err
}
