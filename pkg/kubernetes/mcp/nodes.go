package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
)

// DebugNode debugs a node by name, image and command.
func (s *MCPServer) DebugNode(ctx context.Context, req *mcp.CallToolRequest, in types.DebugNodeParams) (*mcp.CallToolResult, types.DebugNodeResult, error) {
	// Validate paths before creating the pod
	if err := utils.ValidatePath(in.HostPath, "hostPath", true); err != nil {
		return nil, types.DebugNodeResult{}, err
	}

	if err := utils.ValidatePath(in.MountPath, "mountPath", true); err != nil {
		return nil, types.DebugNodeResult{}, err
	}

	stdout, stderr, err := s.clientSet.DebugNode(ctx, in.Name, in.Image, in.Command, in.HostPath, in.MountPath)
	if err != nil {
		return nil, types.DebugNodeResult{}, err
	}

	return nil, types.DebugNodeResult{Stdout: stdout, Stderr: stderr}, nil
}
