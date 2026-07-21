package sosreport

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/sosreport/types"
)

// MCPServer provides sosreport analysis tools
type MCPServer struct{}

// NewMCPServer creates a new sosreport MCP server
func NewMCPServer() *MCPServer {
	return &MCPServer{}
}

// AddTools registers sosreport tools with the MCP server
func (s *MCPServer) AddTools(server *mcp.Server) {
	// List enabled plugins
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "sos-list-plugins",
			Description: `List enabled sosreport plugins with their command counts.
Parameters:
- sosreport_path (required): Path to extracted sosreport directory

Returns a list of  plugins with the number of commands collected by each plugin.

Use this to discover which plugins are available, then use sos-list-commands to see
what commands are available within a specific plugin.

Example output:
{
  "plugins": [
    {"name": "crio", "command_count": 15},
    {"name": "openvswitch", "command_count": 187},
    {"name": "networkmanager", "command_count": 2}
  ],
  "total_commands": 204
}`,
		}, s.ListPlugins)

	// List commands for a plugin
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "sos-list-commands",
			Description: `List all commands collected by a specific sosreport plugin.

Parameters:
- sosreport_path (required): Path to extracted sosreport directory
- plugin (required): Plugin name (e.g. 'openvswitch', 'networking', 'kubernetes')

Returns all commands executed by the plugin with their filepaths. Use the filepath
with sos-get-command to retrieve the actual command output.

Example:
- plugin='openvswitch' returns ovs-vsctl, ovs-ofctl, ovs-appctl commands
- plugin='networking' returns ip, ethtool, netstat commands`,
		}, s.ListCommands)

	// Search for commands
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "sos-search-commands",
			Description: fmt.Sprintf(`Search for commands across all plugins by pattern.

Parameters:
- sosreport_path (required): Path to extracted sosreport directory
- pattern (required): Regex pattern to search in command exec and filepath
- max_results (optional): Maximum results to return (default: %d)

Searches command names and filepaths across all plugins. Returns matching commands
with their plugin, exec string, and filepath. Does NOT return file contents.

Examples:
- pattern='iptables' finds all iptables-related commands
- pattern='ovn.*show' finds OVN show commands
- pattern='journalctl.*kubelet' finds kubelet journal logs`, DefaultMaxResults),
		}, s.SearchCommands)

	// Get command output
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "sos-get-command",
			Description: fmt.Sprintf(`Get command output using filepath from manifest.

Parameters:
- sosreport_path (required): Path to extracted sosreport directory
- filepath (required): Relative filepath from sos-list-commands or sos-search-commands
- pattern (optional): Regex pattern to filter output lines
- head (optional): Return only first N lines. Default: %d lines if tail is not specified
- tail (optional): Return only last N lines
- apply_tail_first (optional): If both head and tail are set and apply_tail_first is true,
apply tail before head. Default: false

Use the filepath returned by sos-list-commands or sos-search-commands to retrieve
the actual command output. Supports optional grep-style filtering.

Example:
- filepath='sos_commands/openvswitch/ovs-vsctl_-t_5_show'
- filepath='sos_commands/firewall_tables/iptables_-vnxL', pattern='KUBE-'`, DefaultMaxLines),
		}, s.GetCommand)

	// Get pod logs
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "sos-get-pod-logs",
			Description: fmt.Sprintf(`Get Kubernetes pod container logs from a sosreport.

Parameters:
- sosreport_path (required): Path to extracted sosreport directory
- namespace (required): Pod namespace
- name (required): Pod name
- container (optional): Container name. If omitted, returns the first matching container log for the pod.
- pattern (optional): Regex pattern to filter log lines (grep-style filtering)
- head (optional): Return only first N lines. Default: %d lines if tail is not specified
- tail (optional): Return only last N lines
- apply_tail_first (optional): If both head and tail are set and apply_tail_first is true,
apply tail before head. Default: false

Reads container logs collected from the sosreport for the specified pod. Returns
log lines.

Examples:
- namespace='openshift-ovn-kubernetes', name='ovnkube-node-abc'
- namespace='openshift-ovn-kubernetes', name='ovnkube-node-abc', container='ovnkube-controller'
- namespace='openshift-ovn-kubernetes', name='ovnkube-node-abc', pattern='ERROR.*'`, DefaultMaxLines),
		}, s.GetPodLogs)
}

// ListPlugins lists plugins with command counts
func (s *MCPServer) ListPlugins(ctx context.Context, req *mcp.CallToolRequest, in types.ListPluginsParams) (*mcp.CallToolResult, types.ListPluginsResult, error) {
	result, err := listPlugins(in.SosreportPath)
	if err != nil {
		return nil, types.ListPluginsResult{}, err
	}
	return nil, result, nil
}

// ListCommands lists all commands for a specific plugin
func (s *MCPServer) ListCommands(ctx context.Context, req *mcp.CallToolRequest, in types.ListCommandsParams) (*mcp.CallToolResult, types.ListCommandsResult, error) {
	result, err := listCommands(in.SosreportPath, in.Plugin)
	if err != nil {
		return nil, types.ListCommandsResult{}, err
	}
	return nil, result, nil
}

// SearchCommands searches for commands across all plugins
func (s *MCPServer) SearchCommands(ctx context.Context, req *mcp.CallToolRequest, in types.SearchCommandsParams) (*mcp.CallToolResult, types.SearchCommandsResult, error) {
	if in.Pattern == "" {
		return nil, types.SearchCommandsResult{}, fmt.Errorf("pattern is required")
	}
	result, err := searchCommands(in.SosreportPath, in.PatternParams, in.MaxResults)
	if err != nil {
		return nil, types.SearchCommandsResult{}, err
	}
	return nil, result, nil
}

// GetCommand retrieves command output by filepath
func (s *MCPServer) GetCommand(ctx context.Context, req *mcp.CallToolRequest, in types.GetCommandParams) (*mcp.CallToolResult, types.GetCommandResult, error) {
	output, err := getCommandOutput(in.SosreportPath, in.Filepath, in.PatternParams, in.HeadTailParams)
	if err != nil {
		return nil, types.GetCommandResult{}, err
	}
	return nil, types.GetCommandResult{Output: output}, nil
}

// GetPodLogs retrieves pod logs using the manifest
func (s *MCPServer) GetPodLogs(ctx context.Context, req *mcp.CallToolRequest, in types.GetPodLogsParams) (*mcp.CallToolResult, types.GetPodLogsResult, error) {
	output, err := getPodLogs(in.SosreportPath, in.Namespace, in.Name, in.Container, in.PatternParams, in.HeadTailParams)
	if err != nil {
		return nil, types.GetPodLogsResult{}, err
	}
	return nil, types.GetPodLogsResult{Output: output}, nil
}
