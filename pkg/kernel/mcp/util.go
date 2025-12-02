package kernel

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	k8stypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
)

const (
	// MaxOutputLines defines the maximum number of lines to return from command output
	MaxOutputLines = 100
)

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

// UtilityExists checks if a utility/command exists in the container
func (s *MCPServer) UtilityExists(ctx context.Context, req *mcp.CallToolRequest, node, image, utility string) (bool, error) {
	cmd := newCommand("chroot", "/host", utility, "-V")
	debugParameter := k8stypes.DebugNodeParams{Name: node, Image: image, Command: cmd.build()}
	_, result, err := s.k8sMcpServer.DebugNode(ctx, req, debugParameter)
	if err != nil || filterWarnings(result.Stderr) != "" {
		return false, fmt.Errorf("error while checking availability of the utility %s: %s : %w", utility, filterWarnings(result.Stderr), err)
	}
	return true, nil
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

// limitOutputLines limits the output to a maximum number of lines.
// If the output exceeds the maximum, it truncates and adds a message indicating truncation.
func limitOutputLines(output string, maxLines int) string {
	if output == "" {
		return output
	}
	if maxLines <= 0 {
		maxLines = MaxOutputLines
	}

	lines := strings.Split(output, "\n")
	if len(lines) <= maxLines {
		return output
	}

	// Truncate to maxLines and add a truncation message
	truncatedLines := lines[:maxLines]
	truncatedLines = append(truncatedLines, fmt.Sprintf("\n... Output truncated. Showing first %d lines out of %d total lines.", maxLines, len(lines)))

	return strings.Join(truncatedLines, "\n")
}

// validateParameters validates whether any parameter contains shell
// metacharacters or not. It returns an error if there are any.
func validateParameters(param string) error {
	shellMetacharacters := regexp.MustCompile(`[;&|$` + "`" + `<>\\()]`)
	if shellMetacharacters.MatchString(param) {
		return fmt.Errorf("invalid use of metacharacters in parameter: %s", param)
	}

	return nil
}
