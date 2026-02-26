package e2e

import (
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	mgtypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/must-gather/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

// getTestdataMustGatherPath computes the path to the testdata must-gather directory
func getTestdataMustGatherPath() string {
	return utils.GetTestdataPath("../../pkg/must-gather/testdata/must-gather")
}

// Must-Gather Tools tests - these tests will fail if the omc command is not available in PATH
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

	type resourceInfo struct {
		Kind      string            `json:"kind"`
		Name      string            `json:"name"`
		Namespace string            `json:"namespace"`
		Labels    map[string]string `json:"labels"`
	}
	marshalResourceInfo := func(info resourceInfo) string {
		jsonData, err := json.Marshal(info)
		Expect(err).NotTo(HaveOccurred())
		return string(jsonData)
	}
	unmarshalResourceInfo := func(jsonData string) resourceInfo {
		var info resourceInfo
		err := json.Unmarshal([]byte(jsonData), &info)
		Expect(err).NotTo(HaveOccurred())
		return info
	}

	DescribeTable("Get Resource",
		func(mustGatherPath string, kind string, namespace string, name string, outputType string, expectedSubstrings []string, shouldFail bool) {
			By("Calling must-gather-get-resource")
			output, err := mcpInspector.
				MethodCall(getResourceToolName, map[string]any{
					"must_gather_path": mustGatherPath,
					"kind":             kind,
					"name":             name,
					"namespace":        namespace,
					"output_type":      outputType,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())

			if shouldFail {
				Expect(string(output)).To(ContainSubstring(`"isError": true`))
			} else {
				Expect(output).NotTo(BeEmpty())

				By("Checking the result")
				result := utils.UnmarshalCallToolResult[mgtypes.ResourceResult](output)
				Expect(result.Data).NotTo(BeEmpty())

				if outputType == "json" || outputType == "yaml" {
					if outputType == "yaml" {
						jsonData, err := yaml.YAMLToJSON([]byte(result.Data))
						Expect(err).NotTo(HaveOccurred())
						result.Data = string(jsonData)
					}

					obj := unstructured.Unstructured{}
					err := obj.UnmarshalJSON([]byte(result.Data))
					Expect(err).NotTo(HaveOccurred())

					unmarshalledInfo := unmarshalResourceInfo(expectedSubstrings[0])
					Expect(obj.GetName()).To(Equal(unmarshalledInfo.Name))
					Expect(obj.GetNamespace()).To(Equal(unmarshalledInfo.Namespace))
					Expect(obj.GetKind()).To(Equal(unmarshalledInfo.Kind))
					Expect(obj.GetLabels()).To(Equal(unmarshalledInfo.Labels))
				} else {
					// Verify data contains expected content
					for _, expected := range expectedSubstrings {
						Expect(result.Data).To(ContainSubstring(expected))
					}
				}
			}
		},
		Entry("should get a pod in default namespace",
			testdataMustGatherPath, "Pod", "default", "test-pod", "",
			[]string{"test-pod", "1/1", "Running"},
			false,
		),
		Entry("should get a pod in default namespace with wide as output type",
			testdataMustGatherPath, "Pod", "default", "test-pod", "wide",
			[]string{"test-pod", "1/1", "Running", "10.128.0.10", "worker-0"},
			false,
		),
		Entry("should get a pod in default namespace with yaml as output type",
			testdataMustGatherPath, "Pod", "default", "test-pod", "yaml",
			[]string{marshalResourceInfo(resourceInfo{
				Kind:      "Pod",
				Name:      "test-pod",
				Namespace: "default",
				Labels:    map[string]string{"app": "test", "component": "test-component"},
			})},
			false,
		),
		Entry("should get a pod in default namespace with json as output type",
			testdataMustGatherPath, "Pod", "default", "test-pod-2", "json",
			[]string{marshalResourceInfo(resourceInfo{
				Kind:      "Pod",
				Name:      "test-pod-2",
				Namespace: "default",
				Labels:    map[string]string{"app": "test", "component": "test-component"},
			})},
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
		func(mustGatherPath string, kind string, namespace string, labelSelector string, outputType string, expectedSubstrings []string, shouldFail bool) {
			By("Calling must-gather-list-resources")
			output, err := mcpInspector.
				MethodCall(listResourcesToolName, map[string]any{
					"must_gather_path": mustGatherPath,
					"kind":             kind,
					"namespace":        namespace,
					"label_selector":   labelSelector,
					"output_type":      outputType,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())

			if shouldFail {
				Expect(string(output)).To(ContainSubstring(`"isError": true`))
			} else {
				Expect(output).NotTo(BeEmpty())

				By("Checking the result")
				result := utils.UnmarshalCallToolResult[mgtypes.ResourceResult](output)
				Expect(result.Data).NotTo(BeEmpty())

				switch outputType {
				case "":
					matchExpectedLines(expectedSubstrings, result.Data, true, "")
				case "wide":
					// Split the expected content by the "Running" separator as the actual
					// result will have an age field that is not part of the expected content.
					matchExpectedLines(expectedSubstrings, result.Data, true, "Running")
				case "json", "yaml":
					if outputType == "yaml" {
						jsonData, err := yaml.YAMLToJSON([]byte(result.Data))
						Expect(err).NotTo(HaveOccurred())
						result.Data = string(jsonData)
					}

					listObj := unstructured.UnstructuredList{}
					err := listObj.UnmarshalJSON([]byte(result.Data))
					Expect(err).NotTo(HaveOccurred())

					Expect(len(listObj.Items)).To(Equal(len(expectedSubstrings)))

					namespacedNameObj := map[types.NamespacedName]unstructured.Unstructured{}
					for _, item := range listObj.Items {
						namespacedNameObj[types.NamespacedName{Namespace: item.GetNamespace(), Name: item.GetName()}] = item
					}
					for _, expected := range expectedSubstrings {
						unmarshalledInfo := unmarshalResourceInfo(expected)
						obj, ok := namespacedNameObj[types.NamespacedName{Namespace: unmarshalledInfo.Namespace, Name: unmarshalledInfo.Name}]
						Expect(ok).To(BeTrue())
						Expect(obj.GetKind()).To(Equal(unmarshalledInfo.Kind))
						Expect(obj.GetLabels()).To(Equal(unmarshalledInfo.Labels))
					}
				}
			}
		},
		Entry("should list all pods in all namespaces",
			testdataMustGatherPath, "Pod", "", "", "",
			[]string{strings.Join([]string{"default", "test-pod", "1/1", "Running"}, " "),
				strings.Join([]string{"default", "test-pod-2", "1/1", "Running"}, " "),
				strings.Join([]string{"openshift-ovn-kubernetes", "ovnkube-node-cp0", "2/2", "Running"}, " "),
				strings.Join([]string{"openshift-ovn-kubernetes", "ovnkube-node-cp1", "2/2", "Running"}, " "),
				strings.Join([]string{"openshift-ovn-kubernetes", "ovnkube-node-cp2", "2/2", "Running"}, " "),
				strings.Join([]string{"openshift-ovn-kubernetes", "ovnkube-node-w0", "2/2", "Running"}, " "),
				strings.Join([]string{"openshift-ovn-kubernetes", "ovnkube-node-w1", "2/2", "Running"}, " "),
				strings.Join([]string{"openshift-ovn-kubernetes", "ovnkube-node-w2", "2/2", "Running"}, " ")},
			false,
		),
		Entry("should list all pods in all namespaces with label selector",
			testdataMustGatherPath, "Pod", "", "app=ovnkube-node", "",
			[]string{strings.Join([]string{"openshift-ovn-kubernetes", "ovnkube-node-cp0", "2/2", "Running"}, " "),
				strings.Join([]string{"openshift-ovn-kubernetes", "ovnkube-node-cp1", "2/2", "Running"}, " "),
				strings.Join([]string{"openshift-ovn-kubernetes", "ovnkube-node-cp2", "2/2", "Running"}, " "),
				strings.Join([]string{"openshift-ovn-kubernetes", "ovnkube-node-w0", "2/2", "Running"}, " "),
				strings.Join([]string{"openshift-ovn-kubernetes", "ovnkube-node-w1", "2/2", "Running"}, " "),
				strings.Join([]string{"openshift-ovn-kubernetes", "ovnkube-node-w2", "2/2", "Running"}, " ")},
			false,
		),
		Entry("should list all pods in a namespace",
			testdataMustGatherPath, "Pod", "default", "", "",
			[]string{strings.Join([]string{"test-pod", "1/1", "Running"}, " "),
				strings.Join([]string{"test-pod-2", "1/1", "Running"}, " ")},
			false,
		),
		Entry("should list all pods in a namespace with wide as output type",
			testdataMustGatherPath, "Pod", "default", "", "wide",
			[]string{strings.Join([]string{"test-pod", "1/1", "Running", "10.128.0.10", "worker-0"}, " "),
				strings.Join([]string{"test-pod-2", "1/1", "Running", "10.128.0.11", "worker-1"}, " ")},
			false,
		),
		Entry("should list all pods in a namespace with yaml as output type",
			testdataMustGatherPath, "Pod", "default", "", "yaml",
			[]string{marshalResourceInfo(resourceInfo{
				Kind:      "Pod",
				Name:      "test-pod",
				Namespace: "default",
				Labels:    map[string]string{"app": "test", "component": "test-component"},
			}), marshalResourceInfo(resourceInfo{
				Kind:      "Pod",
				Name:      "test-pod-2",
				Namespace: "default",
				Labels:    map[string]string{"app": "test", "component": "test-component"},
			})},
			false,
		),
		Entry("should list all pods in a namespace with json as output type",
			testdataMustGatherPath, "Pod", "default", "", "json",
			[]string{marshalResourceInfo(resourceInfo{
				Kind:      "Pod",
				Name:      "test-pod",
				Namespace: "default",
				Labels:    map[string]string{"app": "test", "component": "test-component"},
			}), marshalResourceInfo(resourceInfo{
				Kind:      "Pod",
				Name:      "test-pod-2",
				Namespace: "default",
				Labels:    map[string]string{"app": "test", "component": "test-component"},
			})},
			false,
		),
		Entry("should fail with invalid resource kind",
			testdataMustGatherPath, "InvalidKind", "default", "", "",
			nil,
			true,
		),
	)

	DescribeTable("Pod Logs",
		func(mustGatherPath string, namespace string, podName string, container string, previous bool, rotated bool, pattern string, head int, tail int, applyTailFirst bool, expectedSubstrings []string, shouldFail bool) {
			By("Calling must-gather-pod-logs")
			output, err := mcpInspector.
				MethodCall(podLogsToolName, map[string]any{
					"must_gather_path": mustGatherPath,
					"name":             podName,
					"namespace":        namespace,
					"container":        container,
					"previous":         previous,
					"rotated":          rotated,
					"pattern":          pattern,
					"head":             head,
					"tail":             tail,
					"apply_tail_first": applyTailFirst,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())

			if shouldFail {
				Expect(string(output)).To(ContainSubstring(`"isError": true`))
			} else {
				Expect(output).NotTo(BeEmpty())

				By("Checking the result")
				result := utils.UnmarshalCallToolResult[mgtypes.GetPodLogsResult](output)

				// Verify that the number of logs is equal to the number of expected substrings
				Expect(len(result.Logs)).To(Equal(len(expectedSubstrings)))

				// Verify that the logs contain the expected content
				if len(expectedSubstrings) > 0 {
					for index := range expectedSubstrings {
						Expect(result.Logs[index]).To(ContainSubstring(expectedSubstrings[index]))
					}
				}
			}
		},
		Entry("should get pod logs",
			testdataMustGatherPath, "default", "test-pod", "test-container", false, false, "", 0, 0, false,
			[]string{"Starting application",
				"Application initialized successfully",
				"Listening on port 8080",
				"Ready to accept connections",
				"Health check passed"},
			false,
		),
		Entry("should get pod logs without container",
			testdataMustGatherPath, "default", "test-pod", "", false, false, "", 0, 0, false,
			[]string{"Starting application",
				"Application initialized successfully",
				"Listening on port 8080",
				"Ready to accept connections",
				"Health check passed"},
			false,
		),
		Entry("should get pod logs from previous logs",
			testdataMustGatherPath, "default", "test-pod", "test-container", true, false, "", 0, 0, false,
			[]string{"Starting application...",
				"Loaded config from /etc/app/config.yaml",
				"Listening on 0.0.0.0:8080",
				"INFO: Received first request from 10.128.0.5",
				"WARN: Connection timeout to upstream service (retrying)",
				"ERROR: Failed to connect after 3 attempts; backing off",
				"ERROR: Unrecoverable error: connection refused",
				"Fatal error, exiting"},
			false,
		),
		Entry("should get pod logs from previous logs with pattern matching",
			testdataMustGatherPath, "default", "test-pod", "test-container", true, false, "error|Error|ERROR", 0, 0, false,
			[]string{"ERROR: Failed to connect after 3 attempts; backing off",
				"ERROR: Unrecoverable error: connection refused",
				"Fatal error, exiting"},
			false,
		),
		Entry("should get pod logs from rotated logs",
			testdataMustGatherPath, "default", "test-pod", "test-container", false, true, "", 0, 0, false,
			[]string{"Starting application...",
				"Application initialized",
				"Listening on port 8080",
				"Health check passed",
				"Log rotation triggered (size limit)",
				"Starting application...",
				"Loaded config from /etc/app/config.yaml",
				"Listening on 0.0.0.0:8080",
				"INFO: Received first request from 10.128.0.5",
				"WARN: Connection timeout to upstream (retrying)",
				"ERROR: Failed to connect after 3 attempts",
				"Fatal error, exiting"},
			false,
		),
		Entry("should get pod logs with head",
			testdataMustGatherPath, "default", "test-pod", "", false, false, "", 2, 0, false,
			[]string{"Starting application",
				"Application initialized successfully"},
			false,
		),
		Entry("should get pod logs with tail",
			testdataMustGatherPath, "default", "test-pod-2", "", false, false, "", 0, 2, false,
			[]string{"Ready to accept connections",
				"Health check passed"},
			false,
		),
		Entry("should get pod logs with head and tail",
			testdataMustGatherPath, "default", "test-pod-2", "", false, false, "", 3, 2, false,
			[]string{"Application initialized successfully",
				"Listening on port 8080"},
			false,
		),
		Entry("should get pod logs with head and tail and apply tail first",
			testdataMustGatherPath, "default", "test-pod-2", "", false, false, "", 2, 3, true,
			[]string{"Listening on port 8080",
				"Ready to accept connections"},
			false,
		),
		Entry("should fail with non-existent pod",
			testdataMustGatherPath, "default", "non-existent-pod", "", false, false, "", 0, 0, false,
			nil,
			true,
		),
		Entry("should fail with non-existent container",
			testdataMustGatherPath, "default", "test-pod-2", "non-existent-container", false, false, "", 0, 0, false,
			nil,
			true,
		),
	)

	DescribeTable("OVN-K Info",
		func(mustGatherPath string, infoType mgtypes.InfoType, expectedSubstrings []string, shouldFail bool) {
			By("Calling must-gather-ovnk-info")
			output, err := mcpInspector.
				MethodCall(ovnkInfoToolName, map[string]any{
					"must_gather_path": mustGatherPath,
					"info_type":        string(infoType),
				}).Execute()
			Expect(err).NotTo(HaveOccurred())

			if shouldFail {
				Expect(string(output)).To(ContainSubstring(`"isError": true`))
			} else {
				Expect(output).NotTo(BeEmpty())

				By("Checking the result")
				result := utils.UnmarshalCallToolResult[mgtypes.GetOvnKInfoResult](output)
				Expect(result.Data).NotTo(BeEmpty())

				matchExpectedLines(expectedSubstrings, result.Data, false, "")
			}
		},
		Entry("should get subnet info",
			testdataMustGatherPath, mgtypes.InfoTypeSubnets,
			[]string{strings.Join([]string{"HOST/NODE", "ROLE", "NODE SUBNET", "NODE TRANSIT-SWITCH-IP"}, " "),
				strings.Join([]string{"control-plane-0", "control-plane,master", "10.128.0.0/23", "100.88.0.2/16"}, " "),
				strings.Join([]string{"control-plane-1", "control-plane,master", "10.129.0.0/23", "100.88.0.3/16"}, " "),
				strings.Join([]string{"control-plane-2", "control-plane,master", "10.130.0.0/23", "100.88.0.4/16"}, " "),
				strings.Join([]string{"worker-0", "worker", "10.128.2.0/23", "100.88.0.5/16"}, " "),
				strings.Join([]string{"worker-1", "worker", "10.129.2.0/23", "100.88.0.6/16"}, " "),
				strings.Join([]string{"worker-2", "worker", "10.130.2.0/23", "100.88.0.7/16"}, " ")},
			false,
		),
		Entry("should get host net info",
			testdataMustGatherPath, mgtypes.InfoTypeHostNetInfo,
			[]string{strings.Join([]string{"HOST/NODE", "ROLE", "HOST IP-ADDRESSES", "PRIMARY IF-ADDRESS", "HOST GATEWAY-IP"}, " "),
				strings.Join([]string{"control-plane-0", "control-plane,master", "192.168.1.10/24", "192.168.1.10/24", "192.168.1.1"}, " "),
				strings.Join([]string{"control-plane-1", "control-plane,master", "192.168.1.11/24", "192.168.1.11/24", "192.168.1.1"}, " "),
				strings.Join([]string{"control-plane-2", "control-plane,master", "192.168.1.12/24", "192.168.1.12/24", "192.168.1.1"}, " "),
				strings.Join([]string{"worker-0", "worker", "192.168.1.100/24", "192.168.1.100/24", "192.168.1.1"}, " "),
				strings.Join([]string{"worker-1", "worker", "192.168.1.101/24", "192.168.1.101/24", "192.168.1.1"}, " "),
				strings.Join([]string{"worker-2", "worker", "192.168.1.102/24", "192.168.1.102/24", "192.168.1.1"}, " ")},
			false,
		),
		Entry("should get extra info",
			testdataMustGatherPath, mgtypes.InfoTypeExtraInfo,
			[]string{strings.Join([]string{"HOST/NODE", "ROLE", "NODE ID", "NODE CHASSIS-ID"}, " "),
				strings.Join([]string{"control-plane-0", "control-plane,master", "1", "a1b2c3d4-e5f6-4789-a012-3456789abcde"}, " "),
				strings.Join([]string{"control-plane-1", "control-plane,master", "2", "b2c3d4e5-f6a7-4890-b123-456789abcdef"}, " "),
				strings.Join([]string{"control-plane-2", "control-plane,master", "3", "c3d4e5f6-a7b8-4901-c234-56789abcdef0"}, " "),
				strings.Join([]string{"worker-0", "worker", "4", "d4e5f6a7-b8c9-4012-d345-6789abcdef01"}, " "),
				strings.Join([]string{"worker-1", "worker", "5", "e5f6a7b8-c9d0-4123-e456-789abcdef012"}, " "),
				strings.Join([]string{"worker-2", "worker", "6", "f6a7b8c9-d0e1-4234-f567-89abcdef0123"}, " ")},
			false,
		),
		Entry("should fail with invalid info type",
			testdataMustGatherPath, mgtypes.InfoType("invalid"),
			nil,
			true,
		),
	)

	// OVSDB tool tests - these tests will fail if ovsdb-tool is not available in PATH
	// To run these tests, install ovsdb-tool: dnf install openvswitch (RHEL/Fedora) or apt-get install openvswitch-common (Ubuntu)
	Context("OVSDB Tools", Label("requires-ovsdb-tool"), func() {
		type dbInfo struct {
			Database string `json:"database"`
			Node     string `json:"node"`
		}
		DescribeTable("List OVN Databases",
			func(toolName string, mustGatherPath string, shouldFail bool, expectedOutput []dbInfo) {
				By("Calling " + toolName)
				output, err := mcpInspector.
					MethodCall(toolName, map[string]any{
						"must_gather_path": mustGatherPath,
					}).Execute()
				Expect(err).NotTo(HaveOccurred())

				if shouldFail {
					Expect(string(output)).To(ContainSubstring(`"isError": true`))
				} else {
					Expect(output).NotTo(BeEmpty())

					By("Checking the result")
					result := utils.UnmarshalCallToolResult[mgtypes.ListDatabasesResult](output)
					Expect(result.Data).NotTo(BeEmpty())

					// Unmarshal the result data into a slice of dbInfo
					actualOutput := []dbInfo{}
					err = json.Unmarshal([]byte(result.Data), &actualOutput)
					Expect(err).NotTo(HaveOccurred())

					// Verify that the number of databases is equal to the number of expected databases
					Expect(len(actualOutput)).To(Equal(len(expectedOutput)))
					// Verify that the database to node mapping is correct by creating a map of the actual output
					dbToNodeMap := make(map[string]string)
					for _, actual := range actualOutput {
						dbToNodeMap[actual.Database] = actual.Node
					}
					// Verify that the database to node mapping is correct by checking the expected output
					for _, expected := range expectedOutput {
						Expect(dbToNodeMap[expected.Database]).To(Equal(expected.Node))
					}
				}
			},
			Entry("should list northbound databases",
				listNorthboundDatabasesName, testdataMustGatherPath, false,
				[]dbInfo{
					{Database: "ovnkube-node-cp0_nbdb", Node: "control-plane-0"},
					{Database: "ovnkube-node-cp1_nbdb", Node: "control-plane-1"},
					{Database: "ovnkube-node-cp2_nbdb", Node: "control-plane-2"},
					{Database: "ovnkube-node-w0_nbdb", Node: "worker-0"},
					{Database: "ovnkube-node-w1_nbdb", Node: "worker-1"},
					{Database: "ovnkube-node-w2_nbdb", Node: "worker-2"},
				},
			),
			Entry("should list southbound databases",
				listSouthboundDatabasesName, testdataMustGatherPath, false,
				[]dbInfo{
					{Database: "ovnkube-node-cp0_sbdb", Node: "control-plane-0"},
					{Database: "ovnkube-node-cp1_sbdb", Node: "control-plane-1"},
					{Database: "ovnkube-node-cp2_sbdb", Node: "control-plane-2"},
					{Database: "ovnkube-node-w0_sbdb", Node: "worker-0"},
					{Database: "ovnkube-node-w1_sbdb", Node: "worker-1"},
					{Database: "ovnkube-node-w2_sbdb", Node: "worker-2"},
				},
			),
		)

		DescribeTable("Query Database",
			func(mustGatherPath string, databaseName string, table string, conditions []string, columns []string, expectedSubstrings []string, shouldNotContainSubstrings []string, shouldFail bool) {
				By("Calling must-gather-query-database")
				output, err := mcpInspector.
					MethodCall(queryDatabaseToolName, map[string]any{
						"must_gather_path": mustGatherPath,
						"database_name":    databaseName,
						"table":            table,
						"conditions":       conditions,
						"columns":          columns,
					}).Execute()
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

					result := utils.UnmarshalCallToolResult[mgtypes.QueryDatabaseResult](output)
					Expect(result.Data).NotTo(BeEmpty())

					// Verify output contains expected data
					for _, expected := range expectedSubstrings {
						Expect(result.Data).To(ContainSubstring(expected))
					}

					// Verify output does not contain unexpected data
					for _, shouldNotContainSubstring := range shouldNotContainSubstrings {
						Expect(result.Data).NotTo(ContainSubstring(shouldNotContainSubstring))
					}
				}
			},
			Entry("should query Chassis table from southbound database",
				testdataMustGatherPath, "ovnkube-node-w0_sbdb", "Chassis",
				nil,
				nil,
				[]string{"test-chassis", "hostname"},
				nil,
				false,
			),
			Entry("should query Chassis table with columns from southbound database",
				testdataMustGatherPath, "ovnkube-node-w0_sbdb", "Chassis",
				nil,
				[]string{"name", "hostname"},
				[]string{"test-chassis", "worker-0", "worker-1"},
				[]string{"_uuid", "_version", "encaps", "external_ids", "nb_cfg", "other_config", "transport_zones", "vtep_logical_switches"},
				false,
			),
			Entry("should query Encap table from southbound database",
				testdataMustGatherPath, "ovnkube-node-w0_sbdb", "Encap",
				nil,
				nil,
				[]string{"geneve", "chassis_name", "192.168.1"},
				nil,
				false,
			),
			Entry("should query Logical_Switch table from northbound database",
				testdataMustGatherPath, "ovnkube-node-w0_nbdb", "Logical_Switch",
				nil,
				nil,
				[]string{"test-switch", "test-cluster"},
				nil,
				false,
			),
			Entry("should query Logical_Switch table with columns from northbound database",
				testdataMustGatherPath, "ovnkube-node-w0_nbdb", "Logical_Switch",
				nil,
				[]string{"name", "external_ids"},
				[]string{"test-switch-1", "test-switch-2"},
				[]string{"_uuid", "_version", "acls", "copp", "dns_records", "forwarding_groups", "load_balancer", "load_balancer_group", "other_config", "ports", "qos_rules"},
				false,
			),
			Entry("should fail with non-existent must-gather path",
				"/path/that/does/not/exist", "ovnkube-node-w0_sbdb", "Chassis",
				nil,
				nil,
				nil,
				nil,
				true,
			),
			Entry("should fail with invalid table name",
				testdataMustGatherPath, "ovnkube-node-w0_sbdb", "InvalidTable",
				nil,
				nil,
				nil,
				nil,
				true,
			),
			Entry("should fail with non-existent column name (southbound)",
				testdataMustGatherPath, "ovnkube-node-w0_sbdb", "Chassis",
				nil,
				[]string{"name", "NonExistentColumn"},
				nil,
				nil,
				true,
			),
			Entry("should fail with non-existent column name (northbound)",
				testdataMustGatherPath, "ovnkube-node-w0_nbdb", "Logical_Switch",
				nil,
				[]string{"FakeColumn"},
				nil,
				nil,
				true,
			),
		)
	})
})

func matchExpectedLines(expectedLines []string, actualData string, header bool, cutBy string) {
	// Split the actual data into lines
	actualLines := strings.Split(strings.TrimSpace(actualData), "\n")
	if header {
		// The first line is the header, so we need to subtract 1.
		Expect(len(actualLines) - 1).To(Equal(len(expectedLines)))
	} else {
		Expect(len(actualLines)).To(Equal(len(expectedLines)))
	}

	matched := make([]bool, len(expectedLines))

	// Verify that the data contains the expected content
	for i, actual := range actualLines {
		// If the header is true, we need to skip the first line.
		if i == 0 && header {
			continue
		}
		// Remove the extra spaces from the line
		actual = strings.Join(strings.Fields(actual), " ")
		matchIdx := -1
		for j, expected := range expectedLines {
			// If the expected content has already been matched, skip it.
			if matched[j] {
				continue
			}
			// If cutBy is set, we need to split the expected content by the cutBy separator
			// and verify that the actual content contains both parts.
			if cutBy != "" {
				before, after, found := strings.Cut(expected, cutBy)
				Expect(found).To(BeTrue())
				if strings.Contains(actual, before) &&
					strings.Contains(actual, after) &&
					strings.Contains(actual, cutBy) {
					Expect(matchIdx).To(Equal(-1))
					matchIdx = j
				}
			} else {
				// If cutBy is not set, we can verify that the actual content contains the expected content.
				if strings.Contains(actual, expected) {
					Expect(matchIdx).To(Equal(-1))
					matchIdx = j
				}
			}
		}
		Expect(matchIdx).To(Not(Equal(-1)))
		matched[matchIdx] = true
	}
	// Verify that all expected content was matched exactly once
	for _, matched := range matched {
		Expect(matched).To(BeTrue())
	}
}
