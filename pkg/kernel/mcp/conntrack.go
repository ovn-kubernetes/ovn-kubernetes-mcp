package kernel

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kernel/types"
)

const conntrackSystemFile = "/proc/net/nf_conntrack"

// GetConntrack MCP handler for conntrack operations.
// GetConntrack retrieves connection tracking entries from a Kubernetes node.
// TODO: Add support for conntrack event monitoring (-E flag).
// TODO: Add support for -G (get specific entry) command.
func (s *MCPServer) GetConntrack(ctx context.Context, req *mcp.CallToolRequest, in types.ListConntrackParams) (*mcp.CallToolResult, types.Result, error) {
	// Falls back to /proc/net/nf_conntrack parsing when conntrack CLI unavailable.
	conntrackCliAvailable, _ := s.UtilityExists(ctx, req, in.Node, in.Image, "conntrack")
	if err := validateConntrackCommand(in.Command, conntrackCliAvailable); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting list of conntrack entries: %w", err)
	}
	if in.FilterParameters != "" {
		if err := validateParameters(in.FilterParameters); err != nil {
			return nil, types.Result{}, fmt.Errorf("error while getting list of conntrack entries: %w", err)
		}
	}

	var stdout string
	var err error
	if !conntrackCliAvailable {
		stdout, err = s.getConntrackFromFile(ctx, req, in.Node, in.Image)
		if err != nil {
			return nil, types.Result{}, fmt.Errorf("error while getting list of conntrack entries: %w", err)
		}
	} else {
		stdout, err = s.getConntrackUsingCLI(ctx, req, in.Node, in.Image, in.Command, in.FilterParameters)
		if err != nil {
			return nil, types.Result{}, fmt.Errorf("error while getting list of conntrack entries: %w", err)
		}
	}
	stdout = limitOutputLines(stdout, in.MaxLines)
	return nil, types.Result{Data: stdout}, nil
}

// getConntrackUsingCLI executes conntrack CLI commands.
func (s *MCPServer) getConntrackUsingCLI(ctx context.Context, req *mcp.CallToolRequest, node, image, command, filterParameters string) (string, error) {
	cmd := newCommand("conntrack")
	switch command {
	case "-L", "--dump":
		cmd.add(command)
		cmd.add(strings.Fields(filterParameters)...)
	case "-S", "--stats":
		cmd.add(command)
	case "-C", "--count":
		cmd.add(command)
	default:
		cmd.add("-L")
		cmd.add(strings.Fields(filterParameters)...)
	}
	return s.executeCommand(ctx, req, node, image, cmd.build())
}

// getConntrackFromFile parses /proc/net/nf_conntrack directly.
// TODO: Add filter support while getting conntrack entries from /proc/net/nf_conntrack.
func (s *MCPServer) getConntrackFromFile(ctx context.Context, req *mcp.CallToolRequest, node, image string) (string, error) {
	cmd := newCommand("cat")
	cmd.add(conntrackSystemFile)
	return s.executeCommand(ctx, req, node, image, cmd.build())
}

// validateConntrackCommand validates the command to be used to get list of conntrack entries.
// It returns an error if conntrack CLI is not available and any other operation than list is being performed.
func validateConntrackCommand(command string, cliAvailable bool) error {
	if command == "" {
		return nil
	}
	if !cliAvailable && (strings.TrimSpace(command) != "-L" && strings.TrimSpace(command) != "--dump") {
		return fmt.Errorf("mentioned image does not have conntrack utility, only -L is supported with limited filters")
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
