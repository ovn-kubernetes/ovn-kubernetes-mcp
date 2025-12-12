package omcclient

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var (
	// kubernetesNamePattern is the pattern for Kubernetes resource names.
	kubernetesNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9._-]*[a-z0-9])?$`)
	// shellMetacharacters is the pattern for shell metacharacters.
	shellMetacharacters = regexp.MustCompile(`[;&|$` + "`" + `<>\\()]`)
)

// getOmcCommandPath gets the path to the omc command. If the command is not found in the PATH,
// it will return an error.
func getOmcCommandPath() (string, error) {
	path, err := exec.LookPath("omc")
	if err != nil {
		return "", fmt.Errorf("failed to look path for omc: %w", err)
	}
	return path, nil
}

// validateKubernetesName validates Kubernetes resource names
func validateKubernetesName(name string, allowEmpty bool) error {
	if name == "" {
		if allowEmpty {
			return nil
		}
		return fmt.Errorf("field cannot be empty")
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("name cannot start with '-': %s", name)
	}
	if !kubernetesNamePattern.MatchString(strings.ToLower(name)) {
		return fmt.Errorf("invalid name format: %s", name)
	}
	return nil
}

// validateLabelSelector validates Kubernetes label selectors
func validateLabelSelector(selector string) error {
	if selector == "" {
		return nil
	}
	// Basic validation - no shell metacharacters
	if shellMetacharacters.MatchString(selector) {
		return fmt.Errorf("invalid characters in label selector: %s", selector)
	}
	return nil
}
