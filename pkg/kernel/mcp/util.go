package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils/commandbuilder"
)

const (
	// defaultMaxOutputLines defines the maximum number of lines to return from command output
	defaultMaxOutputLines = 100
)

// utilityExists checks if a utility/command exists in the container
func (s *MCPServer) utilityExists(ctx context.Context, namespace, node, utility string) error {
	cmd := commandbuilder.NewCommand(utility, "-V")
	_, stderr, err := s.runDebugNodeCommand(ctx, namespace, node, s.cfg.Image, cmd.Build(), "", "", 0)
	if err != nil {
		return fmt.Errorf("error while checking availability of the utility %s: %w", utility, err)
	}
	stderr = strings.TrimSpace(filterWarnings(stderr))
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
