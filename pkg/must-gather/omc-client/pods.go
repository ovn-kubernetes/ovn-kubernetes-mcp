package omcclient

import (
	"context"
	"fmt"
	"os/exec"
)

// GetPodLogs uses the omc command to get the logs of a pod. It will return an error if the
// must gather path is not valid or the pod logs are not found.
func (c *OmcClient) GetPodLogs(ctx context.Context, mustGatherPath string, namespace, name, container string, previous bool, rotated bool) (string, error) {
	// Validate the name
	if err := validateKubernetesName(name, false); err != nil {
		return "", fmt.Errorf("failed to validate name: %w", err)
	}
	// Validate the namespace
	if err := validateKubernetesName(namespace, true); err != nil {
		return "", fmt.Errorf("failed to validate namespace: %w", err)
	}
	// Validate the container
	if err := validateKubernetesName(container, true); err != nil {
		return "", fmt.Errorf("failed to validate container: %w", err)
	}

	// Get lock on the omc client
	c.lock.Lock()
	defer c.lock.Unlock()

	// Use the must gather path
	err := c.useMustGather(ctx, mustGatherPath)
	if err != nil {
		return "", fmt.Errorf("failed to use must gather: %w", err)
	}

	// Set the arguments based on the pod logs
	args := []string{}
	// Append the name argument
	args = append(args, "logs", name)
	// Append the container argument
	if container != "" {
		args = append(args, "-c", container)
	}
	// Append the namespace argument
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	// Append the previous argument
	if previous {
		args = append(args, "-p")
	}
	// Append the rotated argument
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
