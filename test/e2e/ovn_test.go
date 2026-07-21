package e2e

import (
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

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

	var namespace, podName string

	BeforeEach(func() {
		By("Finding an ovnkube-node pod")
		var err error
		namespace, podName, err = utils.FindOVNKubeNodePod(kubeClient)
		Expect(err).NotTo(HaveOccurred(), "Failed to find ovnkube-node pod")
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
			
			By("Verifying NBDB output contains expected content")
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
			
			By("Verifying SBDB output contains expected content")
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
			Expect(traceResult.Output).To(Or(
				ContainSubstring("ingress"),
				ContainSubstring("egress"),
				ContainSubstring("output"),
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
		DescribeTable("should reject invalid OVN tool inputs",
			func(toolName string, toolArgs map[string]any, wantError string) {
				expectToolError(toolName, withPod(toolArgs), wantError)
			},
			Entry("ovn-show invalid database", ovnShowToolName, map[string]any{
				"database": "invalid_db",
			}, "invalid database"),
			Entry("ovn-get invalid database", ovnGetToolName, map[string]any{
				"database": "invalid_db",
				"table":    "Logical_Switch",
			}, "invalid database"),
			Entry("ovn-get invalid table name", ovnGetToolName, map[string]any{
				"database": "nbdb",
				"table":    "123_Invalid",
			}, "invalid table name"),
		)

		DescribeTable("should reject metacharacters in OVN tool parameters",
			func(toolName string, toolArgs map[string]any) {
				expectToolError(toolName, withPod(toolArgs), "invalid use of metacharacters in parameter")
			},
			Entry("ovn-get columns with metacharacters", ovnGetToolName, map[string]any{
				"database": "nbdb",
				"table":    "Logical_Switch",
				"columns":  "name; drop",
			}),
			Entry("ovn-lflow-list datapath with metacharacters", ovnLflowListToolName, map[string]any{
				"datapath": "join | true",
			}),
			Entry("ovn-trace datapath with metacharacters", ovnTraceToolName, map[string]any{
				"datapath":  "join; true",
				"microflow": "eth.src==00:00:00:00:00:01",
			}),
			Entry("ovn-trace microflow with metacharacters", ovnTraceToolName, map[string]any{
				"datapath":  "join",
				"microflow": "eth.src==00:00:00:00:00:01 | drop",
			}),
		)
	})
})
