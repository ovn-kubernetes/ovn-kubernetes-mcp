package omcclient

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/must-gather/types"
)

// GetOvnKInfo uses the omc command to get the ovnk info. It will return an error if the
// info type is invalid. The valid info types are InfoTypeExtraInfo, InfoTypeHostNetInfo, and InfoTypeSubnets.
func (c *OmcClient) GetOvnKInfo(ctx context.Context, mustGatherPath string, infoType types.InfoType) (string, error) {
	var args []string
	// Set the arguments based on the info type
	switch infoType {
	case types.InfoTypeExtraInfo:
		args = append(args, "ovnk", "extrainfo")
	case types.InfoTypeHostNetInfo:
		args = append(args, "ovnk", "hostnetinfo")
	case types.InfoTypeSubnets:
		args = append(args, "ovnk", "subnets")
	default:
		return "", fmt.Errorf("invalid info type: %s", infoType)
	}

	// Get lock on the omc client
	c.lock.Lock()
	defer c.lock.Unlock()

	// Use the must gather path
	err := c.useMustGather(ctx, mustGatherPath)
	if err != nil {
		return "", fmt.Errorf("failed to use must gather: %w", err)
	}

	// Get the ovnk info from the omc command
	output, err := exec.CommandContext(ctx, c.commandPath, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get ovnk info. output: %s, error: %w", string(output), err)
	}
	return string(output), nil
}
