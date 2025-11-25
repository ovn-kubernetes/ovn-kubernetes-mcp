package external_tools

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/types"
	kubernetesmcp "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/mcp"
)

type MCPServer struct {
	k8sMcpServer *kubernetesmcp.MCPServer
}

func NewMCPServer(k8sMcpServer *kubernetesmcp.MCPServer) (*MCPServer, error) {
	return &MCPServer{
		k8sMcpServer: k8sMcpServer,
	}, nil
}

func (s *MCPServer) AddTools(server *mcp.Server) {
	s.registerIPTools(server)
	s.registerSocketTools(server)
	s.registerFirewallTools(server)
	s.registerConnectivityTools(server)
	s.registerStatsTools(server)
	s.registerCaptureTools(server)
}

// executeCommand runs a command on a target (node or pod)
func (s *MCPServer) executeCommand(ctx context.Context, target types.TargetParams, command []string) (string, error) {
	clientSet := s.k8sMcpServer.GetClientSet()

	switch target.TargetType {
	case "node":
		if target.NodeName == "" {
			return "", fmt.Errorf("node_name is required when target_type is 'node'")
		}
		if target.NodeImage == "" {
			return "", fmt.Errorf("node_image is required when target_type is 'node'")
		}
		chrootCommand := append([]string{"chroot", "/host"}, command...)
		stdout, stderr, err := clientSet.DebugNode(ctx, target.NodeName, target.NodeImage, chrootCommand)
		if err != nil {
			return stdout + "\n" + stderr, err
		}
		return stdout, nil

	case "pod":
		if target.PodName == "" {
			return "", fmt.Errorf("pod_name is required when target_type is 'pod'")
		}
		namespace := target.PodNamespace
		if namespace == "" {
			namespace = "default"
		}
		stdout, stderr, err := clientSet.ExecPod(ctx, target.PodName, namespace, target.ContainerName, command)
		if err != nil {
			return stdout + "\n" + stderr, err
		}
		return stdout, nil

	default:
		return "", fmt.Errorf("invalid target_type: %s (must be 'node' or 'pod')", target.TargetType)
	}
}
