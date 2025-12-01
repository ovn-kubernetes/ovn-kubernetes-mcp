package e2e

import (
	"flag"
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	k8se2eframework "k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/config"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ipamclaimssscheme "github.com/k8snetworkplumbingwg/ipamclaims/pkg/crd/ipamclaims/v1alpha1/apis/clientset/versioned/scheme"
	multinetworkpolicyscheme "github.com/k8snetworkplumbingwg/multi-networkpolicy/pkg/client/clientset/versioned/scheme"
	networkattchmentdefscheme "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/client/clientset/versioned/scheme"
	frrscheme "github.com/metallb/frr-k8s/pkg/client/clientset/versioned/scheme"
	ocpcloudnetworkscheme "github.com/openshift/client-go/cloudnetwork/clientset/versioned/scheme"
	ocpnetworkscheme "github.com/openshift/client-go/network/clientset/versioned/scheme"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/inspector"
	adminpolicybasedroutescheme "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/adminpolicybasedroute/v1/apis/clientset/versioned/scheme"
	egressfirewallscheme "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/egressfirewall/v1/apis/clientset/versioned/scheme"
	egressipscheme "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/egressip/v1/apis/clientset/versioned/scheme"
	egressqosscheme "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/egressqos/v1/apis/clientset/versioned/scheme"
	egressservicescheme "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/egressservice/v1/apis/clientset/versioned/scheme"
	networkqosscheme "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/networkqos/v1alpha1/apis/clientset/versioned/scheme"
	routeadvertisementsscheme "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/routeadvertisements/v1/apis/clientset/versioned/scheme"
	userdefinednetworkscheme "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/userdefinednetwork/v1/apis/clientset/versioned/scheme"
	anpscheme "sigs.k8s.io/network-policy-api/pkg/client/clientset/versioned/scheme"
)

const (
	mcpServerPathEnvVar = "MCP_SERVER_PATH"
	kubeconfigEnvVar    = "KUBECONFIG"
	mcpModeEnvVar       = "MCP_MODE"
)

var (
	mcpInspector *inspector.MCPInspector
	kubeClient   client.Client
)

var s = runtime.NewScheme()

func init() {
	utilruntime.Must(scheme.AddToScheme(s))
	utilruntime.Must(egressfirewallscheme.AddToScheme(s))
	utilruntime.Must(egressipscheme.AddToScheme(s))
	utilruntime.Must(egressqosscheme.AddToScheme(s))
	utilruntime.Must(egressservicescheme.AddToScheme(s))
	utilruntime.Must(networkqosscheme.AddToScheme(s))
	utilruntime.Must(routeadvertisementsscheme.AddToScheme(s))
	utilruntime.Must(userdefinednetworkscheme.AddToScheme(s))
	utilruntime.Must(adminpolicybasedroutescheme.AddToScheme(s))
	utilruntime.Must(anpscheme.AddToScheme(s))
	utilruntime.Must(ocpcloudnetworkscheme.AddToScheme(s))
	utilruntime.Must(ocpnetworkscheme.AddToScheme(s))
	utilruntime.Must(networkattchmentdefscheme.AddToScheme(s))
	utilruntime.Must(frrscheme.AddToScheme(s))
	utilruntime.Must(ipamclaimssscheme.AddToScheme(s))
	utilruntime.Must(multinetworkpolicyscheme.AddToScheme(s))
}

// handleFlags sets up all flags and parses the command line.
func handleFlags() {
	config.CopyFlags(config.Flags, flag.CommandLine)
	k8se2eframework.RegisterCommonFlags(flag.CommandLine)
	k8se2eframework.RegisterClusterFlags(flag.CommandLine)
	flag.Parse()

	if k8se2eframework.TestContext.Provider == "" {
		k8se2eframework.TestContext.Provider = "skeleton"
	}

	var err error
	k8se2eframework.TestContext.CloudConfig.Provider, err = k8se2eframework.SetupProviderConfig(k8se2eframework.TestContext.Provider)
	Expect(err).NotTo(HaveOccurred())
}

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	handleFlags()
	RunSpecs(t, "E2e Suite")
}

var _ = BeforeSuite(func() {
	mcpServerPath := os.Getenv(mcpServerPathEnvVar)
	Expect(mcpServerPath).NotTo(BeEmpty())

	mcpMode := os.Getenv(mcpModeEnvVar)
	// Skip kubeconfig setup for offline mode tests
	// Offline tests (sosreport, must-gather) don't need a running cluster
	if mcpMode == "Offline Mode" {
		return
	}

	kubeconfig := os.Getenv(kubeconfigEnvVar)
	Expect(kubeconfig).NotTo(BeEmpty())

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	Expect(err).NotTo(HaveOccurred())
	kubeClient, err = client.New(config, client.Options{Scheme: s})
	Expect(err).NotTo(HaveOccurred())

	k8se2eframework.TestContext.KubeConfig = kubeconfig

	mcpInspector = inspector.NewMCPInspector().
		Command(mcpServerPath).
		CommandFlags(map[string]string{
			"kubeconfig": kubeconfig,
		})
})

var _ = AfterSuite(func() {
})
