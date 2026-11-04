// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"maps"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/uuid"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/ovn/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

var _ = Describe("OVN Tools", Ordered, func() {
	const (
		ovnShowToolName      = "ovn-show"
		ovnGetToolName       = "ovn-get"
		ovnLflowListToolName = "ovn-lflow-list"
		ovnTraceToolName     = "ovn-trace"
	)

	var namespace, podName string

	BeforeAll(func() {
		By("Finding an ovnkube-node pod")
		var err error
		namespace, podName, err = utils.FindOVNKubeNodePod(kubeClient)
		Expect(err).NotTo(HaveOccurred(), "Failed to find ovnkube-node pod")
	})

	Context("command execution errors", func() {
		var errorPod string
		BeforeAll(func(ctx SpecContext) {
			source := &corev1.Pod{}
			Expect(kubeClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: podName}, source)).To(Succeed())
			errorPod = utils.CreateToolPod(ctx, kubeClient, namespace, source.Spec.Containers[0].Image, []corev1.EnvVar{
				{Name: "OVN_NB_DAEMON", Value: "/tmp/absent-nbctl.ctl"},
				{Name: "OVN_SB_DAEMON", Value: "/tmp/absent-sbctl.ctl"},
			})
		})
		DescribeTable("should report the native command error",
			func(tool string, args map[string]any, wantError string) {
				args["namespace"] = namespace
				args["name"] = errorPod
				utils.ExpectToolError(mcpInspector, tool, args, wantError)
			},
			Entry("ovn-show", ovnShowToolName, map[string]any{"database": "nbdb"}, "could not connect to ovn-nbctl daemon"),
			Entry("ovn-get", ovnGetToolName, map[string]any{"database": "nbdb", "table": "Logical_Switch"}, "could not connect to ovn-nbctl daemon"),
			Entry("ovn-lflow-list", ovnLflowListToolName, map[string]any{}, "could not connect to ovn-sbctl daemon"),
			Entry("ovn-trace", ovnTraceToolName, map[string]any{"datapath": "zz-absent", "microflow": "ip4"}, "database connection failed"),
		)
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
		DescribeTable("should list logical flows from SBDB", func(datapath string) {
			By("Running ovn-lflow-list")
			output, err := mcpInspector.
				MethodCall(ovnLflowListToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"datapath":  datapath,
					"head":      50,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains logical flows within the head limit")
			lflowResult := utils.UnmarshalCallToolResult[types.LogicalFlowListResult](output)
			Expect(lflowResult.Datapath).To(Equal(datapath))
			Expect(lflowResult.Flows).NotTo(BeEmpty())
			Expect(len(lflowResult.Flows)).To(BeNumerically("<=", 50))

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
		},
			Entry("all datapaths", ""),
			Entry("named datapath", "join"),
		)
	})

	Context("OVN Trace", func() {
		It("should trace a packet through the logical network", func() {
			By("Running ovn-trace with a simple microflow")
			output, err := mcpInspector.
				MethodCall(ovnTraceToolName, map[string]any{
					"namespace": namespace,
					"name":      podName,
					"datapath":  "join", // Use the join switch which should exist
					"microflow": "eth.src==00:00:00:00:00:01 && eth.dst==00:00:00:00:00:02",
					"mode":      "summary",
					"head":      30,
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
		maps.Copy(toolArgs, args)
		toolArgs["namespace"] = namespace
		toolArgs["name"] = podName
		return toolArgs
	}

	Context("validation errors", func() {
		DescribeTable("should reject option-like OVN operands",
			func(toolName string, toolArgs map[string]any, wantError string) {
				args := withPod(toolArgs)
				// A missed validation must not execute injected options on an OVN pod.
				args["namespace"] = "e2e-ovn-validation-" + uuid.NewString()
				utils.ExpectToolError(mcpInspector, toolName, args, wantError)
			},
			Entry("ovn-trace datapath option", ovnTraceToolName, map[string]any{
				"datapath":  "--db=tcp:192.0.2.1:6642",
				"microflow": "eth.src==00:00:00:00:00:01",
			}, "invalid datapath name: must not start with '-'"),
			Entry("ovn-trace microflow option", ovnTraceToolName, map[string]any{
				"datapath":  "join",
				"microflow": "--help",
			}, "invalid microflow specification: must not start with '-'"),
			Entry("ovn-lflow-list datapath option", ovnLflowListToolName, map[string]any{
				"datapath": "--db=tcp:192.0.2.1:6642",
			}, "invalid datapath name: must not start with '-'"),
			Entry("ovn-trace datapath option after whitespace", ovnTraceToolName, map[string]any{
				"datapath":  " \t--db=tcp:192.0.2.1:6642",
				"microflow": "eth.src==00:00:00:00:00:01",
			}, "invalid datapath name: must not start with '-'"),
			Entry("ovn-trace microflow option after whitespace", ovnTraceToolName, map[string]any{
				"datapath":  "join",
				"microflow": " \t--help",
			}, "invalid microflow specification: must not start with '-'"),
		)

		DescribeTable("should reject invalid OVN tool inputs",
			func(toolName string, toolArgs map[string]any, wantError string) {
				utils.ExpectToolError(mcpInspector, toolName, withPod(toolArgs), wantError)
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
				utils.ExpectToolError(mcpInspector, toolName, withPod(toolArgs), "invalid use of metacharacters in parameter")
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
