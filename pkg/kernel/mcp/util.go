package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	k8stypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
)

const (
	// defaultMaxOutputLines defines the maximum number of lines to return from command output
	defaultMaxOutputLines = 100
)

// utilityExists checks if a utility/command exists in the container
func (s *MCPServer) utilityExists(ctx context.Context, req *mcp.CallToolRequest, node, utility string) error {
	cmd := utils.NewCommand(utility, "-V")
	debugParameter := k8stypes.DebugNodeParams{Name: node, Image: s.cfg.Image, Command: cmd.Build()}
	_, result, err := s.k8sMcpServer.DebugNode(ctx, req, debugParameter)
	if err != nil {
		return fmt.Errorf("error while checking availability of the utility %s: %w", utility, err)
	}
	stderr := strings.TrimSpace(filterWarnings(result.Stderr))
	if stderr != "" {
		return fmt.Errorf("utility %s is unavailable in configured image: %s", utility, stderr)
	}
	return nil
}

// filterWarnings filters out lines starting with "Warning" from the output
func filterWarnings(output string) string {
	if output == "" {
		return output
	}

	lines := strings.Split(output, "\n")
	var filteredLines []string

	for _, line := range lines {
		if !strings.Contains(line, "Warning") && !strings.Contains(line, "warning") && !strings.Contains(line, "WARNING") {
			filteredLines = append(filteredLines, line)
		}
	}

	return strings.Join(filteredLines, "\n")
}
