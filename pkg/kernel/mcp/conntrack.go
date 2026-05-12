package mcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kernel/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
)

const conntrackSystemFile = "/proc/net/nf_conntrack"

// GetConntrack MCP handler for conntrack operations.
// GetConntrack retrieves connection tracking entries from a Kubernetes node.
// TODO: Add support for conntrack event monitoring (-E flag).
// TODO: Add support for -G (get specific entry) command.
func (s *MCPServer) GetConntrack(ctx context.Context, req *mcp.CallToolRequest, in types.ListConntrackParams) (*mcp.CallToolResult, types.Result, error) {
	// Falls back to /proc/net/nf_conntrack parsing when conntrack CLI unavailable.
	err := s.utilityExists(ctx, req, in.Node, "conntrack")
	conntrackCliAvailable := err == nil // true if conntrack CLI is available, false otherwise
	if err := validateConntrackCommand(in.Command, conntrackCliAvailable); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting list of conntrack entries: %w", err)
	}
	if err := utils.ValidateSafeString(in.FilterParameters, "filter parameters", true, utils.ShellMetaCharactersTypeDefault); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting list of conntrack entries: %w", err)
	}

	var stdout string
	if !conntrackCliAvailable {
		stdout, err = s.getConntrackFromFile(ctx, req, in.Node)
	} else {
		stdout, err = s.getConntrackUsingCLI(ctx, req, in.Node, strings.TrimSpace(in.Command), in.FilterParameters)
	}
	if err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting list of conntrack entries: %w", err)
	}

	// Strip empty lines from the output
	lines := utils.StripEmptyLines(strings.Split(stdout, "\n"))
	// Apply the head and tail parameters to the lines
	lines = in.HeadTailParams.Apply(lines, defaultMaxOutputLines)
	// Join the lines back into a single string
	stdout = strings.Join(lines, "\n")
	return nil, types.Result{Data: stdout}, nil
}

// getConntrackUsingCLI executes conntrack CLI commands.
func (s *MCPServer) getConntrackUsingCLI(ctx context.Context, req *mcp.CallToolRequest, node, command, filterParameters string) (string, error) {
	cmd := utils.NewCommand("conntrack")
	switch command {
	case "-L", "--dump":
		cmd.Add(command)
		cmd.Add(strings.Fields(filterParameters)...)
	case "-S", "--stats":
		cmd.Add(command)
	case "-C", "--count":
		cmd.Add(command)
	default:
		cmd.Add("-L")
		cmd.Add(strings.Fields(filterParameters)...)
	}
	return s.executeCommand(ctx, req, node, cmd.Build())
}

// getConntrackFromFile parses /proc/net/nf_conntrack directly.
// TODO: Add filter support while getting conntrack entries from /proc/net/nf_conntrack.
func (s *MCPServer) getConntrackFromFile(ctx context.Context, req *mcp.CallToolRequest, node string) (string, error) {
	cmd := utils.NewCommand("cat")
	cmd.Add(conntrackSystemFile)
	return s.executeCommand(ctx, req, node, cmd.Build())
}

// validateConntrackCommand validates the command to be used to get list of conntrack entries.
// It returns an error if conntrack CLI is not available and any other operation than list is being performed.
func validateConntrackCommand(command string, cliAvailable bool) error {
	if command == "" {
		return nil
	}
	if !cliAvailable && (strings.TrimSpace(command) != "-L" && strings.TrimSpace(command) != "--dump") {
		return fmt.Errorf("configured image does not have conntrack utility, only -L/--dump is supported with limited filters")
	}
	if _, err := strconv.Atoi(command); err == nil {
		return fmt.Errorf("invalid command: %s", command)
	}
	validTables := map[string]bool{
		"-L":      true,
		"-S":      true,
		"-C":      true,
		"--dump":  true,
		"--stats": true,
		"--count": true,
	}

	if !validTables[strings.TrimSpace(command)] {
		return fmt.Errorf("invalid command: %s", command)
	}
	return nil
}
