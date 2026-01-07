package omcclient

import (
	"context"
	"fmt"
	"os/exec"
)

// GetPodLogs uses the omc command to get the logs of a pod. It will return an error if the
// must gather path is not valid or the pod logs are not found.
func (c *OmcClient) GetPodLogs(ctx context.Context, mustGatherPath string, namespace, name, container string, previous bool, rotated bool) (string, error) {
	// Use the must gather path
	err := c.useMustGather(ctx, mustGatherPath)
	if err != nil {
		return "", fmt.Errorf("failed to use must gather: %w", err)
	}

	// Set the arguments based on the pod logs
	args := []string{"logs", name}
	if container != "" {
		args = append(args, "-c", container)
	}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	if previous {
		args = append(args, "-p")
	}
	if rotated {
		args = append(args, "-r")
	}

	// Get the pod logs from the omc command
	output, err := exec.CommandContext(ctx, c.commandPath, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs. output: %s, error: %w", string(output), err)
	}
	return string(output), nil
}
