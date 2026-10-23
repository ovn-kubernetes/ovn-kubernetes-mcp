// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kernel/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils/commandbuilder"
)

// GetIptables MCP handler for iptables operations.
// GetIptables retrieves iptables/ip6tables rules from a Kubernetes node.
// Automatically detects IPv6 and uses ip6tables when needed.
func (s *MCPServer) GetIptables(ctx context.Context, req *mcp.CallToolRequest, in types.ListIPTablesParams) (*mcp.CallToolResult, types.Result, error) {
	// If timeout is specified, create a new context with timeout
	var cancel context.CancelFunc
	ctx, cancel = in.TimeoutParams.WithTimeout(ctx)
	if cancel != nil {
		defer cancel()
	}

	if err := validateTableName(in.Table); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting list of iptables rules: %w", err)
	}
	if err := validateIptablesCommand(in.Command); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting list of iptables rules: %w", err)
	}

	binary, filterArgs, err := parseIptablesFilters(in.FilterParameters)
	if err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting list of iptables rules: %w", err)
	}
	if err := s.utilityExists(ctx, in.Namespace, in.Node, binary); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting list of iptables rules: failed to verify iptables utility availability in configured image: %w", err)
	}

	table := strings.TrimSpace(in.Table)
	command := strings.TrimSpace(in.Command)
	cmd := commandbuilder.NewCommand(binary)
	// Defaults to 'filter' table when not specified
	cmd.AddIf(table == "", "-t", "filter")
	cmd.AddIfNotEmpty(table, "-t", table)
	// Defaults to -L (list) when command not specified
	cmd.AddIf(command == "", "-L")
	cmd.AddIfNotEmpty(command, command)
	// FilterParameters are invalid with -S/--list-rules command
	cmd.AddIf(command != "-S" && command != "--list-rules", filterArgs...)

	stdout, stderr, err := s.executeCommand(ctx, in.Namespace, in.Node, cmd.Build())
	if err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting list of iptables rules: %w", err)
	}

	if stderr != "" {
		return nil, types.Result{}, fmt.Errorf("error while running command: %s", stderr)
	}

	// Strip empty lines from the output
	lines := utils.StripEmptyLines(strings.Split(stdout, "\n"))
	// Apply the head and tail parameters to the lines
	lines = in.HeadTailParams.Apply(lines, DefaultMaxOutputLines)
	// Join the lines back into a single string
	stdout = strings.Join(lines, "\n")
	return nil, types.Result{Data: stdout}, nil
}

var iptablesLongFilters = map[string]byte{
	"--verbose": 'v', "--numeric": 'n', "--exact": 'x', "--line-numbers": 'n',
	"--wait": 'w', "--ipv4": '4', "--ipv6": '6',
}

// parseIptablesFilters accepts [chain [rulenum]] followed by listing options.
// It returns the original tokens and selects IPv6 only from an explicit flag.
func parseIptablesFilters(filters string) (string, []string, error) {
	if err := utils.ValidateSafeString(filters, "filter parameters", true, utils.ShellMetaCharactersTypeDefault); err != nil {
		return "", nil, err
	}
	args := strings.Fields(filters)
	binary := "iptables"
	firstOption := 0
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		if strings.HasPrefix(args[0], "!") || strings.ContainsRune(args[0], 0) {
			return "", nil, fmt.Errorf("invalid iptables chain: %q", args[0])
		}
		firstOption = 1
		if len(args) > 1 && !strings.HasPrefix(args[1], "-") {
			rule, err := strconv.ParseUint(args[1], 0, 31)
			if err != nil || rule == 0 {
				return "", nil, fmt.Errorf("invalid iptables rule number: %q", args[1])
			}
			firstOption = 2
		}
	}
	for i := firstOption; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			return "", nil, fmt.Errorf("unexpected iptables filter argument: %q", arg)
		}
		options := arg[1:]
		if strings.HasPrefix(arg, "--") {
			name, value, hasValue := strings.Cut(arg, "=")
			flag, ok := iptablesLongFilters[name]
			if !ok || (hasValue && (value == "" || flag != 'w')) {
				return "", nil, fmt.Errorf("invalid iptables filter option: %q", arg)
			}
			options = string(flag) + value
		}
		if options == "" {
			return "", nil, fmt.Errorf("invalid iptables filter option: %q", arg)
		}
		for len(options) > 0 {
			flag := options[0]
			options = options[1:]
			switch flag {
			case 'v', 'n', 'x', '4':
			case '6':
				binary = "ip6tables"
			case 'w':
				value := options
				if value == "" && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
					value = args[i]
				}
				if value == "" {
					break // Waiting without a numeric limit is supported by iptables.
				}
				if _, err := strconv.ParseUint(value, 0, 31); err != nil {
					return "", nil, fmt.Errorf("invalid value for iptables filter option %q: %q", arg, value)
				}
				options = "" // A value consumes the remainder of a short-option bundle.
			default:
				return "", nil, fmt.Errorf("invalid iptables filter option: %q", arg)
			}
		}
	}
	return binary, args, nil
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
