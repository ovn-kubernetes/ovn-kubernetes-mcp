package client

import (
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/kubectl/pkg/cmd/exec"
)

// OVNKMCPServerClientSet is a client set for the OVN Kubernetes MCP server.
type OVNKMCPServerClientSet struct {
	clientSet                   kubernetes.Interface
	dynamicClient               dynamic.Interface
	deferredDiscoveryRESTMapper *restmapper.DeferredDiscoveryRESTMapper
	config                      *rest.Config
	corev1RestClient            rest.Interface
	podExecutor                 exec.RemoteExecutor
	debugPodNamespace           string
}

// NewOVNKMCPServerClientSet creates a new OVNKMCPServerClientSet.
func NewOVNKMCPServerClientSet(config *rest.Config, debugPodNamespace string) (*OVNKMCPServerClientSet, error) {
	if debugPodNamespace == "" {
		debugPodNamespace = "default"
	}
	clientSet := kubernetes.NewForConfigOrDie(config)
	dynamicClient := dynamic.NewForConfigOrDie(config)
	memoryClient := memory.NewMemCacheClient(clientSet.Discovery())
	deferredDiscoveryRESTMapper := restmapper.NewDeferredDiscoveryRESTMapper(memoryClient)
	return &OVNKMCPServerClientSet{
		clientSet:                   clientSet,
		dynamicClient:               dynamicClient,
		deferredDiscoveryRESTMapper: deferredDiscoveryRESTMapper,
		config:                      config,
		corev1RestClient:            clientSet.CoreV1().RESTClient(),
		podExecutor:                 &exec.DefaultRemoteExecutor{},
		debugPodNamespace:           debugPodNamespace,
	}, nil
}
