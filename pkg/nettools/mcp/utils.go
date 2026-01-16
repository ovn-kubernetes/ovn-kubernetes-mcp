package nettools

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	k8stypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/nettools/types"
)

var interfaceNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

// runDebugNode runs a command on a node
func (s *MCPServer) runDebugNode(ctx context.Context, req *mcp.CallToolRequest, target k8stypes.DebugNodeParams) (types.CommandResult, error) {
	if target.Name == "" {
		return types.CommandResult{}, fmt.Errorf("node's name is required when target type is 'node'")
	}
	if target.Image == "" {
		return types.CommandResult{}, fmt.Errorf("node's image is required when target type is 'node'")
	}
	_, output, err := s.k8sMcpServer.DebugNode(ctx, req, target)
	if err != nil {
		return types.CommandResult{}, err
	}
	return types.CommandResult{Output: output.Stdout}, nil
}

// runExecPod runs a command on a pod
func (s *MCPServer) runExecPod(ctx context.Context, req *mcp.CallToolRequest, target k8stypes.ExecPodParams) (types.CommandResult, error) {
	if target.NamespacedNameParams.Name == "" {
		return types.CommandResult{}, fmt.Errorf("pod's name is required when target type is 'pod'")
	}
	if target.NamespacedNameParams.Namespace == "" {
		target.NamespacedNameParams.Namespace = "default"
	}
	_, output, err := s.k8sMcpServer.ExecPod(ctx, req, target)
	if err != nil {
		return types.CommandResult{}, err
	}
	return types.CommandResult{Output: output.Stdout}, nil
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

// validateInterface validates a network interface name for security and correctness.
func validateInterface(iface string) error {
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

// validatePacketFilter validates a packet filter expression for security.
func validatePacketFilter(filter string) error {
	if filter == "" {
		return nil
	}
	if len(filter) > 1024 {
		return fmt.Errorf("packet filter too long (max 1024 characters)")
	}

	dangerous := []string{";", "|", "&", "`", "$", "$(", "\n", "\x00", ">", "<", "'", "\""}
	for _, pattern := range dangerous {
		if strings.Contains(filter, pattern) {
			return fmt.Errorf("packet filter contains potentially dangerous characters")
		}
	}
	return nil
}

// validateIntMax checks if a value exceeds the maximum or is negative and returns an error if it does
func validateIntMax(value, max int, fieldName, unit string) error {
	if value < 0 {
		return fmt.Errorf("%s cannot be negative", fieldName)
	}
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
