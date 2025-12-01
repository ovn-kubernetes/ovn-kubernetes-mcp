package e2e

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sostypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/sosreport/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/inspector"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

var _ = Describe("Sosreport Tools", func() {
	var sosreportInspector *inspector.MCPInspector
	var testdataSosreportPath string

	BeforeEach(func() {
		mcpServerPath := os.Getenv(mcpServerPathEnvVar)
		Expect(mcpServerPath).NotTo(BeEmpty())

		sosreportInspector = inspector.NewMCPInspector().
			Command(mcpServerPath).
			CommandFlags(map[string]string{
				"mode": "offline",
			})

		var err error
		testdataSosreportPath, err = filepath.Abs("../../pkg/sosreport/mcp/testdata/sosreport")
		Expect(err).NotTo(HaveOccurred())
	})
	const (
		sosListPluginsToolName    = "sos-list-plugins"
		sosListCommandsToolName   = "sos-list-commands"
		sosSearchCommandsToolName = "sos-search-commands"
		sosGetCommandToolName     = "sos-get-command"
		sosSearchPodLogsToolName  = "sos-search-pod-logs"
	)

	Context("List Plugins", func() {
		It("should list all enabled plugins with command counts", func() {
			By("Calling sos-list-plugins")
			output, err := sosreportInspector.
				MethodCall(sosListPluginsToolName, map[string]string{
					"sosreport_path": testdataSosreportPath,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			result := utils.UnmarshalCallToolResult[sostypes.ListPluginsResult](output)

			// Should have 3 plugins
			Expect(result.Plugins).To(HaveLen(3))
			Expect(result.TotalCommands).To(Equal(3))

			// Verify plugin names and command counts
			pluginMap := make(map[string]int)
			for _, plugin := range result.Plugins {
				pluginMap[plugin.Name] = plugin.CommandCount
			}
			Expect(pluginMap["openvswitch"]).To(Equal(2))
			Expect(pluginMap["networking"]).To(Equal(1))
			Expect(pluginMap["container_log"]).To(Equal(0))
		})
	})

	Context("List Commands", func() {
		It("should list all commands for openvswitch plugin", func() {
			By("Calling sos-list-commands for openvswitch")
			output, err := sosreportInspector.
				MethodCall(sosListCommandsToolName, map[string]string{
					"sosreport_path": testdataSosreportPath,
					"plugin":         "openvswitch",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			result := utils.UnmarshalCallToolResult[sostypes.ListCommandsResult](output)

			Expect(result.Plugin).To(Equal("openvswitch"))
			Expect(result.CommandCount).To(Equal(2))
			Expect(result.Commands).To(HaveLen(2))

			// Verify command details
			commandMap := make(map[string]string)
			for _, cmd := range result.Commands {
				commandMap[cmd.Exec] = cmd.Filepath
			}
			Expect(commandMap["ovs-vsctl -t 5 show"]).To(Equal("sos_commands/openvswitch/ovs-vsctl_-t_5_show"))
			Expect(commandMap["ovs-ofctl dump-flows br-int"]).To(Equal("sos_commands/openvswitch/ovs-ofctl_dump-flows_br-int"))
		})

		It("should list all commands for networking plugin", func() {
			By("Calling sos-list-commands for networking")
			output, err := sosreportInspector.
				MethodCall(sosListCommandsToolName, map[string]string{
					"sosreport_path": testdataSosreportPath,
					"plugin":         "networking",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			result := utils.UnmarshalCallToolResult[sostypes.ListCommandsResult](output)

			Expect(result.Plugin).To(Equal("networking"))
			Expect(result.CommandCount).To(Equal(1))
			Expect(result.Commands).To(HaveLen(1))
			Expect(result.Commands[0].Exec).To(Equal("ip addr show"))
		})
	})

	Context("Search Commands", func() {
		It("should find commands matching ovs pattern", func() {
			By("Calling sos-search-commands with pattern 'ovs'")
			output, err := sosreportInspector.
				MethodCall(sosSearchCommandsToolName, map[string]string{
					"sosreport_path": testdataSosreportPath,
					"pattern":        "ovs",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			result := utils.UnmarshalCallToolResult[sostypes.SearchCommandsResult](output)

			Expect(result.Total).To(Equal(2))
			Expect(result.Matches).To(HaveLen(2))

			// Verify all matches are from openvswitch plugin
			for _, match := range result.Matches {
				Expect(match.Plugin).To(Equal("openvswitch"))
				Expect(match.Exec).To(ContainSubstring("ovs"))
			}
		})

		It("should find commands matching ip pattern", func() {
			By("Calling sos-search-commands with pattern 'ip'")
			output, err := sosreportInspector.
				MethodCall(sosSearchCommandsToolName, map[string]string{
					"sosreport_path": testdataSosreportPath,
					"pattern":        "ip",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			result := utils.UnmarshalCallToolResult[sostypes.SearchCommandsResult](output)

			Expect(result.Total).To(Equal(1))
			Expect(result.Matches).To(HaveLen(1))
			Expect(result.Matches[0].Plugin).To(Equal("networking"))
			Expect(result.Matches[0].Exec).To(Equal("ip addr show"))
		})
	})

	Context("Get Command", func() {
		It("should get command output by filepath", func() {
			By("Calling sos-get-command for ovs-vsctl show")
			output, err := sosreportInspector.
				MethodCall(sosGetCommandToolName, map[string]string{
					"sosreport_path": testdataSosreportPath,
					"filepath":       "sos_commands/openvswitch/ovs-vsctl_-t_5_show",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			result := utils.UnmarshalCallToolResult[sostypes.GetCommandResult](output)

			// Verify output contains expected content
			Expect(result.Output).To(ContainSubstring("Bridge br-int"))
			Expect(result.Output).To(ContainSubstring("ovn-k8s-mp0"))
			Expect(result.Output).To(ContainSubstring("ovs_version: \"2.17.0\""))
		})

		It("should get command output with pattern filter", func() {
			By("Calling sos-get-command with pattern filter")
			output, err := sosreportInspector.
				MethodCall(sosGetCommandToolName, map[string]string{
					"sosreport_path": testdataSosreportPath,
					"filepath":       "sos_commands/networking/ip_addr_show",
					"pattern":        "eth0",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			result := utils.UnmarshalCallToolResult[sostypes.GetCommandResult](output)

			// Should only contain lines matching the pattern
			Expect(result.Output).To(ContainSubstring("eth0"))
			// Verify lo interface is not in filtered output
			lines := strings.Split(strings.TrimSpace(result.Output), "\n")
			for _, line := range lines {
				if !strings.Contains(line, "eth0") {
					Expect(line).To(BeEmpty())
				}
			}
		})

		It("should get full ip addr show output without filter", func() {
			By("Calling sos-get-command without pattern")
			output, err := sosreportInspector.
				MethodCall(sosGetCommandToolName, map[string]string{
					"sosreport_path": testdataSosreportPath,
					"filepath":       "sos_commands/networking/ip_addr_show",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			result := utils.UnmarshalCallToolResult[sostypes.GetCommandResult](output)

			// Should contain both interfaces
			Expect(result.Output).To(ContainSubstring("lo"))
			Expect(result.Output).To(ContainSubstring("eth0"))
			Expect(result.Output).To(ContainSubstring("127.0.0.1"))
			Expect(result.Output).To(ContainSubstring("192.168.1.100"))
		})
	})

	Context("Search Pod Logs", func() {
		It("should search for ERROR pattern in pod logs", func() {
			By("Calling sos-search-pod-logs with ERROR pattern")
			output, err := sosreportInspector.
				MethodCall(sosSearchPodLogsToolName, map[string]string{
					"sosreport_path": testdataSosreportPath,
					"pattern":        "ERROR",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			result := utils.UnmarshalCallToolResult[sostypes.SearchPodLogsResult](output)

			// Should find the error line
			Expect(result.Output).To(ContainSubstring("ERROR"))
			Expect(result.Output).To(ContainSubstring("Failed to connect to ovn-controller"))
			Expect(result.Output).To(ContainSubstring("ovnkube-node"))
		})

		It("should search pod logs with pod filter", func() {
			By("Calling sos-search-pod-logs with pod filter")
			output, err := sosreportInspector.
				MethodCall(sosSearchPodLogsToolName, map[string]string{
					"sosreport_path": testdataSosreportPath,
					"pattern":        "ovnkube",
					"pod_filter":     "ovnkube-node",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			result := utils.UnmarshalCallToolResult[sostypes.SearchPodLogsResult](output)

			// Should find matching lines from the ovnkube-node pod
			Expect(result.Output).To(ContainSubstring("ovnkube-node"))
			Expect(result.Output).To(ContainSubstring("Starting ovnkube-node"))
		})

		It("should search for successful connection in pod logs", func() {
			By("Calling sos-search-pod-logs for successful connection")
			output, err := sosreportInspector.
				MethodCall(sosSearchPodLogsToolName, map[string]string{
					"sosreport_path": testdataSosreportPath,
					"pattern":        "Successfully connected",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			result := utils.UnmarshalCallToolResult[sostypes.SearchPodLogsResult](output)

			Expect(result.Output).To(ContainSubstring("Successfully connected to ovn-controller"))
		})
	})
})
