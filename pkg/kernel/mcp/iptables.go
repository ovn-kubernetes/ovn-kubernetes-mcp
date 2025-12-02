package kernel

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kernel/types"
)

// GetIptables MCP handler for iptables operations.
// GetIptables retrieves iptables/ip6tables rules from a Kubernetes node.
// Automatically detects IPv6 and uses ip6tables when needed.
func (s *MCPServer) GetIptables(ctx context.Context, req *mcp.CallToolRequest, in types.ListIPTablesParams) (*mcp.CallToolResult, types.Result, error) {
	iptablesCliAvailable, err := s.UtilityExists(ctx, req, in.Node, in.Image, "iptables")
	if !iptablesCliAvailable {
		return nil, types.Result{}, fmt.Errorf("error while getting list of iptables rules: mentioned image does not have iptables utility: %w", err)
	}

	if err := validateTableName(in.Table); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting list of iptables rules: %w", err)
	}
	if err := validateIptablesCommand(in.Command); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting list of iptables rules: %w", err)
	}
	if in.FilterParameters != "" {
		if err := validateParameters(in.FilterParameters); err != nil {
			return nil, types.Result{}, fmt.Errorf("error while getting list of iptables rules: %w", err)
		}
	}

	cmd := newCommand(iptablesCommand(in.FilterParameters))
	// Defaults to 'filter' table when not specified
	cmd.addIf(in.Table == "", "-t", "filter")
	cmd.addIfNotEmpty(in.Table, "-t", in.Table)
	// Defaults to -L (list) when command not specified
	cmd.addIf(in.Command == "", "-L")
	cmd.addIfNotEmpty(in.Command, in.Command)
	// FilterParameters are invalid with -S/--list-rules command
	cmd.addIf(in.FilterParameters != "" && in.Command != "-S" && in.Command != "--list-rules", strings.Fields(in.FilterParameters)...)

	stdout, err := s.executeCommand(ctx, req, in.Node, in.Image, cmd.build())
	if err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting list of iptables rules: %w", err)
	}
	stdout = limitOutputLines(stdout, in.MaxLines)
	return nil, types.Result{Data: stdout}, nil
}

// iptablesCommand determines whether to use iptables or ip6tables.
func iptablesCommand(filterParameters string) string {
	for _, item := range strings.Split(filterParameters, " ") {
		// Use of ip6tables CLI is required while either --ipv6 or -6 flag is mentioned
		if item == "--ipv6" || (strings.Contains(item, "-") && strings.Contains(item, "6")) {
			return "ip6tables"
		}
	}
	return "iptables"
}

// validateTableName validates iptables table name
func validateTableName(table string) error {
	if table == "" {
		return nil
	}

	if _, err := strconv.Atoi(table); err == nil {
		return fmt.Errorf("invalid table name: %s", table)
	}
	validTables := map[string]bool{
		"filter":   true,
		"nat":      true,
		"mangle":   true,
		"raw":      true,
		"security": true,
	}

	if !validTables[strings.TrimSpace(table)] {
		return fmt.Errorf("invalid table name: %s", table)
	}
	return nil
}

// validateIptablesCommand only allow list operation.
func validateIptablesCommand(command string) error {
	if command == "" {
		return nil
	}

	if _, err := strconv.Atoi(command); err == nil {
		return fmt.Errorf("invalid iptables command: %s", command)
	}
	validCommand := map[string]bool{
		"-L":           true,
		"-S":           true,
		"--list":       true,
		"--list-rules": true,
	}

	if !validCommand[strings.TrimSpace(command)] {
		return fmt.Errorf("invalid iptables command: %s", command)
	}
	return nil
}
