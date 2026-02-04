package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/must-gather/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
)

const defaultMaxLines = 100

// GetPodLogs uses the omc command to get the logs of a pod. It will return an error if the
// must gather path is not valid or the pod logs are not found.
func (s *MustGatherMCPServer) GetPodLogs(ctx context.Context, req *mcp.CallToolRequest, in types.GetPodLogsParams) (*mcp.CallToolResult, types.GetPodLogsResult, error) {
	// Check if name is not set, return an error.
	if in.Name == "" {
		return nil, types.GetPodLogsResult{}, fmt.Errorf("name is required")
	}

	// Match the pattern to the logs
	lines, err := in.PatternParams.ExecuteWithMatch(func() ([]string, error) {
		// Get the logs from the omc command
		output, err := s.omcClient.GetPodLogs(ctx, in.MustGatherPath, in.Namespace, in.Name, in.Container, in.Previous, in.Rotated)
		if err != nil {
			return nil, err
		}
		// Strip empty lines from the logs
		return utils.StripEmptyLines(strings.Split(output, "\n")), nil
	})
	if err != nil {
		return nil, types.GetPodLogsResult{}, err
	}

	// Apply the head and tail parameters to the logs
	lines = in.HeadTailParams.Apply(lines, defaultMaxLines)

	return nil, types.GetPodLogsResult{Logs: lines}, nil
}
