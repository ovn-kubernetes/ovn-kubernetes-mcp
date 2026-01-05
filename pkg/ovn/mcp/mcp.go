package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	kubernetesmcp "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/mcp"
	ovntypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/ovn/types"
)

// MCPServer provides OVN layer analysis tools
type MCPServer struct {
	k8sMcpServer *kubernetesmcp.MCPServer
}

// NewMCPServer creates a new OVN MCP server
func NewMCPServer(k8sMcpServer *kubernetesmcp.MCPServer) *MCPServer {
	return &MCPServer{
		k8sMcpServer: k8sMcpServer,
	}
}

// AddTools registers OVN tools with the MCP server
func (s *MCPServer) AddTools(server *mcp.Server) {
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ovn-show",
			Description: `Display a comprehensive overview of OVN configuration from either the Northbound or Southbound database.

For Northbound (nbdb): Runs 'ovn-nbctl show' and displays logical switches, logical routers, 
their ports, and connections between them.

For Southbound (sbdb): Runs 'ovn-sbctl show' and displays chassis information, port bindings,
and their relationships.

Parameters:
- namespace: Kubernetes namespace of the OVN pod (e.g., "openshift-ovn-kubernetes")
- name: Name of the pod running OVN (e.g., "ovnkube-node-xxxxx")
- database: OVN database to query - "nbdb" for Northbound or "sbdb" for Southbound
- max_lines (optional): Limit the number of output lines returned

Example output for nbdb:
{
  "database": "nbdb",
  "output": "switch 1234-5678 (node1)\n    port node1-k8s\n        addresses: [\"00:00:00:00:00:01\"]\n..."
}`,
		}, s.Show)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ovn-list-table",
			Description: `List records from a specific table in the OVN Northbound or Southbound database.

Common Northbound tables: Logical_Switch, Logical_Router, Logical_Switch_Port, 
Logical_Router_Port, ACL, Address_Set, Port_Group, Load_Balancer, NAT

Common Southbound tables: Chassis, Port_Binding, Datapath_Binding, Logical_Flow,
MAC_Binding, Multicast_Group, SB_Global

Parameters:
- namespace: Kubernetes namespace of the OVN pod
- name: Name of the pod running OVN
- database: OVN database to query - "nbdb" for Northbound or "sbdb" for Southbound
- table: Name of the table to list (e.g., "Logical_Switch", "Port_Binding")
- columns (optional): Comma-separated list of columns to display (e.g., "name,_uuid,addresses")
- filter (optional): Regex pattern to filter results
- max_lines (optional): Limit the number of records returned

Example output:
{
  "database": "nbdb",
  "table": "Logical_Switch",
  "records": [
    "_uuid: 1234-5678, name: \"node1\", ports: [...]",
    "_uuid: abcd-efgh, name: \"node2\", ports: [...]"
  ]
}`,
		}, s.ListTable)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ovn-get",
			Description: `Get a specific record or column value from an OVN database table.

When a column is specified, runs 'ovn-nbctl/ovn-sbctl get' to retrieve the specific column value.
When no column is specified, runs 'ovn-nbctl/ovn-sbctl list' to show all columns of the record.

Parameters:
- namespace: Kubernetes namespace of the OVN pod
- name: Name of the pod running OVN
- database: OVN database to query - "nbdb" for Northbound or "sbdb" for Southbound
- table: Name of the table (e.g., "Logical_Router", "Port_Binding")
- record: Record identifier (UUID or name)
- column (optional): Specific column to retrieve. If not specified, returns all columns of the record

Example getting a specific column:
{
  "database": "nbdb",
  "table": "Logical_Router",
  "record": "4c4a0a35-348c-41cc-8417-53a618e0c383",
  "column": "name",
  "output": "ovn_cluster_router"
}

Example getting entire record (no column specified):
{
  "database": "nbdb", 
  "table": "Logical_Router",
  "record": "4c4a0a35-348c-41cc-8417-53a618e0c383",
  "output": "_uuid: 4c4a0a35-348c-41cc-8417-53a618e0c383\nname: ovn_cluster_router\nports: [...]"
}`,
		}, s.Get)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ovn-ls-list",
			Description: `List all logical switches in the OVN Northbound database.

Runs 'ovn-nbctl ls-list' to retrieve all logical switches.

Parameters:
- namespace: Kubernetes namespace of the OVN pod
- name: Name of the pod running OVN
- filter (optional): Regex pattern to filter switches by name
- max_lines (optional): Limit the number of switches returned

Example output:
{
  "switches": [
    "1234-5678 (node1)",
    "abcd-efgh (join)",
    "ijkl-mnop (ext_node1)"
  ]
}`,
		}, s.ListLogicalSwitches)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ovn-lr-list",
			Description: `List all logical routers in the OVN Northbound database.

Runs 'ovn-nbctl lr-list' to retrieve all logical routers.

Parameters:
- namespace: Kubernetes namespace of the OVN pod
- name: Name of the pod running OVN
- filter (optional): Regex pattern to filter routers by name
- max_lines (optional): Limit the number of routers returned

Example output:
{
  "routers": [
    "1234-5678 (ovn_cluster_router)",
    "abcd-efgh (GR_node1)"
  ]
}`,
		}, s.ListLogicalRouters)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ovn-lsp-list",
			Description: `List logical switch ports, optionally for a specific logical switch.

Runs 'ovn-nbctl lsp-list' to retrieve logical switch ports.

Parameters:
- namespace: Kubernetes namespace of the OVN pod
- name: Name of the pod running OVN
- switch (optional): Name of the logical switch. If not specified, lists ports from all switches
- filter (optional): Regex pattern to filter ports
- max_lines (optional): Limit the number of ports returned

Example output:
{
  "switch": "node1",
  "ports": [
    "1234-5678 (pod1-port)",
    "abcd-efgh (stor-node1)",
    "ijkl-mnop (k8s-node1)"
  ]
}`,
		}, s.ListLogicalSwitchPorts)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ovn-acl-list",
			Description: `List ACLs (Access Control Lists) for a specific logical switch or port group.

Runs 'ovn-nbctl acl-list' to retrieve ACLs that control traffic filtering.
ACLs can be attached to either logical switches or port groups.

Parameters:
- namespace: Kubernetes namespace of the OVN pod
- name: Name of the pod running OVN
- entity: Name of the logical switch or port group (required)
- filter (optional): Regex pattern to filter ACLs
- max_lines (optional): Limit the number of ACLs returned

Example output:
{
  "entity": "pg_cluster_default",
  "acls": [
    "from-lport 1000 (inport == \"pod1\") allow-related",
    "to-lport 1000 (outport == \"pod1\" && ip4.src == 10.0.0.0/8) allow"
  ]
}`,
		}, s.ListACLs)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ovn-pg-list",
			Description: `List port groups in the OVN Northbound database.

Runs 'ovn-nbctl list Port_Group' to retrieve port groups which are used to group 
logical switch ports for applying ACLs.

Parameters:
- namespace: Kubernetes namespace of the OVN pod
- name: Name of the pod running OVN
- filter (optional): Regex pattern to filter port groups
- max_lines (optional): Limit the number of port groups returned

Example output:
{
  "port_groups": [
    "_uuid: 1234-5678, name: \"pg_default_deny\", ports: [...]",
    "_uuid: abcd-efgh, name: \"pg_allow_from_ns\", ports: [...]"
  ]
}`,
		}, s.ListPortGroups)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ovn-nat-list",
			Description: `List NAT rules, optionally for a specific logical router.

Runs 'ovn-nbctl lr-nat-list' to retrieve NAT rules configured on logical routers.
NAT rules handle SNAT (source NAT) and DNAT (destination NAT) for traffic.

Parameters:
- namespace: Kubernetes namespace of the OVN pod
- name: Name of the pod running OVN
- router (optional): Name of the logical router. If not specified, lists NAT rules from all routers
- filter (optional): Regex pattern to filter NAT rules
- max_lines (optional): Limit the number of NAT rules returned

Example output:
{
  "router": "GR_node1",
  "nats": [
    "TYPE             EXTERNAL_IP        EXTERNAL_PORT    LOGICAL_IP          EXTERNAL_MAC         LOGICAL_PORT",
    "snat             192.168.1.1                         10.244.0.0/24",
    "dnat_and_snat    192.168.1.100                       10.244.0.5          00:00:00:00:00:01    pod1-port"
  ]
}`,
		}, s.ListNATs)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ovn-address-set-list",
			Description: `List address sets in the OVN Northbound database.

Runs 'ovn-nbctl list Address_Set' to retrieve address sets which contain groups of 
IP addresses used in ACL rules for efficient matching.

Parameters:
- namespace: Kubernetes namespace of the OVN pod
- name: Name of the pod running OVN
- filter (optional): Regex pattern to filter address sets
- max_lines (optional): Limit the number of address sets returned

Example output:
{
  "address_sets": [
    "_uuid: 1234-5678, name: \"ns_default_v4\", addresses: [\"10.244.0.5\", \"10.244.0.6\"]",
    "_uuid: abcd-efgh, name: \"ns_kube-system_v4\", addresses: [\"10.244.1.2\", \"10.244.1.3\"]"
  ]
}`,
		}, s.ListAddressSets)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ovn-lflow-list",
			Description: `List logical flows from the OVN Southbound database.

Runs 'ovn-sbctl lflow-list' to retrieve logical flows which represent the compiled 
logical network pipeline. This is essential for debugging packet forwarding.

Parameters:
- namespace: Kubernetes namespace of the OVN pod
- name: Name of the pod running OVN
- datapath (optional): Datapath name or UUID to filter flows for a specific logical switch/router
- filter (optional): Regex pattern to filter flows
- max_lines (optional): Limit the number of flows returned

Example output:
{
  "datapath": "node1",
  "flows": [
    "table=0 (ls_in_port_sec_l2), priority=100, match=(inport == \"pod1\"), action=(next;)",
    "table=1 (ls_in_port_sec_ip), priority=90, match=(ip4), action=(next;)"
  ]
}`,
		}, s.ListLogicalFlows)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ovn-chassis-list",
			Description: `List all chassis registered in the OVN Southbound database.

Runs 'ovn-sbctl list Chassis' to retrieve information about all OVN nodes/hypervisors.

Parameters:
- namespace: Kubernetes namespace of the OVN pod
- name: Name of the pod running OVN
- filter (optional): Regex pattern to filter chassis
- max_lines (optional): Limit the number of chassis returned

Example output:
{
  "chassis": [
    "_uuid: 1234-5678, name: \"node1\", hostname: \"worker-1\", encaps: [geneve:10.0.0.1]",
    "_uuid: abcd-efgh, name: \"node2\", hostname: \"worker-2\", encaps: [geneve:10.0.0.2]"
  ]
}`,
		}, s.ListChassis)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ovn-port-binding-list",
			Description: `List port bindings from the OVN Southbound database.

Runs 'ovn-sbctl list Port_Binding' to retrieve information about where logical ports 
are bound to physical chassis.

Parameters:
- namespace: Kubernetes namespace of the OVN pod
- name: Name of the pod running OVN
- filter (optional): Regex pattern to filter port bindings
- max_lines (optional): Limit the number of bindings returned

Example output:
{
  "bindings": [
    "logical_port: \"pod1-port\", chassis: \"node1\", type: \"\", mac: [\"00:00:00:00:00:01 10.244.0.5\"]",
    "logical_port: \"stor-node1\", chassis: [], type: \"patch\", mac: [\"router\"]"
  ]
}`,
		}, s.ListPortBindings)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ovn-trace",
			Description: `Trace a packet through the OVN logical network.

Runs 'ovn-trace' to simulate packet processing through the logical network pipeline.
This shows which logical flows match, what actions are taken, and the final disposition.

The trace is essential for debugging connectivity issues and understanding how traffic
flows through the OVN logical network.

Parameters:
- namespace: Kubernetes namespace of the OVN pod
- name: Name of the pod running OVN
- datapath: Name of the logical switch or router to start the trace
- microflow: Microflow specification describing the packet (e.g., "inport==\"pod1\" && eth.src==00:00:00:00:00:01 && ip4.src==10.244.0.5 && ip4.dst==10.244.1.5")
- mode (optional): Output verbosity mode - "detailed" (default), "summary", or "minimal"
- filter (optional): Regex pattern to filter trace output
- max_lines (optional): Limit the number of output lines returned

Microflow specification examples:
- inport=="pod1" && eth.src==00:00:00:00:00:01 && ip4.src==10.244.0.5 && ip4.dst==10.244.1.5
- inport=="pod1" && eth.src==00:00:00:00:00:01 && icmp && ip4.src==10.244.0.5 && ip4.dst==8.8.8.8

Example output:
{
  "datapath": "node1",
  "microflow": "inport==\"pod1\" && ...",
  "output": "ingress(dp=\"node1\", inport=\"pod1\")\n  0. ls_in_port_sec_l2: inport == \"pod1\", priority 50, uuid 1234\n     next;\n..."
}`,
		}, s.Trace)
}

