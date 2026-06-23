package mcp

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
)

// DebugNode debugs a node by name, by using a debug pod in a namespace, with an image and command.
func (s *MCPServer) DebugNode(ctx context.Context, req *mcp.CallToolRequest, in types.DebugNodeParams) (*mcp.CallToolResult, types.DebugNodeResult, error) {
	// Convert timeout seconds to duration
	timeout := in.TimeoutParams.ToDuration()

	stdout, stderr, err := s.clientSet.DebugNode(ctx, in.Namespace, in.Name, in.Image, in.Command, in.HostPath, in.MountPath, timeout)
	if err != nil {
		return nil, types.DebugNodeResult{}, err
	}

	return nil, types.DebugNodeResult{Stdout: stdout, Stderr: stderr}, nil
}

// RunDebugNode runs a debug node pod on the given node by name, namespace, image and command.
func (s *MCPServer) RunDebugNode(ctx context.Context, namespace string, nodeName string, image string, command []string, hostPath string, mountPath string, timeout time.Duration) (string, string, error) {
	return s.clientSet.DebugNode(ctx, namespace, nodeName, image, command, hostPath, mountPath, timeout)
}
