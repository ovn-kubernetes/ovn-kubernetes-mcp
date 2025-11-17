package mcp

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	k8stypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
)

const defaultMaxLines = 100

func (s *MCPServer) runCommand(ctx context.Context, req *mcp.CallToolRequest, namespacedName k8stypes.NamespacedNameParams,
	commands []string) ([]string, error) {
	_, result, err := s.k8sMcpServer.ExecPod(ctx, req, k8stypes.ExecPodParams{NamespacedNameParams: namespacedName, Command: commands})
	if err != nil {
		return nil, err
	}
	if result.Stderr != "" {
		return nil, fmt.Errorf("error occurred while running command %v on pod %s/%s: %s", commands, namespacedName.Namespace,
			namespacedName.Name, result.Stderr)
	}
	output := []string{} // Initialize with empty slice to ensure valid JSON when there's no output
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			output = append(output, line)
		}
	}
	return output, nil
}

// filterLines filters lines using a regex pattern.
func filterLines(lines []string, pattern string) ([]string, error) {
	if pattern == "" {
		return lines, nil
	}

	filterPattern, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid filter pattern %s: %w", pattern, err)
	}

	filtered := []string{} // Initialize with empty slice to ensure valid JSON when there's no output
	for _, line := range lines {
		if filterPattern.MatchString(line) {
			filtered = append(filtered, line)
		}
	}
	return filtered, nil
}

// limitLines limits the number of lines returned.
func limitLines(lines []string, maxLines int) []string {
	if maxLines <= 0 {
		maxLines = defaultMaxLines
	}
	if len(lines) > maxLines {
		return lines[:maxLines]
	}
	return lines
}

// validateBridgeName validates that a bridge name is safe and non-empty.
// Bridge names should only contain alphanumeric characters, hyphens, and underscores.
func validateBridgeName(bridge string) error {
	if bridge == "" {
		return fmt.Errorf("bridge name cannot be empty")
	}

	// OVS bridge names typically follow naming conventions: alphanumeric, hyphens, underscores
	validBridgeName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validBridgeName.MatchString(bridge) {
		return fmt.Errorf("invalid bridge name %q: must contain only alphanumeric characters, hyphens, and underscores", bridge)
	}

	return nil
}

// validateFlowSpec validates that a flow specification is safe and non-empty.
func validateFlowSpec(flow string) error {
	if flow == "" {
		return fmt.Errorf("flow specification cannot be empty")
	}

	// Check for potentially dangerous characters that shouldn't appear in flow specs
	// Flow specs should contain: alphanumeric, commas, equals, colons, periods, slashes, parentheses, brackets
	// We explicitly block: semicolons, pipes, backticks, dollar signs, and other shell metacharacters
	dangerousChars := regexp.MustCompile(`[;&|$` + "`" + `<>\\]`)
	if dangerousChars.MatchString(flow) {
		return fmt.Errorf("invalid flow specification: contains potentially dangerous characters")
	}

	return nil
}

// validateConntrackParams validates that conntrack additional parameters are safe.
// Valid parameters for dpctl/dump-conntrack include: zone=N, mark=0xN, labels=0xN, -m, -s, etc.
func validateConntrackParams(params []string) error {
	for _, param := range params {
		if param == "" {
			return fmt.Errorf("conntrack parameter cannot be empty")
		}

		// Check for potentially dangerous characters
		// Valid conntrack params should contain: alphanumeric, equals, hyphens, underscores, periods, colons, commas, forward slashes
		// We explicitly block: semicolons, pipes, backticks, dollar signs, ampersands, and other shell metacharacters
		dangerousChars := regexp.MustCompile(`[;&|$` + "`" + `<>\\()]`)
		if dangerousChars.MatchString(param) {
			return fmt.Errorf("invalid conntrack parameter %q: contains potentially dangerous characters", param)
		}

		// Additional validation for common parameter patterns
		// Valid patterns include:
		// - Single-char flags: -m, -s (single hyphen followed by single letter)
		// - Key=value pairs: zone=5, mark=0x1, src=10.0.0.1 (key must contain only alphanumeric, underscore, hyphen)
		validParam := regexp.MustCompile(`^(-[a-zA-Z]|[a-zA-Z0-9_-]+=[a-zA-Z0-9x.:,/_-]+)$`)
		if !validParam.MatchString(param) {
			return fmt.Errorf("invalid conntrack parameter format %q: must be a flag (e.g., '-m') or key=value pair (e.g., 'zone=5')", param)
		}
	}

	return nil
}
