package kubernetes

import (
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
			Description: `Get the logs of a pod. Examples:` +
				`Get the logs of a my-pod in the default namespace: {"name": "my-pod", "namespace": "default"}` +
				`Get the logs of a my-pod in the default namespace with the my-container container and previous logs: {"name": "my-pod", "namespace": "default", "container": "my-container", "previous": true}`,
		}, s.GetPodLogs)
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "resource-get",
			Description: `Get a resource. Examples:` +
				`Get a pod named my-pod in the default namespace: {"version": "v1", "kind": "Pod", "name": "my-pod", "namespace": "default"}` +
				`Get a deployment named my-deployment in the default namespace in YAML format: {"group": "apps", "version": "v1", "kind": "Deployment", "name": "my-deployment", "namespace": "default", "outputType": "yaml"}`,
		}, s.GetResource)
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "resource-list",
			Description: `List resources in a namespace or across all namespaces. Examples:` +
				`Get all pods in the default namespace: {"version": "v1", "kind": "Pod", "namespace": "default"}` +
				`Get all services: {"version": "v1", "kind": "Service"}`,
		}, s.ListResources)
}

func (s *MCPServer) GetClientSet() *client.OVNKMCPServerClientSet {
	return s.clientSet
}
