package mcp

import (
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	omcclient "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/must-gather/omc-client"
	ovsdbtool "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/must-gather/ovsdb-tool"
)

// MustGatherMCPServer is a server for the must gather MCP.
type MustGatherMCPServer struct {
	omcClient *omcclient.OmcClient
	ovsdbTool *ovsdbtool.OvsdbTool
}

// NewMCPServer creates a new MustGatherMCPServer. It will return an error if the omc client
// cannot be created.
func NewMCPServer() (*MustGatherMCPServer, error) {
	omcClient, err := omcclient.NewOmcClient()
	if err != nil {
		return nil, err
	}
	ovsdbTool, err := ovsdbtool.NewOvsdbTool(omcClient)
	if err != nil {
		log.Printf("Failed to create ovsdb tool, will not be able to use ovsdb-tool tools: %v", err)
	}
	return &MustGatherMCPServer{
		omcClient: omcClient,
		ovsdbTool: ovsdbTool,
	}, nil
}

// AddTools registers must gather tools with the MCP server
func (s *MustGatherMCPServer) AddTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "must-gather-get-resource",
		Description: `Get a specific Kubernetes resource from a must-gather archive.

Parameters:
- must_gather_path (required): Absolute path to extracted must-gather directory
- kind (required): Kubernetes resource kind (e.g., Pod, Service, Node, Deployment, ConfigMap)
- name (required): Name of the resource to retrieve
- namespace (optional): Namespace of the resource (defaults to "default" for namespaced resources)
- outputType (optional): Output format - 'yaml', 'json', or 'wide' (default: table format)

Returns the resource definition in the requested format. Use this to inspect specific
resource configurations, status, and metadata from the must-gather snapshot.

Examples:
- Get a pod: {"must_gather_path": "/path/to/must-gather", "kind": "Pod", "namespace": "default", "name": "my-pod"}
- Get a node as YAML: {"must_gather_path": "/path/to/must-gather", "kind": "Node", "name": "worker-0", "outputType": "yaml"}
- Get a config map: {"must_gather_path": "/path/to/must-gather", "kind": "ConfigMap", "namespace": "kube-system", "name": "my-config"}`,
	}, s.GetResource)

	mcp.AddTool(server, &mcp.Tool{
		Name: "must-gather-list-resources",
		Description: `List Kubernetes resources of a specific kind from a must-gather archive.

Parameters:
- must_gather_path (required): Absolute path to extracted must-gather directory
- kind (required): Kubernetes resource kind (e.g., Pod, Service, Node, Deployment)
- namespace (optional): Filter by namespace. If omitted, lists resources across all namespaces
- labelSelector (optional): Filter by label selector (e.g., 'app=ovnkube-node', 'component=network')
- outputType (optional): Output format - 'yaml', 'json', or 'wide' (default: table format)

Returns a list of matching resources. Use this to discover what resources exist in the
must-gather snapshot before retrieving specific ones with must-gather-get-resource.

Examples:
- List all pods in a namespace: {"must_gather_path": "/path/to/must-gather", "kind": "Pod", "namespace": "default"}
- List all nodes: {"must_gather_path": "/path/to/must-gather", "kind": "Node"}
- List pods by label: {"must_gather_path": "/path/to/must-gather", "kind": "Pod", "labelSelector": "app=my-app"}
- List services as JSON: {"must_gather_path": "/path/to/must-gather", "kind": "Service", "namespace": "kube-system", "outputType": "json"}`,
	}, s.ListResources)

	mcp.AddTool(server, &mcp.Tool{
		Name: "must-gather-pod-logs",
		Description: fmt.Sprintf(`Get container logs from a pod in a must-gather archive.

Parameters:
- must_gather_path (required): Absolute path to extracted must-gather directory
- name (required): Name of the pod
- namespace (optional): Namespace of the pod (defaults to "default" namespace)
- container (optional): Specific container name (required for multi-container pods)
- previous (optional): If true, get logs from previous container instance (useful for crash analysis)
- rotated (optional): If true, include rotated log files
- pattern (optional): Regex pattern to filter log lines (grep-style filtering)
- head (optional): Return only first N lines (cannot be used with tail). Default: %d if tail is not specified
- tail (optional): Return only last N lines (cannot be used with head)

Returns log lines as an array. If neither head nor tail is specified, returns the first %d lines
by default. Use pattern matching to search for specific errors, warnings, or events in the logs.

Examples:
- Get pod logs: {"must_gather_path": "/path/to/must-gather", "namespace": "default", "name": "my-pod"}
- Get specific container logs: {"must_gather_path": "/path/to/must-gather", "namespace": "default", "name": "my-pod", "container": "my-container"}
- Search for errors: {"must_gather_path": "/path/to/must-gather", "namespace": "default", "name": "my-pod", "pattern": "error|Error|ERROR"}
- Get last 50 lines: {"must_gather_path": "/path/to/must-gather", "namespace": "default", "name": "my-pod", "tail": 50}
- Get previous container logs: {"must_gather_path": "/path/to/must-gather", "namespace": "default", "name": "my-pod", "previous": true}`, defaultMaxLines, defaultMaxLines),
	}, s.GetPodLogs)

	mcp.AddTool(server, &mcp.Tool{
		Name: "must-gather-ovnk-info",
		Description: `Get OVN-Kubernetes networking information from a must-gather archive.

Parameters:
- must_gather_path (required): Absolute path to extracted must-gather directory
- info_type (required): Type of OVN-K info to retrieve. Valid values:
  - 'extrainfo': Get extra OVN-Kubernetes debugging information
  - 'hostnetinfo': Get host networking configuration and status
  - 'subnets': Get OVN subnet allocations and network topology

Use this to retrieve high-level OVN-Kubernetes networking information that helps
diagnose network configuration and connectivity issues.

Examples:
- Get subnet info: {"must_gather_path": "/path/to/must-gather", "info_type": "subnets"}
- Get host network info: {"must_gather_path": "/path/to/must-gather", "info_type": "hostnetinfo"}
- Get extra debugging info: {"must_gather_path": "/path/to/must-gather", "info_type": "extrainfo"}`,
	}, s.GetOvnKInfo)

	// Add ovsdb-tool tools if the ovsdb-tool binary is available in the PATH
	if s.ovsdbTool != nil {
		mcp.AddTool(server, &mcp.Tool{
			Name: "must-gather-list-northbound-databases",
			Description: `List OVN Northbound database files available in a must-gather archive.

Parameters:
- must_gather_path (required): Absolute path to extracted must-gather directory

Returns a mapping of database files to their source nodes. The Northbound database (nbdb)
contains the logical network configuration: logical switches, routers, ports, ACLs, and
load balancers.

Use this to discover available database files before querying with must-gather-query-database.
Each OVN controller node may have its own database snapshot.

Example:
- {"must_gather_path": "/path/to/must-gather"}

Output format:
Database        Node
ovnkube-node-abc123_nbdb    worker-0
ovnkube-node-def456_nbdb    worker-1`,
		}, s.ListNorthboundDatabases)

		mcp.AddTool(server, &mcp.Tool{
			Name: "must-gather-list-southbound-databases",
			Description: `List OVN Southbound database files available in a must-gather archive.

Parameters:
- must_gather_path (required): Absolute path to extracted must-gather directory

Returns a mapping of database files to their source nodes. The Southbound database (sbdb)
contains the physical network bindings: chassis info, port bindings, MAC bindings, and
datapath flows.

Use this to discover available database files before querying with must-gather-query-database.
Each OVN controller node may have its own database snapshot.

Example:
- {"must_gather_path": "/path/to/must-gather"}

Output format:
Database        Node
ovnkube-node-abc123_sbdb    worker-0
ovnkube-node-def456_sbdb    worker-1`,
		}, s.ListSouthboundDatabases)

		mcp.AddTool(server, &mcp.Tool{
			Name: "must-gather-query-database",
			Description: `Query an OVN database from a must-gather archive using ovsdb-tool.

Parameters:
- must_gather_path (required): Absolute path to extracted must-gather directory
- database_name (required): Database file name from must-gather-list-northbound-databases or
  must-gather-list-southbound-databases. Must end with '_nbdb' or '_sbdb'
- table (required): OVN database table to query. Common tables:
  - Northbound: Logical_Switch, Logical_Router, Logical_Switch_Port, ACL, Load_Balancer, NAT
  - Southbound: Chassis, Port_Binding, MAC_Binding, Datapath_Binding, SB_Global
- conditions (optional): Array of condition strings in OVSDB format: ["column","op","value"]. If omitted, returns all rows
  Example: ["[\"hostname\",\"==\",\"worker-0\"]"]
- columns (optional): Array of column names to return. If omitted, returns all columns

Returns query results in JSON format. Use this to inspect OVN database state for debugging
network connectivity, policy enforcement, and load balancing issues.

Examples:
- List all chassis: {"must_gather_path": "/path/to/must-gather", "database_name": "ovnkube-node-abc123_sbdb", "table": "Chassis"}
- Get specific chassis: {"must_gather_path": "/path/to/must-gather", "database_name": "ovnkube-node-abc123_sbdb", "table": "Chassis", "conditions": ["[\"hostname\",\"==\",\"worker-0\"]"]}
- List logical switches with specific columns: {"must_gather_path": "/path/to/must-gather", "database_name": "ovnkube-node-abc123_nbdb", "table": "Logical_Switch", "columns": ["name", "ports"]}
- Query port bindings: {"must_gather_path": "/path/to/must-gather", "database_name": "ovnkube-node-abc123_sbdb", "table": "Port_Binding", "columns": ["logical_port", "chassis", "type"]}`,
		}, s.QueryDatabase)
	}
}
