package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
)

const defaultMaxLines = 100

// GetPodLogs gets the logs of a pod by name and namespace.
func (s *MCPServer) GetPodLogs(ctx context.Context, req *mcp.CallToolRequest, in types.GetPodLogsParams) (*mcp.CallToolResult, types.GetPodLogsResult, error) {
	ctx, cancel := utils.ApplyTimeout(ctx, s.ToolTimeout)
	defer cancel()

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
		err = utils.WrapTimeoutError(err, fmt.Sprintf("pod-logs for pod %s/%s", in.Namespace, in.Name), s.ToolTimeout)
		return nil, types.GetPodLogsResult{}, err
	}

	// Apply the head and tail parameters to the logs
	lines = in.HeadTailParams.Apply(lines, defaultMaxLines)

	return nil, types.GetPodLogsResult{Logs: lines}, nil
}

// ExecPod executes a command in a pod by name and namespace.
// Note: This is an internal helper. Timeout should already be applied by the calling tool,
// but we apply it here as well for offline tools that don't support timeouts like Must Gather, SOS reports, etc.
func (s *MCPServer) ExecPod(ctx context.Context, req *mcp.CallToolRequest, in types.ExecPodParams) (*mcp.CallToolResult, types.ExecPodResult, error) {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = utils.ApplyTimeout(ctx, s.ToolTimeout)
		defer cancel()
	}

	stdout, stderr, err := s.clientSet.ExecPod(ctx, in.Name, in.Namespace, in.Container, in.Command)
	if err != nil {
		err = utils.WrapTimeoutError(err, fmt.Sprintf("exec-pod for pod %s/%s", in.Namespace, in.Name), s.ToolTimeout)
		return nil, types.ExecPodResult{}, err
	}

	return nil, types.ExecPodResult{Stdout: stdout, Stderr: stderr}, nil
}
