package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// shellMetaCharacters is the pattern for shell metacharacters.
	shellMetaCharacters = regexp.MustCompile(`[;&|$` + "`" + `<>\\()]`)

	// shellMetaCharactersNoBrackets is like shellMetaCharacters but allows brackets.
	shellMetaCharactersNoBrackets = regexp.MustCompile(`[;&|$` + "`" + `<>\\]`)
)

// ShellMetaCharactersType is the type of shell metacharacters to validate.
type ShellMetaCharactersType string

// ShellMetaCharactersType values.
const (
	ShellMetaCharactersTypeDefault       ShellMetaCharactersType = "default"
	ShellMetaCharactersTypeAllowBrackets ShellMetaCharactersType = "allow_brackets"
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

// validateShellMetacharacters validates whether the given parameter contains shell
// metacharacters or not. It returns an error if there are any shell metacharacters.
// If shellMetaCharactersType is ShellMetaCharactersTypeDefault, no shell metacharacters are allowed.
// If shellMetaCharactersType is ShellMetaCharactersTypeAllowAmp, the & character is allowed.
// If shellMetaCharactersType is ShellMetaCharactersTypeAllowBrackets, the ( and ) characters are allowed.
func validateShellMetacharacters(param string, shellMetaCharactersType ShellMetaCharactersType) error {
	switch shellMetaCharactersType {
	case ShellMetaCharactersTypeAllowBrackets:
		if shellMetaCharactersNoBrackets.MatchString(param) {
			return fmt.Errorf("invalid use of metacharacters in parameter: %s", param)
		}
	case ShellMetaCharactersTypeDefault:
		if shellMetaCharacters.MatchString(param) {
			return fmt.Errorf("invalid use of metacharacters in parameter: %s", param)
		}
	default:
		return fmt.Errorf("invalid shell metacharacters type: %s", shellMetaCharactersType)
	}

	return nil
}

// ValidateSafeString is same as validateShellMetacharacters but it also checks if the string is empty.
// If allowEmpty is true, an empty string is allowed and no error is returned.
func ValidateSafeString(value, fieldName string, allowEmpty bool, shellMetaCharactersType ShellMetaCharactersType) error {
	if value == "" {
		if allowEmpty {
			return nil
		}
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	if err := validateShellMetacharacters(value, shellMetaCharactersType); err != nil {
		return fmt.Errorf("invalid %s: contains potentially dangerous characters: %w", fieldName, err)
	}
	return nil
}
