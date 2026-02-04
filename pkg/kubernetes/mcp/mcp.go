package kubernetes

import (
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/client"
)

type Config struct {
	Kubeconfig string
}

type MCPServer struct {
	clientSet *client.OVNKMCPServerClientSet
}

func NewMCPServer(cfg Config) (*MCPServer, error) {
	var config *rest.Config
	var err error
	if cfg.Kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
		if err != nil {
			return nil, err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	clientSet, err := client.NewOVNKMCPServerClientSet(config)
	if err != nil {
		return nil, err
	}

	return &MCPServer{
		clientSet: clientSet,
	}, nil
}

func (s *MCPServer) AddTools(server *mcp.Server) {
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "pod-logs",
			Description: fmt.Sprintf(`Get container logs from a pod in the Kubernetes cluster.

Retrieves logs from a running or terminated pod. Supports filtering by pattern,
limiting output with head/tail, and retrieving logs from previous container instances.

Parameters:
- name (required): Name of the pod
- namespace (optional): Namespace of the pod (defaults to "default")
- container (optional): Specific container name (required for multi-container pods)
- previous (optional): If true, get logs from previous container instance (useful for crash analysis)
- pattern (optional): Regex pattern to filter log lines (grep-style filtering)
- head (optional): Return only first N lines. Default: %d lines if tail is not specified
- tail (optional): Return only last N lines
- apply_tail_first (optional): If both head and tail are set and apply_tail_first is true,
apply tail before head. Default: false

Returns log lines as an array. If neither head nor tail is specified, returns the first %d lines
by default. If either of those are set, only the one that is set will be applied. If both are set,
apply_tail_first will be used to determine the order. Use pattern matching to search for specific
errors, warnings, or events in the logs.

Examples:
- Get pod logs (defaults to first %d lines): {"name": "my-pod", "namespace": "default"}
- Get specific container logs (defaults to first %d lines): {"name": "my-pod", "namespace": "default", "container": "my-container"}
- Get previous container logs (defaults to first %d lines): {"name": "my-pod", "namespace": "default", "previous": true}
- Search for errors (defaults to first %d lines): {"name": "my-pod", "namespace": "default", "pattern": "error|Error|ERROR"}
- Get last 50 lines: {"name": "my-pod", "namespace": "default", "tail": 50}
- Get first 100 lines: {"name": "my-pod", "namespace": "default", "head": 100}
- Get first 50 lines of the last 100 lines: {"name": "my-pod", "namespace": "default", "head": 50, "tail": 100, "apply_tail_first": true}
- Get last 50 lines of the first 100 lines: {"name": "my-pod", "namespace": "default", "head": 100, "tail": 50, "apply_tail_first": false}`,
				defaultMaxLines, defaultMaxLines, defaultMaxLines, defaultMaxLines, defaultMaxLines, defaultMaxLines),
		}, s.GetPodLogs)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "resource-get",
			Description: `Get a specific Kubernetes resource by name.

Retrieves a single resource from the cluster using its group, version, kind, and name.
Supports different output formats for viewing the resource data.

Parameters:
- group (optional): API group of the resource (e.g., "apps", "networking.k8s.io"). Empty for core resources
- version (required): API version of the resource (e.g., "v1", "v1beta1")
- kind (required): Kind of the resource (e.g., "Pod", "Service", "Deployment", "ConfigMap")
- name (required): Name of the resource to retrieve
- namespace (optional): Namespace of the resource. If omitted, defaults to "default" for namespaced resources and empty for cluster-scoped resources
- output_type (optional): Output format - 'yaml', 'json', or 'wide' (default: table format with name, namespace, age. wide will include labels and annotations)

Returns the resource definition in the requested format. Use this to inspect specific
resource configurations, status, and metadata.

Examples:
- Get a pod (defaults to "default" namespace): {"version": "v1", "kind": "Pod", "name": "my-pod"}
- Get a deployment as YAML: {"group": "apps", "version": "v1", "kind": "Deployment", "name": "my-deployment", "namespace": "default", "output_type": "yaml"}
- Get a node (cluster-scoped): {"version": "v1", "kind": "Node", "name": "worker-0"}
- Get a config map as JSON: {"version": "v1", "kind": "ConfigMap", "name": "my-config", "namespace": "kube-system", "output_type": "json"}
- Get detailed info: {"version": "v1", "kind": "Pod", "name": "my-pod", "namespace": "default", "output_type": "wide"}`,
		}, s.GetResource)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "resource-list",
			Description: `List Kubernetes resources of a specific kind.

Lists resources in a namespace or across all namespaces. Supports filtering by
label selector and different output formats.

Parameters:
- group (optional): API group of the resource (e.g., "apps", "networking.k8s.io"). Empty for core resources
- version (required): API version of the resource (e.g., "v1", "v1beta1")
- kind (required): Kind of the resource (e.g., "Pod", "Service", "Deployment")
- namespace (optional): Filter by namespace. If omitted, lists resources across all namespaces
- label_selector (optional): Filter by label selector (e.g., "app=my-app", "component=network")
- output_type (optional): Output format - 'yaml', 'json', or 'wide' (default: table format with name, namespace, age. wide will include labels and annotations)

Returns a list of matching resources. Use this to discover what resources exist in the
cluster before retrieving specific ones with resource-get.

Examples:
- List all pods in a namespace: {"version": "v1", "kind": "Pod", "namespace": "default"}
- List all services across namespaces: {"version": "v1", "kind": "Service"}
- List pods by label: {"version": "v1", "kind": "Pod", "namespace": "default", "label_selector": "app=my-app"}
- List deployments as YAML: {"group": "apps", "version": "v1", "kind": "Deployment", "namespace": "default", "output_type": "yaml"}
- List all nodes: {"version": "v1", "kind": "Node"}
- List with detailed info: {"version": "v1", "kind": "Pod", "namespace": "kube-system", "output_type": "wide"}`,
		}, s.ListResources)
}
