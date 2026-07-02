package e2e

import (
	"context"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/ovn/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

var _ = Describe("OVN Tools", func() {
	const (
		ovnShowToolName      = "ovn-show"
		ovnGetToolName       = "ovn-get"
		ovnLflowListToolName = "ovn-lflow-list"
		ovnTraceToolName     = "ovn-trace"
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

	Context("OVN Show", func() {
		It("should show OVN Northbound database configuration", func() {
			By("Running ovn-show for NBDB")
			output, err := mcpInspector.
				MethodCall(ovnShowToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"database":  "nbdb",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			showResult := utils.UnmarshalCallToolResult[types.ShowResult](output)
			Expect(showResult.Database).To(Equal(types.NorthboundDB))
			Expect(showResult.Output).NotTo(BeEmpty())
			
			// Log database verification step with actual output
			By("Verifying OVN Northbound Database output contains expected content")
			GinkgoWriter.Print("\n=== OVN Northbound Database Configuration ===\n")
			GinkgoWriter.Print("Database: ", showResult.Database, "\n")
			GinkgoWriter.Print("Output:\n", showResult.Output, "\n")
			GinkgoWriter.Print("==============================================\n\n")
			
			// NBDB show output typically contains "switch" for logical switches
			Expect(showResult.Output).To(Or(
				ContainSubstring("switch"),
				ContainSubstring("router"),
			))
		})

		It("should show OVN Southbound database configuration", func() {
			By("Running ovn-show for SBDB")
			output, err := mcpInspector.
				MethodCall(ovnShowToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"database":  "sbdb",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			showResult := utils.UnmarshalCallToolResult[types.ShowResult](output)
			Expect(showResult.Database).To(Equal(types.SouthboundDB))
			Expect(showResult.Output).NotTo(BeEmpty())
			
			// Log database verification step with actual output
			By("Verifying OVN Southbound Database output contains expected content")
			GinkgoWriter.Print("\n=== OVN Southbound Database Configuration ===\n")
			GinkgoWriter.Print("Database: ", showResult.Database, "\n")
			GinkgoWriter.Print("Output:\n", showResult.Output, "\n")
			GinkgoWriter.Print("==============================================\n\n")
			
			// SBDB show output typically contains "Chassis" information
			Expect(showResult.Output).To(ContainSubstring("Chassis"))
		})
	})

	Context("OVN Get", func() {
		It("should get Logical_Switch records from NBDB", func() {
			By("Running ovn-get for Logical_Switch table")
			output, err := mcpInspector.
				MethodCall(ovnGetToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"database":  "nbdb",
					"table":     "Logical_Switch",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains logical switch records")
			getResult := utils.UnmarshalCallToolResult[types.GetResult](output)
			Expect(getResult.Database).To(Equal(types.NorthboundDB))
			Expect(getResult.Table).To(Equal("Logical_Switch"))
			Expect(getResult.Output).NotTo(BeEmpty())

			By("Verifying logical switch records output")
			GinkgoWriter.Print("\n=== OVN Logical Switch Records ===\n")
			GinkgoWriter.Print("Database: ", getResult.Database, "\n")
			GinkgoWriter.Print("Table: ", getResult.Table, "\n")
			GinkgoWriter.Print("Output:\n", getResult.Output, "\n")
			GinkgoWriter.Print("===================================\n\n")

			// Logical switch output typically contains UUIDs and external_ids
			Expect(getResult.Output).To(Or(
				ContainSubstring("external_ids"),
				ContainSubstring("name"),
				ContainSubstring("_uuid"),
			))
		})
	})

	Context("OVN Lflow List", func() {
		It("should list logical flows from SBDB", func() {
			By("Running ovn-lflow-list")
			output, err := mcpInspector.
				MethodCall(ovnLflowListToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"head":      50,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains logical flows")
			lflowResult := utils.UnmarshalCallToolResult[types.LogicalFlowListResult](output)
			Expect(lflowResult.Flows).NotTo(BeEmpty())

			By("Verifying logical flows output")
			GinkgoWriter.Print("\n=== OVN Logical Flows ===\n")
			GinkgoWriter.Print("Number of flows: ", len(lflowResult.Flows), "\n")
			for i, flow := range lflowResult.Flows {
				if i >= 5 {
					GinkgoWriter.Print("... (showing first 5 flows only)\n")
					break
				}
				GinkgoWriter.Print("Flow ", i+1, ": ", flow, "\n")
			}
			GinkgoWriter.Print("=========================\n\n")

			// Logical flows contain table numbers, priorities, and actions
			// Check the second flow (first is often a header)
			flowToCheck := lflowResult.Flows[0]
			if len(lflowResult.Flows) > 1 && !strings.Contains(lflowResult.Flows[0], "table=") {
				flowToCheck = lflowResult.Flows[1]
			}
			Expect(flowToCheck).To(Or(
				ContainSubstring("table="),
				ContainSubstring("priority="),
				ContainSubstring("action="),
			))
		})
	})

	Context("OVN Trace", func() {
		It("should trace a packet through the logical network", func() {
			By("Running ovn-trace with a simple microflow")
			output, err := mcpInspector.
				MethodCall(ovnTraceToolName, map[string]any{
					"namespace":  namespace,
					"name":       podName,
					"datapath":   "join", // Use the join switch which should exist
					"microflow":  "eth.src==00:00:00:00:00:01 && eth.dst==00:00:00:00:00:02",
					"mode":       "summary",
					"head":       30,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains trace information")
			traceResult := utils.UnmarshalCallToolResult[types.OVNTraceResult](output)
			Expect(traceResult.Datapath).To(Equal("join"))
			Expect(traceResult.Output).NotTo(BeEmpty())

			By("Verifying trace output")
			GinkgoWriter.Print("\n=== OVN Trace Results ===\n")
			GinkgoWriter.Print("Datapath: ", traceResult.Datapath, "\n")
			GinkgoWriter.Print("Microflow: ", traceResult.Microflow, "\n")
			GinkgoWriter.Print("Trace Output:\n", traceResult.Output, "\n")
			GinkgoWriter.Print("=========================\n\n")

			// Trace output contains packet processing steps
			Expect(traceResult.Output).To(Or(
				ContainSubstring("ingress"),
				ContainSubstring("egress"),
				ContainSubstring("output"),
			))
		})
	})
})
