package kernel

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kernel/types"
)

// GetIPCommandOutput MCP handler for ip utility operations.
// GetIPCommandOutput executes 'ip' utility commands on a node.
// Requires ip utility in the debug container image.
func (s *MCPServer) GetIPCommandOutput(ctx context.Context, req *mcp.CallToolRequest, in types.ListIPParams) (*mcp.CallToolResult, types.Result, error) {
	ipCliAvailable, err := s.UtilityExists(ctx, req, in.Node, in.Image, "ip")
	if !ipCliAvailable {
		return nil, types.Result{}, fmt.Errorf("error while getting ip data: mentioned image does not have ip utility: %w", err)
	}

	if err := validateIPCommand(in.Command); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting ip data: %w", err)
	}
	if in.FilterParameters != "" {
		if err := validateParameters(in.FilterParameters); err != nil {
			return nil, types.Result{}, fmt.Errorf("error while getting ip data: %w", err)
		}
	}
	if in.Options != "" {
		if err := validateParameters(in.Options); err != nil {
			return nil, types.Result{}, fmt.Errorf("error while getting ip data: %w", err)
		}
	}

	cmd := newCommand("ip")
	cmd.addIfNotEmpty(in.Options, in.Options)
	cmd.add(strings.Fields(in.Command)...)
	cmd.addIfNotEmpty(in.FilterParameters, strings.Fields(in.FilterParameters)...)

	stdout, err := s.executeCommand(ctx, req, in.Node, in.Image, cmd.build())
	if err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting ip data: %w", err)
	}
	stdout = limitOutputLines(stdout, in.MaxLines)
	return nil, types.Result{Data: stdout}, nil
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
	validIpCommand := map[string]bool{
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
	for command := range validIpCommand {
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
