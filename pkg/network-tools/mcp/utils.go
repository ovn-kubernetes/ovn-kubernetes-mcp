package mcp

import (
	"context"
	"fmt"
	"regexp"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	k8stypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/network-tools/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
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
	if len(filter) > 1024 {
		return fmt.Errorf("packet filter too long (max 1024 characters)")
	}
	if err := utils.ValidateSafeString(filter, "packet filter", true, utils.ShellMetaCharactersTypeDisallowSpecialCharacters); err != nil {
		return fmt.Errorf("packet filter contains potentially dangerous characters: %s, error: %w", filter, err)
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
