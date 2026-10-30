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

// GetIPCommandOutput MCP handler for ip utility operations.
// GetIPCommandOutput executes 'ip' utility commands on a node.
// Requires ip utility in the debug container image.
func (s *MCPServer) GetIPCommandOutput(ctx context.Context, req *mcp.CallToolRequest, in types.ListIPParams) (*mcp.CallToolResult, types.Result, error) {
	// If timeout is specified, create a new context with timeout
	var cancel context.CancelFunc
	ctx, cancel = in.TimeoutParams.WithTimeout(ctx)
	if cancel != nil {
		defer cancel()
	}

	if err := validateIPCommand(in.Command); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting ip data: %w", err)
	}
	if err := utils.ValidateSafeString(in.FilterParameters, "filter parameters", true, utils.ShellMetaCharactersTypeDefault); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting ip data: %w", err)
	}
	options, err := parseIPOptions(in.Options)
	if err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting ip data: %w", err)
	}
	if err := s.utilityExists(ctx, in.Namespace, in.Node, "ip"); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting ip data: failed to verify ip utility availability in configured image: %w", err)
	}

	cmd := commandbuilder.NewCommand("ip")
	cmd.Add(options...)
	cmd.Add(strings.Fields(in.Command)...)
	cmd.AddIfNotEmpty(in.FilterParameters, strings.Fields(in.FilterParameters)...)

	stdout, stderr, err := s.executeCommand(ctx, in.Namespace, in.Node, cmd.Build())
	if err != nil {
		lines := in.HeadTailParams.Apply(strings.Split(strings.TrimSpace(stdout), "\n"), DefaultMaxOutputLines)
		err = utils.FormatCommandExecutionError(strings.Join(lines, "\n"), stderr, err)
		return nil, types.Result{}, fmt.Errorf("error while getting ip data: %w", err)
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

// parseIPOptions accepts diagnostic global options without ip's prefix matching.
// In particular, -b selects batch mode, while -br selects brief output.
func parseIPOptions(options string) ([]string, error) {
	if err := utils.ValidateSafeString(options, "options", true, utils.ShellMetaCharactersTypeDefault); err != nil {
		return nil, err
	}
	args := strings.Fields(options)
	for i := 0; i < len(args); i++ {
		opt := args[i]
		if strings.HasPrefix(opt, "--") {
			opt = opt[1:] // ip accepts either one or two leading hyphens.
		}
		if name, value, hasValue := strings.Cut(opt, "="); hasValue {
			if (name != "-c" && name != "-color") || (value != "" && value != "always" && value != "auto" && value != "never") {
				return nil, fmt.Errorf("invalid ip option: %q", args[i])
			}
			continue
		}
		switch opt {
		case "-4", "-6", "-0", "-B", "-M",
			"-d", "-details", "-s", "-stats", "-statistics", "-o", "-oneline",
			"-r", "-resolve", "-a", "-all", "-N", "-Numeric",
			"-j", "-json", "-p", "-pretty", "-br", "-brief",
			"-h", "-human", "-human-readable", "-iec",
			"-t", "-timestamp", "-ts", "-tshort", "-c", "-color":
		case "-n", "-netns", "-f", "-family", "-rc", "-rcvbuf":
			if i+1 == len(args) || strings.HasPrefix(args[i+1], "-") || strings.ContainsRune(args[i+1], 0) {
				return nil, fmt.Errorf("missing or invalid value for ip option: %q", args[i])
			}
			i++
			switch opt {
			case "-f", "-family":
				switch args[i] {
				case "inet", "inet6", "link", "bridge", "mpls":
				default:
					return nil, fmt.Errorf("invalid ip address family: %q", args[i])
				}
			case "-rc", "-rcvbuf":
				if _, err := strconv.ParseUint(args[i], 0, 32); err != nil {
					return nil, fmt.Errorf("invalid ip receive buffer size: %q", args[i])
				}
			}
		default:
			return nil, fmt.Errorf("invalid ip option: %q", args[i])
		}
	}
	return args, nil
}

// validateIPCommand validates that the IP command is allowed.
func validateIPCommand(ipCommand string) error {
	if _, err := strconv.Atoi(ipCommand); err == nil {
		return fmt.Errorf("invalid ip command: %s", ipCommand)
	}
	splitCommand := strings.Fields(strings.TrimSpace(ipCommand))
	if len(splitCommand) < 2 {
		return fmt.Errorf("invalid ip command: %s", ipCommand)
	}
	validIPCommand := map[string]bool{
		"address":     true,
		"link":        true,
		"neighbour":   true,
		"netns":       true,
		"route":       true,
		"rule":        true,
		"vrf":         true,
		"xfrm state":  true,
		"xfrm policy": true,
	}

	// "ip l s" can be used to set a link up or down. This should be considered as an invalid command.
	if strings.HasPrefix(splitCommand[0], "l") && splitCommand[1] == "s" {
		return fmt.Errorf("invalid ip command: %s", ipCommand)
	}

	var valid bool
	for command := range validIPCommand {
		commandFields := strings.Fields(command)
		// Check if the input command matches the expected command pattern
		// For single-word commands like "address", we check: splitCommand[0] matches "address" and splitCommand[1] matches "show"
		// For multi-word xfrm commands, we check: splitCommand[0:2] matches ["xfrm", "state"/"policy"] and splitCommand[2] matches "list"

		if len(commandFields) == 1 {
			// Single-word command (e.g., "address show", "link show")
			if strings.HasPrefix(command, splitCommand[0]) && strings.HasPrefix("show", splitCommand[1]) {
				valid = true
				break
			}
		} else {
			// Multi-word command (e.g., "xfrm state list", "xfrm policy list")
			// Need at least len(commandFields) + 1 words (command words + subcommand)
			if len(splitCommand) >= len(commandFields)+1 {
				allMatch := true
				for i, cmdField := range commandFields {
					if !strings.HasPrefix(cmdField, splitCommand[i]) {
						allMatch = false
						break
					}
				}
				// For xfrm state and xfrm policy, only accept "list"
				if allMatch && (command == "xfrm state" || command == "xfrm policy") {
					if strings.HasPrefix("list", splitCommand[len(commandFields)]) {
						valid = true
						break
					}
				}
			}
		}
	}
	if !valid {
		return fmt.Errorf("invalid ip command: %s", ipCommand)
	}

	return nil
}