// Show displays a comprehensive overview of OVN configuration.
func (s *MCPServer) Show(ctx context.Context, req *mcp.CallToolRequest,
	in ovntypes.ShowParams) (*mcp.CallToolResult, ovntypes.ShowResult, error) {
	result := ovntypes.ShowResult{
		Database: in.Database,
	}

	// Validate database
	if err := validateDatabase(in.Database); err != nil {
		return nil, result, err
	}

	// Build command
	cmd := getDBCommand(in.Database)
	lines, err := s.runCommand(ctx, req, in.NamespacedNameParams, []string{cmd, "show"})
	if err != nil {
		return nil, result, fmt.Errorf("failed to retrieve OVN configuration from pod %s/%s: %w",
			in.Namespace, in.Name, err)
	}

	// Limit to MaxLines if specified
	lines = limitLines(lines, in.MaxLines)

	// Join all lines into a single output string
	result.Output = strings.Join(lines, "\n")
	return nil, result, nil
}

// ListTable lists records from a specific table in the OVN database.
func (s *MCPServer) ListTable(ctx context.Context, req *mcp.CallToolRequest,
	in ovntypes.ListTableParams) (*mcp.CallToolResult, ovntypes.ListTableResult, error) {
	result := ovntypes.ListTableResult{
		Database: in.Database,
		Table:    in.Table,
		Records:  []string{},
	}

	// Validate inputs
	if err := validateDatabase(in.Database); err != nil {
		return nil, result, err
	}
	if err := validateTableName(in.Table); err != nil {
		return nil, result, err
	}
	if err := validateColumnSpec(in.Columns); err != nil {
		return nil, result, err
	}

	// Build command
	// --columns flag must come before the list subcommand
	cmd := getDBCommand(in.Database)
	cmdArgs := []string{cmd}
	if in.Columns != "" {
		cmdArgs = append(cmdArgs, "--columns="+in.Columns)
	}
	cmdArgs = append(cmdArgs, "list", in.Table)

	lines, err := s.runCommand(ctx, req, in.NamespacedNameParams, cmdArgs)
	if err != nil {
		return nil, result, fmt.Errorf("failed to list table %s from pod %s/%s: %w",
			in.Table, in.Namespace, in.Name, err)
	}

	// Filter records if pattern provided
	lines, err = filterLines(lines, in.Filter)
	if err != nil {
		return nil, result, fmt.Errorf("invalid filter pattern: %w", err)
	}

	// Limit to MaxLines if specified
	lines = limitLines(lines, in.MaxLines)

	result.Records = lines
	return nil, result, nil
}

