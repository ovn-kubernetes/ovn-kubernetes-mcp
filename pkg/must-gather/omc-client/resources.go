package omcclient

import (
	"fmt"
	"os/exec"

	k8sTypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
)

// GetResource uses the omc command to get a resource. It will return an error if the
// must gather path is not valid or the resource is not found.
func (c *OmcClient) GetResource(mustGatherPath string, kind string, namespace, name string, outputType k8sTypes.OutputType) (string, error) {
	// Use the must gather path
	err := c.useMustGather(mustGatherPath)
	if err != nil {
		return "", fmt.Errorf("failed to use must gather: %w", err)
	}

	// Set the arguments based on the resource
	args := []string{"get", kind, name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	// Set the output type
	switch outputType {
	case k8sTypes.JSONOutputType:
		args = append(args, "-o", "json")
	case k8sTypes.YAMLOutputType:
		args = append(args, "-o", "yaml")
	case k8sTypes.WideOutputType:
		args = append(args, "-o", "wide")
	}

	// Get the resource from the omc command
	output, err := exec.Command(c.CommandPath, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get resource. output: %s, error: %w", string(output), err)
	}
	return string(output), nil
}

// ListResources uses the omc command to list resources. It will return an error if the
// must gather path is not valid or the resources are not found.
func (c *OmcClient) ListResources(mustGatherPath string, kind string, namespace string, labelSelector string, outputType k8sTypes.OutputType) (string, error) {
	// Use the must gather path
	err := c.useMustGather(mustGatherPath)
	if err != nil {
		return "", fmt.Errorf("failed to use must gather: %w", err)
	}

	// Set the arguments based on the resource
	args := []string{"get", kind}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "-A")
	}
	if labelSelector != "" {
		args = append(args, "-l", labelSelector)
	}

	// Set the output type
	switch outputType {
	case k8sTypes.JSONOutputType:
		args = append(args, "-o", "json")
	case k8sTypes.YAMLOutputType:
		args = append(args, "-o", "yaml")
	case k8sTypes.WideOutputType:
		args = append(args, "-o", "wide")
	}

	// Get the resources from the omc command
	output, err := exec.Command(c.CommandPath, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to list resources. output: %s, error: %w", string(output), err)
	}
	return string(output), nil
}
