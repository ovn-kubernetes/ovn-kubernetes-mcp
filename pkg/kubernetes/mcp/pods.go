package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
)

const defaultMaxLines = 100

// GetPodLogs gets the logs of a pod by name and namespace.
func (s *MCPServer) GetPodLogs(ctx context.Context, req *mcp.CallToolRequest, in types.GetPodLogsParams) (*mcp.CallToolResult, types.GetPodLogsResult, error) {

	// Match the pattern to the logs
	lines, err := in.PatternParams.ExecuteWithMatch(func() ([]string, error) {
		lines, err := s.clientSet.GetPodLogs(ctx, in.Namespace, in.Name, in.Container, in.Previous)
		if err != nil {
			return nil, err
		}
		// Strip empty lines from the logs
		return utils.StripEmptyLines(lines), nil
	})
	if err != nil {
		return nil, types.GetPodLogsResult{}, err
	}

	// Apply the head and tail parameters to the logs
	lines = in.HeadTailParams.Apply(lines, defaultMaxLines)

	return nil, types.GetPodLogsResult{Logs: lines}, nil
}

// ExecPod executes a command in a pod by name and namespace.
func (s *MCPServer) ExecPod(ctx context.Context, req *mcp.CallToolRequest, in types.ExecPodParams) (*mcp.CallToolResult, types.ExecPodResult, error) {
	stdout, stderr, err := s.clientSet.ExecPod(ctx, in.Name, in.Namespace, in.Container, in.Command)
	if err != nil {
		return nil, types.ExecPodResult{}, err
	}

	return nil, types.ExecPodResult{Stdout: stdout, Stderr: stderr}, nil
}