// Get retrieves a specific record or column from an OVN table.
func (s *MCPServer) Get(ctx context.Context, req *mcp.CallToolRequest,
	in ovntypes.GetParams) (*mcp.CallToolResult, ovntypes.GetResult, error) {
	result := ovntypes.GetResult{
		Database: in.Database,
		Table:    in.Table,
		Record:   in.Record,
	}

	// Validate inputs
	if err := validateDatabase(in.Database); err != nil {
		return nil, result, err
	}
	if err := validateTableName(in.Table); err != nil {
		return nil, result, err
	}
	if err := validateRecordName(in.Record); err != nil {
		return nil, result, err
	}
	if err := validateColumnSpec(in.Column); err != nil {
		return nil, result, err
	}

	// Build command
	// If column is specified, use 'get <table> <record> <column>'
	// If no column, use 'list <table> <record>' to show all columns
	cmd := getDBCommand(in.Database)
	var cmdArgs []string
	if in.Column != "" {
		cmdArgs = []string{cmd, "get", in.Table, in.Record, in.Column}
	} else {
		cmdArgs = []string{cmd, "list", in.Table, in.Record}
	}

	lines, err := s.runCommand(ctx, req, in.NamespacedNameParams, cmdArgs)
	if err != nil {
		return nil, result, fmt.Errorf("failed to get record %s from table %s on pod %s/%s: %w",
			in.Record, in.Table, in.Namespace, in.Name, err)
	}

	result.Output = strings.Join(lines, "\n")
	return nil, result, nil
}

