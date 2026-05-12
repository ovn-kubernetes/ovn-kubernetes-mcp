package omcclient

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
)

var (
	// kubernetesNamePattern is the pattern for Kubernetes resource names.
	kubernetesNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9._-]*[a-z0-9])?$`)
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
	return utils.ValidateSafeString(selector, "label selector", true, utils.ShellMetaCharactersTypeAllowBrackets)
}
