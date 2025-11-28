package external_tools

import (
	"context"
	"fmt"
	k8stypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/types"
)

var interfaceNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

// executeCommand runs a command on a node
func (s *MCPServer) runDebugNode(ctx context.Context, req *mcp.CallToolRequest, target types.DebugNodeParams, command []string) (*mcp.CallToolResult, types.CommandResult, error) {
	if target.NodeName == "" {
		return nil, types.CommandResult{}, fmt.Errorf("node_name is required when target_type is 'node'")
	}
	if target.NodeImage == "" {
		return nil, types.CommandResult{}, fmt.Errorf("node_image is required when target_type is 'node'")
	}
	_, output, err := s.k8sMcpServer.DebugNode(ctx, req, k8stypes.DebugNodeParams{Name: target.NodeName, Image: target.NodeImage, Command: command})
	if err != nil {
		return nil, types.CommandResult{}, err
	}
	return nil, types.CommandResult{Output: output.Stdout}, nil
}

// executeCommand runs a command on a pod
func (s *MCPServer) runExecPod(ctx context.Context, req *mcp.CallToolRequest, target types.ExecPodParams, command []string) (*mcp.CallToolResult, types.CommandResult, error) {
	if target.PodName == "" {
		return nil, types.CommandResult{}, fmt.Errorf("pod_name is required when target_type is 'pod'")
	}
	if target.PodNamespace == "" {
		target.PodNamespace = "default"
	}
	_, output, err := s.k8sMcpServer.ExecPod(ctx, req, k8stypes.ExecPodParams{
		NamespacedNameParams: k8stypes.NamespacedNameParams{
			Name:      target.PodName,
			Namespace: target.PodNamespace,
		},
		Container: target.ContainerName,
		Command:   command,
	})
	if err != nil {
		return nil, types.CommandResult{}, err
	}
	return nil, types.CommandResult{Output: output.Stdout}, nil
}

// commandBuilder helps build commands with a fluent interface
type commandBuilder struct {
	args []string
}

// newCommand creates a new command builder with the base command
func newCommand(baseCmd ...string) *commandBuilder {
	return &commandBuilder{args: baseCmd}
}

// add adds arguments to the command
func (cb *commandBuilder) add(args ...string) *commandBuilder {
	cb.args = append(cb.args, args...)
	return cb
}

// addIf adds arguments to the command only if the condition is true
func (cb *commandBuilder) addIf(condition bool, args ...string) *commandBuilder {
	if condition {
		cb.args = append(cb.args, args...)
	}
	return cb
}

// addIfNotEmpty adds arguments to the command only if the value is not empty
func (cb *commandBuilder) addIfNotEmpty(value string, args ...string) *commandBuilder {
	if value != "" {
		cb.args = append(cb.args, args...)
	}
	return cb
}

// build returns the final command slice
func (cb *commandBuilder) build() []string {
	return cb.args
}

func ValidateInterface(iface string) error {
	if iface == "" {
		return nil
	}
	if iface == "any" {
		return nil
	}
	if len(iface) > 15 {
		return fmt.Errorf("interface name too long: %s", iface)
	}
	if !interfaceNamePattern.MatchString(iface) {
		return fmt.Errorf("invalid interface name: %s", iface)
	}
	return nil
}

func ValidateBPFFilter(filter string) error {
	if filter == "" {
		return nil
	}
	if len(filter) > 1024 {
		return fmt.Errorf("BPF filter too long (max 1024 characters)")
	}

	dangerous := []string{";", "|", "&", "`", "$", "$("}
	for _, pattern := range dangerous {
		if strings.Contains(filter, pattern) {
			return fmt.Errorf("BPF filter contains potentially dangerous characters")
		}
	}
	return nil
}

// requireAtLeastNParams validates that at least N of the provided parameters are set
func requireAtLeastNParams(required int, params map[string]bool) error {
	count := 0
	var paramNames []string
	for name, isSet := range params {
		paramNames = append(paramNames, name)
		if isSet {
			count++
		}
	}
	if count < required {
		return fmt.Errorf("requires at least %d of: %v for safety", required, paramNames)
	}
	return nil
}

// validateIntMax checks if a value exceeds the maximum and returns an error if it does
func validateIntMax(value, max int, fieldName, unit string) error {
	if value > max {
		if unit != "" {
			return fmt.Errorf("%s cannot exceed %d %s", fieldName, max, unit)
		}
		return fmt.Errorf("%s cannot exceed %d", fieldName, max)
	}
	return nil
}

// stringWithDefault returns the value if non-empty, otherwise returns the default
func stringWithDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