// ListLogicalSwitches lists all logical switches.
func (s *MCPServer) ListLogicalSwitches(ctx context.Context, req *mcp.CallToolRequest,
	in ovntypes.LogicalSwitchListParams) (*mcp.CallToolResult, ovntypes.LogicalSwitchListResult, error) {
	result := ovntypes.LogicalSwitchListResult{
		Switches: []string{},
	}

	lines, err := s.runCommand(ctx, req, in.NamespacedNameParams, []string{"ovn-nbctl", "ls-list"})
	if err != nil {
		return nil, result, fmt.Errorf("failed to list logical switches from pod %s/%s: %w",
			in.Namespace, in.Name, err)
	}

	// Filter switches if pattern provided
	lines, err = filterLines(lines, in.Filter)
	if err != nil {
		return nil, result, fmt.Errorf("invalid filter pattern: %w", err)
	}

	// Limit to MaxLines if specified
	lines = limitLines(lines, in.MaxLines)

	result.Switches = lines
	return nil, result, nil
}

// ListLogicalRouters lists all logical routers.
func (s *MCPServer) ListLogicalRouters(ctx context.Context, req *mcp.CallToolRequest,
	in ovntypes.LogicalRouterListParams) (*mcp.CallToolResult, ovntypes.LogicalRouterListResult, error) {
	result := ovntypes.LogicalRouterListResult{
		Routers: []string{},
	}

	lines, err := s.runCommand(ctx, req, in.NamespacedNameParams, []string{"ovn-nbctl", "lr-list"})
	if err != nil {
		return nil, result, fmt.Errorf("failed to list logical routers from pod %s/%s: %w",
			in.Namespace, in.Name, err)
	}

	// Filter routers if pattern provided
	lines, err = filterLines(lines, in.Filter)
	if err != nil {
		return nil, result, fmt.Errorf("invalid filter pattern: %w", err)
	}

	// Limit to MaxLines if specified
	lines = limitLines(lines, in.MaxLines)

	result.Routers = lines
	return nil, result, nil
}

