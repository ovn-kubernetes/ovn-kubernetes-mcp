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

	// Check if head and tail are used together. If they are, return an error.
	if in.Head != 0 && in.Tail != 0 {
		return nil, types.GetPodLogsResult{}, fmt.Errorf("head and tail cannot be used together")
	}

	// Get the logs from the omc command
	output, err := s.omcClient.GetPodLogs(ctx, in.MustGatherPath, in.Namespace, in.Name, in.Container, in.Previous, in.Rotated)
	if err != nil {
		return nil, types.GetPodLogsResult{}, err
	}

	// Strip empty lines from the logs
	lines := utils.StripEmptyLines(strings.Split(output, "\n"))
	// If a pattern is provided, match the pattern to the logs
	if in.Pattern != "" {
		// Match the pattern to the logs
		lines, err = utils.MatchPattern(in.Pattern, lines)
		if err != nil {
			return nil, types.GetPodLogsResult{}, err
		}
	}

	// Set the default maximum number of lines to return
	maxLines := defaultMaxLines
	// If tail is not used (including when both are unset), use head
	if in.Tail == 0 {
		// Override default if Head is explicitly set
		if in.Head > 0 {
			maxLines = in.Head
		}
		lines = utils.Head(lines, maxLines)
	} else {
		// Tail is set, use tail (override default if Tail > 0)
		if in.Tail > 0 {
			maxLines = in.Tail
		}
		lines = utils.Tail(lines, maxLines)
	}

	return nil, types.GetPodLogsResult{Logs: lines}, nil
}
