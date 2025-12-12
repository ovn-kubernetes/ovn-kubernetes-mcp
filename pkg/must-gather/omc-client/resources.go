package omcclient

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	k8sTypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetResource uses the omc command to get a resource. It will return an error if the
// must gather path is not valid or the resource is not found.
func (c *OmcClient) GetResource(ctx context.Context, mustGatherPath string, kind string, namespace, name string, outputType k8sTypes.OutputType) (string, error) {
	// Use the must gather path
	err := c.useMustGather(ctx, mustGatherPath)
	if err != nil {
		return "", fmt.Errorf("failed to use must gather: %w", err)
	}

	// Set the arguments based on the resource
	args := []string{"get", kind, name}
	if namespace == "" {
		namespace = metav1.NamespaceDefault
	}
	args = append(args, "-n", namespace)
	// Append the output type arguments
	args = appendOutputTypeArgs(args, outputType)

	// Get the resource from the omc command
	output, err := exec.CommandContext(ctx, c.commandPath, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get resource. output: %s, error: %w", string(output), err)
	}

	// If the output contains "No resources", return an error.
	if strings.Contains(strings.ToLower(string(output)), "no resources") {
		return "", fmt.Errorf("failed to get resource %s with name %s in namespace %s: %s", kind, name, namespace, string(output))
	}
	return string(output), nil
}

// ListResources uses the omc command to list resources. It will return an error if the
// must gather path is not valid.
func (c *OmcClient) ListResources(ctx context.Context, mustGatherPath string, kind string, namespace string, labelSelector string, outputType k8sTypes.OutputType) (string, error) {
	// Use the must gather path
	err := c.useMustGather(ctx, mustGatherPath)
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

	// Append the output type arguments
	args = appendOutputTypeArgs(args, outputType)

	// Get the resources from the omc command
	output, err := exec.CommandContext(ctx, c.commandPath, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to list resources. output: %s, error: %w", string(output), err)
	}

	// If the output contains "No resources", return an empty string.
	if strings.Contains(strings.ToLower(string(output)), "no resources") {
		return "", nil
	}
	return string(output), nil
}

// appendOutputTypeArgs appends the output type arguments to the arguments.
func appendOutputTypeArgs(args []string, outputType k8sTypes.OutputType) []string {
	switch outputType {
	case k8sTypes.JSONOutputType:
		return append(args, "-o", "json")
	case k8sTypes.YAMLOutputType:
		return append(args, "-o", "yaml")
	case k8sTypes.WideOutputType:
		return append(args, "-o", "wide")
	}
	return args
}