// ListLogicalSwitchPorts lists logical switch ports.
func (s *MCPServer) ListLogicalSwitchPorts(ctx context.Context, req *mcp.CallToolRequest,
	in ovntypes.LogicalSwitchPortListParams) (*mcp.CallToolResult, ovntypes.LogicalSwitchPortListResult, error) {
	result := ovntypes.LogicalSwitchPortListResult{
		Switch: in.Switch,
		Ports:  []string{},
	}

	// Validate switch name if provided
	if err := validateSwitchName(in.Switch); err != nil {
		return nil, result, err
	}

	// Build command
	cmdArgs := []string{"ovn-nbctl", "lsp-list"}
	if in.Switch != "" {
		cmdArgs = append(cmdArgs, in.Switch)
	}

	lines, err := s.runCommand(ctx, req, in.NamespacedNameParams, cmdArgs)
	if err != nil {
		return nil, result, fmt.Errorf("failed to list logical switch ports from pod %s/%s: %w",
			in.Namespace, in.Name, err)
	}

	// Filter ports if pattern provided
	lines, err = filterLines(lines, in.Filter)
	if err != nil {
		return nil, result, fmt.Errorf("invalid filter pattern: %w", err)
	}

	// Limit to MaxLines if specified
	lines = limitLines(lines, in.MaxLines)

	result.Ports = lines
	return nil, result, nil
}

// ListACLs lists ACLs.
func (s *MCPServer) ListACLs(ctx context.Context, req *mcp.CallToolRequest,
	in ovntypes.ACLListParams) (*mcp.CallToolResult, ovntypes.ACLListResult, error) {
	result := ovntypes.ACLListResult{
		Entity: in.Entity,
		ACLs:   []string{},
	}

	// Entity is required - ovn-nbctl acl-list requires a switch or port group name
	if in.Entity == "" {
		return nil, result, fmt.Errorf("entity (logical switch or port group name) is required")
	}

	// Validate entity name (switch or port group)
	if err := validateEntityName(in.Entity); err != nil {
		return nil, result, err
	}

	// Build command
	cmdArgs := []string{"ovn-nbctl", "acl-list", in.Entity}

	lines, err := s.runCommand(ctx, req, in.NamespacedNameParams, cmdArgs)
	if err != nil {
		return nil, result, fmt.Errorf("failed to list ACLs from pod %s/%s: %w",
			in.Namespace, in.Name, err)
	}

	// Filter ACLs if pattern provided
	lines, err = filterLines(lines, in.Filter)
	if err != nil {
		return nil, result, fmt.Errorf("invalid filter pattern: %w", err)
	}

	// Limit to MaxLines if specified
	lines = limitLines(lines, in.MaxLines)

	result.ACLs = lines
	return nil, result, nil
}

