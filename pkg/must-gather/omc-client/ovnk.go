package omcclient

import (
	"fmt"
	"os/exec"
)

// GetOvnKInfo uses the omc command to get the ovnk info. It will return an error if the
// info type is invalid.
func (c *OmcClient) GetOvnKInfo(mustGatherPath string, infoType string) (string, error) {
	var args []string
	// Set the arguments based on the info type
	switch infoType {
	case "extrainfo":
		args = append(args, "ovnk", "extrainfo")
	case "hostnetinfo":
		args = append(args, "ovnk", "hostnetinfo")
	case "subnets":
		args = append(args, "ovnk", "subnets")
	default:
		return "", fmt.Errorf("invalid info type: %s", infoType)
	}

	// Use the must gather path
	err := c.useMustGather(mustGatherPath)
	if err != nil {
		return "", fmt.Errorf("failed to use must gather: %w", err)
	}

	// Get the ovnk info from the omc command
	output, err := exec.Command(c.CommandPath, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get ovnk info. output: %s, error: %w", string(output), err)
	}
	return string(output), nil
}
