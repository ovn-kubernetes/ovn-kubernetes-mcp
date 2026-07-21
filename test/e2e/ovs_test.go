package e2e

import (
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/ovs/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

var _ = Describe("OVS Tools", func() {
	const (
		ovsVsctlToolName  = "ovs-vsctl"
		ovsOfctlToolName  = "ovs-ofctl"
		ovsAppctlToolName = "ovs-appctl"
	)

	var namespace, podName string

	BeforeEach(func() {
		By("Finding an ovnkube-node pod")
		var err error
		namespace, podName, err = utils.FindOVNKubeNodePod(kubeClient)
		Expect(err).NotTo(HaveOccurred(), "Failed to find ovnkube-node pod")
	})

	Context("OVS List Bridges", func() {
		It("should list OVS bridges", func() {
			By("Running ovs-vsctl list-br")
			output, err := mcpInspector.
				MethodCall(ovsVsctlToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"action":    "list-br",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains bridge information")
			result := utils.UnmarshalCallToolResult[types.VsctlResult](output)
			Expect(result.Bridges).NotTo(BeEmpty())

			By("Verifying OVS bridges output")
			bridgeNames := strings.Join(result.Bridges, " ")
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
			By("Running ovs-vsctl list-ports for br-int")
			output, err := mcpInspector.
				MethodCall(ovsVsctlToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"action":    "list-ports",
					"bridge":    "br-int",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains port information")
			result := utils.UnmarshalCallToolResult[types.VsctlResult](output)
			Expect(result.Ports).NotTo(BeEmpty())

			By("Verifying OVS ports output")
			portNames := strings.Join(result.Ports, " ")
			Expect(portNames).To(Or(
				ContainSubstring("patch"),
				ContainSubstring("ovn-k8s-mp0"),
				MatchRegexp(`[a-f0-9]{15}`),
			))
		})
	})

	Context("OVS List Interfaces", func() {
		It("should list OVS interfaces for br-int bridge", func() {
			By("Running ovs-vsctl list-ifaces for br-int")
			output, err := mcpInspector.
				MethodCall(ovsVsctlToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"action":    "list-ifaces",
					"bridge":    "br-int",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains interface information")
			result := utils.UnmarshalCallToolResult[types.VsctlResult](output)
			Expect(result.Interfaces).NotTo(BeNil())
			Expect(result.Interfaces).NotTo(BeEmpty())

			By("Verifying OVS interfaces output")
			interfaceNames := strings.Join(result.Interfaces, " ")
			Expect(interfaceNames).To(Or(
				ContainSubstring("br-int"),
				ContainSubstring("patch"),
				ContainSubstring("genev"),
			))
		})
	})

	Context("OVS Vsctl Show", func() {
		It("should show OVS database configuration", func() {
			By("Running ovs-vsctl show")
			output, err := mcpInspector.
				MethodCall(ovsVsctlToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"action":    "show",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains OVS configuration")
			result := utils.UnmarshalCallToolResult[types.VsctlResult](output)
			Expect(result.Output).NotTo(BeEmpty())

			By("Verifying OVS vsctl show output")
			Expect(result.Output).To(And(
				ContainSubstring("Bridge"),
				ContainSubstring("Port"),
				ContainSubstring("Interface"),
			))
		})
	})

	Context("OVS Ofctl Dump Flows", func() {
		It("should dump OpenFlow flows for br-int", func() {
			By("Running ovs-ofctl dump-flows for br-int")
			var flowResult types.OfctlResult
			Eventually(func() bool {
				output, err := mcpInspector.
					MethodCall(ovsOfctlToolName, map[string]any{
						"namespace": namespace,
						"name":      podName,
						"action":    "dump-flows",
						"bridge":    "br-int",
						"head":      5,
					}).Execute()
				if err != nil || len(output) == 0 {
					return false
				}
				var callResult mcp.CallToolResult
				if callResult.UnmarshalJSON(output) != nil || callResult.IsError {
					return false
				}
				flowResult = utils.UnmarshalCallToolResult[types.OfctlResult](output)
				return len(flowResult.Flows) > 0
			}, 30*time.Second, 5*time.Second).Should(BeTrue())

			By("Checking the result contains OpenFlow rules")
			Expect(flowResult.Bridge).To(Equal("br-int"))
			Expect(flowResult.Flows).NotTo(BeEmpty())

			By("Verifying OpenFlow flows output")
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
			By("Running ovs-appctl dpctl/dump-conntrack")
			output, err := mcpInspector.
				MethodCall(ovsAppctlToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"action":    "dpctl/dump-conntrack",
					"head":      30,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains conntrack information")
			result := utils.UnmarshalCallToolResult[types.AppctlResult](output)
			Expect(result.Entries).NotTo(BeNil())

		})
	})

	Context("OVS Appctl Ofproto Trace", func() {
		It("should trace packet through OpenFlow pipeline", func() {
			By("Running ovs-appctl ofproto/trace")
			output, err := mcpInspector.
				MethodCall(ovsAppctlToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"action":    "ofproto/trace",
					"bridge":    "br-int",
					"flow":      "in_port=1,dl_src=00:00:00:00:00:01,dl_dst=00:00:00:00:00:02",
					"head":      50,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains trace information")
			result := utils.UnmarshalCallToolResult[types.AppctlResult](output)
			Expect(result.Bridge).To(Equal("br-int"))
			Expect(result.Output).NotTo(BeEmpty())

			By("Verifying ofproto trace output")
			Expect(result.Output).To(Or(
				ContainSubstring("Flow:"),
				ContainSubstring("Rule:"),
				ContainSubstring("OpenFlow"),
			))
		})
	})

	withPod := func(args map[string]any) map[string]any {
		toolArgs := make(map[string]any, len(args)+2)
		for key, value := range args {
			toolArgs[key] = value
		}
		toolArgs["namespace"] = namespace
		toolArgs["name"] = podName
		return toolArgs
	}

	expectToolError := func(toolName string, toolArgs map[string]any, wantError string) {
		output, err := mcpInspector.
			MethodCall(toolName, toolArgs).
			Execute()
		Expect(err).NotTo(HaveOccurred())

		var result mcp.CallToolResult
		Expect(result.UnmarshalJSON(output)).To(Succeed())
		Expect(result.IsError).To(BeTrue())
		Expect(result.Content).NotTo(BeEmpty())

		var messages []string
		for _, content := range result.Content {
			textContent, ok := content.(*mcp.TextContent)
			if ok {
				messages = append(messages, textContent.Text)
			}
		}

		Expect(messages).NotTo(BeEmpty(), "expected error text content in CallToolResult")
		Expect(strings.Join(messages, "\n")).To(ContainSubstring(wantError))
	}

	Context("validation errors", func() {
		DescribeTable("should reject invalid OVS tool inputs",
			func(toolName string, toolArgs map[string]any, wantError string) {
				expectToolError(toolName, withPod(toolArgs), wantError)
			},
			Entry("ovs-vsctl list-ports invalid bridge", ovsVsctlToolName, map[string]any{
				"action": "list-ports",
				"bridge": "br-int; drop",
			}, "invalid bridge name"),
			Entry("ovs-vsctl list-ifaces invalid bridge", ovsVsctlToolName, map[string]any{
				"action": "list-ifaces",
				"bridge": "br-int; drop",
			}, "invalid bridge name"),
			Entry("ovs-ofctl dump-flows invalid bridge", ovsOfctlToolName, map[string]any{
				"action": "dump-flows",
				"bridge": "br-int; drop",
			}, "invalid bridge name"),
			Entry("ovs-appctl ofproto/trace invalid bridge", ovsAppctlToolName, map[string]any{
				"action": "ofproto/trace",
				"bridge": "br-int; drop",
				"flow":   "in_port=1",
			}, "invalid bridge name"),
		)

		DescribeTable("should reject metacharacters in OVS tool parameters",
			func(toolName string, toolArgs map[string]any) {
				expectToolError(toolName, withPod(toolArgs), "invalid use of metacharacters in parameter")
			},
			Entry("ovs-appctl ofproto/trace flow with metacharacters", ovsAppctlToolName, map[string]any{
				"action": "ofproto/trace",
				"bridge": "br-int",
				"flow":   "in_port=1,dl_src=00:00:00:00:00:01; drop",
			}),
		)
	})
})