// ListPortGroups lists port groups from the Northbound database.
func (s *MCPServer) ListPortGroups(ctx context.Context, req *mcp.CallToolRequest,
	in ovntypes.PortGroupListParams) (*mcp.CallToolResult, ovntypes.PortGroupListResult, error) {
	result := ovntypes.PortGroupListResult{
		PortGroups: []string{},
	}

	lines, err := s.runCommand(ctx, req, in.NamespacedNameParams, []string{"ovn-nbctl", "list", "Port_Group"})
	if err != nil {
		return nil, result, fmt.Errorf("failed to list port groups from pod %s/%s: %w",
			in.Namespace, in.Name, err)
	}

	// Filter port groups if pattern provided
	lines, err = filterLines(lines, in.Filter)
	if err != nil {
		return nil, result, fmt.Errorf("invalid filter pattern: %w", err)
	}

	// Limit to MaxLines if specified
	lines = limitLines(lines, in.MaxLines)

	result.PortGroups = lines
	return nil, result, nil
}

// ListNATs lists NAT rules, optionally for a specific logical router.
func (s *MCPServer) ListNATs(ctx context.Context, req *mcp.CallToolRequest,
	in ovntypes.NATListParams) (*mcp.CallToolResult, ovntypes.NATListResult, error) {
	result := ovntypes.NATListResult{
		Router: in.Router,
		NATs:   []string{},
	}

	// Validate router name if provided
	if err := validateRouterName(in.Router); err != nil {
		return nil, result, err
	}

	var lines []string
	var err error

	if in.Router != "" {
		// List NAT rules for specific router
		lines, err = s.runCommand(ctx, req, in.NamespacedNameParams, []string{"ovn-nbctl", "lr-nat-list", in.Router})
	} else {
		// List all NAT rules from NAT table
		lines, err = s.runCommand(ctx, req, in.NamespacedNameParams, []string{"ovn-nbctl", "list", "NAT"})
	}

	if err != nil {
		return nil, result, fmt.Errorf("failed to list NAT rules from pod %s/%s: %w",
			in.Namespace, in.Name, err)
	}

	// Filter NAT rules if pattern provided
	lines, err = filterLines(lines, in.Filter)
	if err != nil {
		return nil, result, fmt.Errorf("invalid filter pattern: %w", err)
	}

	// Limit to MaxLines if specified
	lines = limitLines(lines, in.MaxLines)

	result.NATs = lines
	return nil, result, nil
}

// ListAddressSets lists address sets from the Northbound database.
func (s *MCPServer) ListAddressSets(ctx context.Context, req *mcp.CallToolRequest,
	in ovntypes.AddressSetListParams) (*mcp.CallToolResult, ovntypes.AddressSetListResult, error) {
	result := ovntypes.AddressSetListResult{
		AddressSets: []string{},
	}

	lines, err := s.runCommand(ctx, req, in.NamespacedNameParams, []string{"ovn-nbctl", "list", "Address_Set"})
	if err != nil {
		return nil, result, fmt.Errorf("failed to list address sets from pod %s/%s: %w",
			in.Namespace, in.Name, err)
	}

	// Filter address sets if pattern provided
	lines, err = filterLines(lines, in.Filter)
	if err != nil {
		return nil, result, fmt.Errorf("invalid filter pattern: %w", err)
	}

	// Limit to MaxLines if specified
	lines = limitLines(lines, in.MaxLines)

	result.AddressSets = lines
	return nil, result, nil
}

// ListLogicalFlows lists logical flows from the Southbound database.
func (s *MCPServer) ListLogicalFlows(ctx context.Context, req *mcp.CallToolRequest,
	in ovntypes.LogicalFlowListParams) (*mcp.CallToolResult, ovntypes.LogicalFlowListResult, error) {
	result := ovntypes.LogicalFlowListResult{
		Datapath: in.Datapath,
		Flows:    []string{},
	}

	// Validate datapath if provided
	if in.Datapath != "" {
		if err := validateDatapath(in.Datapath); err != nil {
			return nil, result, err
		}
	}

	// Build command
	cmdArgs := []string{"ovn-sbctl", "lflow-list"}
	if in.Datapath != "" {
		cmdArgs = append(cmdArgs, in.Datapath)
	}

	lines, err := s.runCommand(ctx, req, in.NamespacedNameParams, cmdArgs)
	if err != nil {
		return nil, result, fmt.Errorf("failed to list logical flows from pod %s/%s: %w",
			in.Namespace, in.Name, err)
	}

	// Filter flows if pattern provided
	lines, err = filterLines(lines, in.Filter)
	if err != nil {
		return nil, result, fmt.Errorf("invalid filter pattern: %w", err)
	}

	// Limit to MaxLines if specified
	lines = limitLines(lines, in.MaxLines)

	result.Flows = lines
	return nil, result, nil
}

