package e2e

import (
	"context"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/ovs/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

var _ = Describe("OVS Tools", func() {
	const (
		ovsListBrToolName              = "ovs-list-br"
		ovsListPortsToolName           = "ovs-list-ports"
		ovsListIfacesToolName          = "ovs-list-ifaces"
		ovsVsctlShowToolName           = "ovs-vsctl-show"
		ovsOfctlDumpFlowsToolName      = "ovs-ofctl-dump-flows"
		ovsAppctlDumpConntrackToolName = "ovs-appctl-dump-conntrack"
		ovsAppctlOfprotoTraceToolName  = "ovs-appctl-ofproto-trace"
	)

	// Common OVN-Kubernetes namespaces to search for ovnkube-node pods
	var ovnNamespaces = []string{
		"openshift-ovn-kubernetes",
		"ovn-kubernetes",
	}

	// Suite-scoped variables for pod discovery
	var namespace, podName string

	// findOVNKubeNodePod finds a running ovnkube-node pod with bounded retry
	findOVNKubeNodePod := func() (namespace, name string) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		
		var lastErr error
		
		for {
			for _, ns := range ovnNamespaces {
				podList := &corev1.PodList{}
				err := kubeClient.List(ctx, podList,
					client.InNamespace(ns),
					client.MatchingLabels{"app": "ovnkube-node"},
				)
				if err != nil {
					lastErr = err
					continue
				}
				for _, pod := range podList.Items {
					if pod.Status.Phase != corev1.PodRunning {
						continue
					}
					ready := false
					for _, c := range pod.Status.Conditions {
						if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
							ready = true
							break
						}
					}
					if ready {
						return ns, pod.Name
					}
				}
			}
			
			// Check if context deadline exceeded
			select {
			case <-ctx.Done():
				if lastErr != nil {
					GinkgoWriter.Printf("Pod discovery failed after timeout. Last error: %v\n", lastErr)
				}
				return "", ""
			default:
				time.Sleep(2 * time.Second)
			}
		}
	}

	// BeforeEach finds the ovnkube-node pod once before each test
	BeforeEach(func() {
		By("Finding an ovnkube-node pod")
		namespace, podName = findOVNKubeNodePod()
		Expect(namespace).NotTo(BeEmpty(), "No running ovnkube-node pod found")
		Expect(podName).NotTo(BeEmpty(), "No running ovnkube-node pod found")
	})

	Context("OVS List Bridges", func() {
		It("should list OVS bridges", func() {
			By("Running ovs-list-br")
			output, err := mcpInspector.
				MethodCall(ovsListBrToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains bridge information")
			bridgeResult := utils.UnmarshalCallToolResult[types.BridgeResult](output)
			Expect(bridgeResult.Bridges).NotTo(BeEmpty())

			By("Verifying OVS bridges output")
			GinkgoWriter.Print("\n=== OVS Bridges ===\n")
			GinkgoWriter.Print("Node: ", podName, "\n")
			GinkgoWriter.Print("Bridges:\n")
			for _, bridge := range bridgeResult.Bridges {
				GinkgoWriter.Print("  ", bridge, "\n")
			}
			GinkgoWriter.Print("===================\n\n")

			// OVS bridges must include br-int (always) and an external bridge
			// The external bridge is br-ex in production but breth0 in Kind clusters
			bridgeNames := strings.Join(bridgeResult.Bridges, " ")
			Expect(bridgeNames).To(And(
				ContainSubstring("br-int"),
				Or(
					ContainSubstring("br-ex"),
					ContainSubstring("breth0"),
				),
			))
		})
	})

	Context("OVS List Ports", func() {
		It("should list OVS ports for br-int bridge", func() {
			By("Running ovs-list-ports for br-int")
			output, err := mcpInspector.
				MethodCall(ovsListPortsToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"bridge":    "br-int",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains port information")
			portResult := utils.UnmarshalCallToolResult[types.PortResult](output)
			Expect(portResult.Ports).NotTo(BeEmpty())

			By("Verifying OVS ports output")
			GinkgoWriter.Print("\n=== OVS Ports (br-int) ===\n")
			GinkgoWriter.Print("Node: ", podName, "\n")
			GinkgoWriter.Print("Ports:\n")
			for _, port := range portResult.Ports {
				GinkgoWriter.Print("  ", port, "\n")
			}
			GinkgoWriter.Print("==========================\n\n")

			// br-int ports typically include patch ports and container interfaces
			portNames := strings.Join(portResult.Ports, " ")
			Expect(portNames).To(Or(
				ContainSubstring("patch"),
				ContainSubstring("ovn-k8s-mp0"),
				MatchRegexp(`[a-f0-9]{15}`), // Container interface names (15-char hex)
			))
		})
	})

	Context("OVS List Interfaces", func() {
		It("should list OVS interfaces for br-int bridge", func() {
			By("Running ovs-list-ifaces for br-int")
			output, err := mcpInspector.
				MethodCall(ovsListIfacesToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"bridge":    "br-int",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains interface information")
			interfaceResult := utils.UnmarshalCallToolResult[types.InterfaceResult](output)
			Expect(interfaceResult.Interfaces).NotTo(BeNil())
			Expect(interfaceResult.Interfaces).NotTo(BeEmpty())

			By("Verifying OVS interfaces output")
			GinkgoWriter.Print("\n=== OVS Interfaces (br-int) ===\n")
			GinkgoWriter.Print("Node: ", podName, "\n")
			GinkgoWriter.Print("Interfaces:\n")
			if interfaceResult.Interfaces != nil {
				for _, iface := range interfaceResult.Interfaces {
					GinkgoWriter.Print("  ", iface, "\n")
				}
			}
			GinkgoWriter.Print("======================\n\n")

			// OVS interfaces should contain various interface types
			interfaceNames := strings.Join(interfaceResult.Interfaces, " ")
			Expect(interfaceNames).To(Or(
				ContainSubstring("br-int"),
				ContainSubstring("patch"),
				ContainSubstring("genev"),
			))
		})
	})

	Context("OVS Vsctl Show", func() {
		It("should show OVS database configuration", func() {
			By("Running ovs-vsctl-show")
			output, err := mcpInspector.
				MethodCall(ovsVsctlShowToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains OVS configuration")
			showResult := utils.UnmarshalCallToolResult[types.ShowResult](output)
			Expect(showResult.Output).NotTo(BeEmpty())

			By("Verifying OVS vsctl show output")
			GinkgoWriter.Print("\n=== OVS Database Configuration ===\n")
			GinkgoWriter.Print("Node: ", podName, "\n")
			GinkgoWriter.Print("Configuration:\n", showResult.Output, "\n")
			GinkgoWriter.Print("==================================\n\n")

			// OVS show output must contain all three: Bridge, Port, and Interface (nested hierarchy)
			Expect(showResult.Output).To(And(
				ContainSubstring("Bridge"),
				ContainSubstring("Port"),
				ContainSubstring("Interface"),
			))
		})
	})

	Context("OVS Ofctl Dump Flows", func() {
		It("should dump OpenFlow flows for br-int", func() {
			By("Running ovs-ofctl-dump-flows for br-int")
			output, err := mcpInspector.
				MethodCall(ovsOfctlDumpFlowsToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"bridge":    "br-int",
					"head": 5,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains OpenFlow rules")
			flowResult := utils.UnmarshalCallToolResult[types.FlowsResult](output)
			Expect(flowResult.Bridge).To(Equal("br-int"))
			Expect(flowResult.Flows).NotTo(BeEmpty())

			By("Verifying OpenFlow flows output")
			GinkgoWriter.Print("\n=== OpenFlow Flows (br-int) ===\n")
			GinkgoWriter.Print("Node: ", podName, "\n")
			GinkgoWriter.Print("Bridge: ", flowResult.Bridge, "\n")
			GinkgoWriter.Print("Number of flows: ", len(flowResult.Flows), "\n")
			for i, flow := range flowResult.Flows {
				if i >= 5 {
					GinkgoWriter.Print("... (showing first 5 flows only)\n")
					break
				}
				GinkgoWriter.Print("Flow ", i+1, ": ", flow, "\n")
			}
			GinkgoWriter.Print("===============================\n\n")

			// OpenFlow flows contain table, priority, and actions
			// Check that at least one flow contains expected markers
			foundValidFlow := false
			for _, flow := range flowResult.Flows {
				if strings.Contains(flow, "table=") || strings.Contains(flow, "priority=") || strings.Contains(flow, "actions=") {
					foundValidFlow = true
					break
				}
			}
			Expect(foundValidFlow).To(BeTrue(), "Expected at least one flow to contain table=, priority=, or actions= markers")
		})
	})

	Context("OVS Appctl Dump Conntrack", func() {
		It("should dump connection tracking entries", func() {
			By("Running ovs-appctl-dump-conntrack")
			output, err := mcpInspector.
				MethodCall(ovsAppctlDumpConntrackToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"head": 30,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains conntrack information")
			conntrackResult := utils.UnmarshalCallToolResult[types.ConntrackResult](output)
			// Conntrack may be empty if no active connections
			Expect(conntrackResult.Entries).NotTo(BeNil())

			By("Verifying conntrack dump output")
			GinkgoWriter.Print("\n=== Connection Tracking Entries ===\n")
			GinkgoWriter.Print("Node: ", podName, "\n")
			if len(conntrackResult.Entries) == 0 {
				GinkgoWriter.Print("No active connections\n")
			} else {
				GinkgoWriter.Print("Connections:\n")
				for _, entry := range conntrackResult.Entries {
					GinkgoWriter.Print("  ", entry, "\n")
				}
			}
			GinkgoWriter.Print("===================================\n\n")

			// Conntrack output may be empty or contain protocol info
			// We just verify the call succeeded
		})
	})

	Context("OVS Appctl Ofproto Trace", func() {
		It("should trace packet through OpenFlow pipeline", func() {
			By("Running ovs-appctl-ofproto-trace")
			output, err := mcpInspector.
				MethodCall(ovsAppctlOfprotoTraceToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"bridge":    "br-int",
					"flow":      "in_port=1,dl_src=00:00:00:00:00:01,dl_dst=00:00:00:00:00:02",
					"head": 50,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains trace information")
			traceResult := utils.UnmarshalCallToolResult[types.OfprotoTraceResult](output)
			Expect(traceResult.Bridge).To(Equal("br-int"))
			Expect(traceResult.Output).NotTo(BeEmpty())

			By("Verifying ofproto trace output")
			GinkgoWriter.Print("\n=== OpenFlow Pipeline Trace ===\n")
			GinkgoWriter.Print("Node: ", podName, "\n")
			GinkgoWriter.Print("Bridge: ", traceResult.Bridge, "\n")
			GinkgoWriter.Print("Flow: ", traceResult.Flow, "\n")
			GinkgoWriter.Print("Trace Output:\n", traceResult.Output, "\n")
			GinkgoWriter.Print("===============================\n\n")

			// Trace output contains pipeline processing steps
			Expect(traceResult.Output).To(Or(
				ContainSubstring("Flow:"),
				ContainSubstring("Rule:"),
				ContainSubstring("OpenFlow"),
			))
		})
	})
})