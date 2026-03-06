package e2e

import (
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mgtypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/must-gather/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

// getTestdataMustGatherPath computes the path to the testdata must-gather directory
func getTestdataMustGatherPath() string {
	return utils.GetTestdataPath("../../pkg/must-gather/testdata/must-gather")
}

var _ = Describe("[offline] Must-Gather Tools", func() {
	testdataMustGatherPath := getTestdataMustGatherPath()
	const (
		getResourceToolName         = "must-gather-get-resource"
		listResourcesToolName       = "must-gather-list-resources"
		podLogsToolName             = "must-gather-pod-logs"
		ovnkInfoToolName            = "must-gather-ovnk-info"
		listNorthboundDatabasesName = "must-gather-list-northbound-databases"
		listSouthboundDatabasesName = "must-gather-list-southbound-databases"
		queryDatabaseToolName       = "must-gather-query-database"
	)

	DescribeTable("Get Resource",
		func(mustGatherPath string, kind string, namespace string, name string, outputType string, expectedSubstrings []string, shouldFail bool) {
			params := map[string]string{
				"must_gather_path": mustGatherPath,
				"kind":             kind,
				"name":             name,
			}
			if namespace != "" {
				params["namespace"] = namespace
			}
			if outputType != "" {
				params["outputType"] = outputType
			}

			By("Calling must-gather-get-resource")
			output, err := mcpInspector.
				MethodCall(getResourceToolName, params).Execute()

			Expect(err).NotTo(HaveOccurred())
			if shouldFail {
				Expect(string(output)).To(ContainSubstring(`"isError": true`))
			} else {
				Expect(output).NotTo(BeEmpty())

				By("Checking the result")
				result := utils.UnmarshalCallToolResult[mgtypes.ResourceResult](output)
				Expect(result.Data).NotTo(BeEmpty())

				// Verify data contains expected content
				for _, expected := range expectedSubstrings {
					Expect(result.Data).To(ContainSubstring(expected))
				}
			}
		},
		Entry("should get a pod in default namespace",
			testdataMustGatherPath, "Pod", "default", "test-pod", "",
			[]string{"test-pod", "Running"},
			false,
		),
		Entry("should fail with non-existent must-gather path",
			"/path/that/does/not/exist", "Pod", "default", "test-pod", "",
			nil,
			true,
		),
		Entry("should fail with non-existent resource",
			testdataMustGatherPath, "Pod", "default", "non-existent-pod", "",
			nil,
			true,
		),
	)

	DescribeTable("List Resources",
		func(mustGatherPath string, kind string, namespace string, labelSelector string, outputType string, shouldFail bool) {
			params := map[string]string{
				"must_gather_path": mustGatherPath,
				"kind":             kind,
			}
			if namespace != "" {
				params["namespace"] = namespace
			}
			if labelSelector != "" {
				params["labelSelector"] = labelSelector
			}
			if outputType != "" {
				params["outputType"] = outputType
			}

			By("Calling must-gather-list-resources")
			output, err := mcpInspector.
				MethodCall(listResourcesToolName, params).Execute()

			Expect(err).NotTo(HaveOccurred())
			if shouldFail {
				Expect(string(output)).To(ContainSubstring(`"isError": true`))
			} else {
				Expect(output).NotTo(BeEmpty())

				By("Checking the result")
				result := utils.UnmarshalCallToolResult[mgtypes.ResourceResult](output)
				Expect(result.Data).NotTo(BeEmpty())
			}
		},
		Entry("should list all pods in a namespace",
			testdataMustGatherPath, "Pod", "default", "", "",
			false,
		),
		Entry("should fail with invalid resource kind",
			testdataMustGatherPath, "InvalidKind", "default", "", "",
			true,
		),
	)

	DescribeTable("Pod Logs",
		func(mustGatherPath string, namespace string, podName string, container string, previous bool, pattern string, head int, tail int, expectedSubstrings []string, shouldFail bool) {
			params := map[string]string{
				"must_gather_path": mustGatherPath,
				"name":             podName,
			}
			if namespace != "" {
				params["namespace"] = namespace
			}
			if container != "" {
				params["container"] = container
			}
			if previous {
				params["previous"] = "true"
			}
			if pattern != "" {
				params["pattern"] = pattern
			}
			if head > 0 {
				params["head"] = strconv.Itoa(head)
			}
			if tail > 0 {
				params["tail"] = strconv.Itoa(tail)
			}

			By("Calling must-gather-pod-logs")
			output, err := mcpInspector.
				MethodCall(podLogsToolName, params).Execute()

			Expect(err).NotTo(HaveOccurred())
			if shouldFail {
				Expect(string(output)).To(ContainSubstring(`"isError": true`))
			} else {
				Expect(output).NotTo(BeEmpty())

				By("Checking the result")
				result := utils.UnmarshalCallToolResult[mgtypes.GetPodLogsResult](output)

				// Verify logs contain expected content
				if len(expectedSubstrings) > 0 {
					var logContent strings.Builder
					for _, line := range result.Logs {
						logContent.WriteString(line)
						logContent.WriteString("\n")
					}
					for _, expected := range expectedSubstrings {
						Expect(logContent.String()).To(ContainSubstring(expected))
					}
				}
			}
		},
		Entry("should get pod logs",
			testdataMustGatherPath, "default", "test-pod", "", false, "", 0, 0,
			[]string{"Starting application"},
			false,
		),
		Entry("should fail with non-existent pod",
			testdataMustGatherPath, "default", "non-existent-pod", "", false, "", 0, 0,
			nil,
			true,
		),
	)

	DescribeTable("OVN-K Info",
		func(mustGatherPath string, infoType mgtypes.InfoType, expectedSubstrings []string, shouldFail bool) {
			params := map[string]string{
				"must_gather_path": mustGatherPath,
				"info_type":        string(infoType),
			}

			By("Calling must-gather-ovnk-info")
			output, err := mcpInspector.
				MethodCall(ovnkInfoToolName, params).Execute()

			Expect(err).NotTo(HaveOccurred())
			if shouldFail {
				Expect(string(output)).To(ContainSubstring(`"isError": true`))
			} else {
				Expect(output).NotTo(BeEmpty())

				By("Checking the result")
				result := utils.UnmarshalCallToolResult[mgtypes.GetOvnKInfoResult](output)
				Expect(result.Data).NotTo(BeEmpty())

				// Verify data contains expected content
				for _, expected := range expectedSubstrings {
					Expect(result.Data).To(ContainSubstring(expected))
				}
			}
		},
		Entry("should get subnet info",
			testdataMustGatherPath, mgtypes.InfoTypeSubnets,
			[]string{"HOST/NODE"},
			false,
		),
		Entry("should fail with invalid info type",
			testdataMustGatherPath, mgtypes.InfoType("invalid"),
			nil,
			true,
		),
	)

	// OVSDB tool tests - these are skipped if ovsdb-tool is not available in PATH
	// To run these tests, install ovsdb-tool: dnf install openvswitch (RHEL/Fedora) or apt-get install openvswitch-common (Ubuntu)
	Context("OVSDB Tools", Label("requires-ovsdb-tool"), func() {
		DescribeTable("List OVN Databases",
			func(toolName string, mustGatherPath string, shouldFail bool, expectedDatabases []string) {
				params := map[string]string{
					"must_gather_path": mustGatherPath,
				}

				By("Calling " + toolName)
				output, err := mcpInspector.
					MethodCall(toolName, params).Execute()
				Expect(err).NotTo(HaveOccurred())
				if shouldFail {
					Expect(string(output)).To(ContainSubstring(`"isError": true`))
				} else {
					Expect(output).NotTo(BeEmpty())

					// Basic validation - if tool is available, check for expected databases
					outputStr := string(output)
					for _, dbName := range expectedDatabases {
						Expect(outputStr).To(ContainSubstring(dbName))
					}
				}
			},
			Entry("should list northbound databases",
				listNorthboundDatabasesName, testdataMustGatherPath, false,
				[]string{"_nbdb"},
			),
			Entry("should list southbound databases",
				listSouthboundDatabasesName, testdataMustGatherPath, false,
				[]string{"_sbdb"},
			),
		)

		DescribeTable("Query Database",
			func(mustGatherPath string, databaseName string, table string, expectedSubstrings []string, shouldFail bool) {
				params := map[string]string{
					"must_gather_path": mustGatherPath,
					"database_name":    databaseName,
					"table":            table,
				}

				By("Calling must-gather-query-database")
				output, err := mcpInspector.
					MethodCall(queryDatabaseToolName, params).Execute()
				Expect(err).NotTo(HaveOccurred())
				if shouldFail {
					// For database query errors, ovsdb-tool returns error details in the JSON output
					// rather than failing with a non-zero exit code, so we check for error indicators
					outputStr := string(output)
					Expect(outputStr).To(Or(
						ContainSubstring(`"isError": true`),
						ContainSubstring(`"error"`),
						ContainSubstring("syntax error"),
					))
				} else {
					Expect(output).NotTo(BeEmpty())

					// Verify output contains expected data
					outputStr := string(output)
					for _, expected := range expectedSubstrings {
						Expect(outputStr).To(ContainSubstring(expected))
					}
				}
			},
			Entry("should query Chassis table from southbound database",
				testdataMustGatherPath, "ovnkube-node-test123_sbdb", "Chassis",
				[]string{"test-chassis", "hostname"},
				false,
			),
			Entry("should fail with non-existent must-gather path",
				"/path/that/does/not/exist", "ovnkube-node-test123_sbdb", "Chassis",
				nil,
				true,
			),
			Entry("should fail with invalid table name",
				testdataMustGatherPath, "ovnkube-node-test123_sbdb", "InvalidTable",
				nil,
				true,
			),
		)
	})
})