// ListChassis lists all chassis from the Southbound database.
func (s *MCPServer) ListChassis(ctx context.Context, req *mcp.CallToolRequest,
	in ovntypes.ChassisListParams) (*mcp.CallToolResult, ovntypes.ChassisListResult, error) {
	result := ovntypes.ChassisListResult{
		Chassis: []string{},
	}

	lines, err := s.runCommand(ctx, req, in.NamespacedNameParams, []string{"ovn-sbctl", "list", "Chassis"})
	if err != nil {
		return nil, result, fmt.Errorf("failed to list chassis from pod %s/%s: %w",
			in.Namespace, in.Name, err)
	}

	// Filter chassis if pattern provided
	lines, err = filterLines(lines, in.Filter)
	if err != nil {
		return nil, result, fmt.Errorf("invalid filter pattern: %w", err)
	}

	// Limit to MaxLines if specified
	lines = limitLines(lines, in.MaxLines)

	result.Chassis = lines
	return nil, result, nil
}

// ListPortBindings lists port bindings from the Southbound database.
func (s *MCPServer) ListPortBindings(ctx context.Context, req *mcp.CallToolRequest,
	in ovntypes.PortBindingListParams) (*mcp.CallToolResult, ovntypes.PortBindingListResult, error) {
	result := ovntypes.PortBindingListResult{
		Bindings: []string{},
	}

	lines, err := s.runCommand(ctx, req, in.NamespacedNameParams, []string{"ovn-sbctl", "list", "Port_Binding"})
	if err != nil {
		return nil, result, fmt.Errorf("failed to list port bindings from pod %s/%s: %w",
			in.Namespace, in.Name, err)
	}

	// Filter bindings if pattern provided
	lines, err = filterLines(lines, in.Filter)
	if err != nil {
		return nil, result, fmt.Errorf("invalid filter pattern: %w", err)
	}

	// Limit to MaxLines if specified
	lines = limitLines(lines, in.MaxLines)

	result.Bindings = lines
	return nil, result, nil
}

// Trace traces a packet through the OVN logical network.
func (s *MCPServer) Trace(ctx context.Context, req *mcp.CallToolRequest,
	in ovntypes.OVNTraceParams) (*mcp.CallToolResult, ovntypes.OVNTraceResult, error) {
	result := ovntypes.OVNTraceResult{
		Datapath:  in.Datapath,
		Microflow: in.Microflow,
	}

	// Validate inputs
	if err := validateDatapath(in.Datapath); err != nil {
		return nil, result, err
	}
	if err := validateMicroflow(in.Microflow); err != nil {
		return nil, result, err
	}

	// Build command: ovn-trace <datapath> '<microflow>'
	cmdArgs := []string{"ovn-trace"}

	// Add output format flag based on mode (default to detailed)
	switch in.Mode {
	case ovntypes.TraceModeSummary:
		cmdArgs = append(cmdArgs, "--summary")
	case ovntypes.TraceModeMinimal:
		cmdArgs = append(cmdArgs, "--minimal")
	case ovntypes.TraceModeDetailed, "":
		cmdArgs = append(cmdArgs, "--detailed")
	}

	cmdArgs = append(cmdArgs, in.Datapath, in.Microflow)

	lines, err := s.runCommand(ctx, req, in.NamespacedNameParams, cmdArgs)
	if err != nil {
		return nil, result, fmt.Errorf("failed to trace packet on pod %s/%s: %w",
			in.Namespace, in.Name, err)
	}

	// Filter lines if pattern provided
	lines, err = filterLines(lines, in.Filter)
	if err != nil {
		return nil, result, fmt.Errorf("invalid filter pattern: %w", err)
	}

	// Limit to MaxLines if specified
	lines = limitLines(lines, in.MaxLines)

	result.Output = strings.Join(lines, "\n")
	return nil, result, nil
}
