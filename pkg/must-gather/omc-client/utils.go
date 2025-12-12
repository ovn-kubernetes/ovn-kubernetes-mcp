package omcclient

import (
	"fmt"
	"os/exec"
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
