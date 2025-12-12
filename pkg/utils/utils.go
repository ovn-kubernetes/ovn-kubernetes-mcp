package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

// MatchPattern matches a pattern to a slice of strings. It will return a new slice of strings
// with the matched lines. It will return an error if the pattern is invalid.
func MatchPattern(pattern string, lines []string) ([]string, error) {
	if len(lines) == 0 {
		return lines, nil
	}
	searchPattern, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid search pattern: %w", err)
	}
	matchedLines := []string{}
	for _, line := range lines {
		if searchPattern.MatchString(line) {
			matchedLines = append(matchedLines, line)
		}
	}
	return matchedLines, nil
}

// Head returns the first n lines of a slice of strings. It will return a new slice of strings
// with the first n lines. If n is less than or equal to 0, or greater than or equal to the
// length of the slice, it will return the entire slice.
func Head(lines []string, n int) []string {
	if len(lines) == 0 {
		return lines
	}
	if n <= 0 || n >= len(lines) {
		return lines
	}
	return lines[:n]
}

// Tail returns the last n lines of a slice of strings. It will return a new slice of strings
// with the last n lines. If n is less than or equal to 0, or greater than or equal to the
// length of the slice, it will return the entire slice.
func Tail(lines []string, n int) []string {
	if len(lines) == 0 {
		return lines
	}
	if n <= 0 || n >= len(lines) {
		return lines
	}
	return lines[len(lines)-n:]
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
