package e2e

import (
	"context"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

var _ = Describe("Network Tools", func() {
	const (
		tcpdumpToolName = "tcpdump"
		pwruToolName    = "pwru"
	)

	var nodeName string

	// generateTraffic creates deterministic network traffic for reliable packet capture tests
	generateTraffic := func() {
		By("Generating real network traffic")
		// Send ICMP packets to generate actual traffic on the wire that tcpdump/pwru can capture.
		// Note: "ip route show" just reads local routing table (no packets), so we use ping instead.
		_, err := mcpInspector.
			MethodCall("get-ip", map[string]any{
				"target_type": "node",
				"node_name":   nodeName,
				"command":     "ping",
				"args":        []string{"-c", "3", "8.8.8.8"},
			}).Execute()
		Expect(err).NotTo(HaveOccurred())

		// Small sleep to ensure traffic is generated before capture
		time.Sleep(100 * time.Millisecond)
	}

	// BeforeEach finds a ready worker node once per test
	BeforeEach(func() {
		nodeName = ""
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		nodeList := &corev1.NodeList{}
		err := kubeClient.List(ctx, nodeList)
		Expect(err).NotTo(HaveOccurred(), "Failed to list nodes")

		// Look for a ready worker node (not control-plane)
		for _, node := range nodeList.Items {
			isReady := false
			isControlPlane := false

			for _, condition := range node.Status.Conditions {
				if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
					isReady = true
					break
				}
			}

			if _, hasLabel := node.Labels["node-role.kubernetes.io/control-plane"]; hasLabel {
				isControlPlane = true
			}
			if _, hasLabel := node.Labels["node-role.kubernetes.io/master"]; hasLabel {
				isControlPlane = true
			}

			if isReady && !isControlPlane {
				nodeName = node.Name
				break
			}
		}

		if nodeName == "" {
			// If no worker nodes found, use any ready node
			for _, node := range nodeList.Items {
				isReady := false
				for _, condition := range node.Status.Conditions {
					if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
						isReady = true
						break
					}
				}
				if isReady {
					nodeName = node.Name
					break
				}
			}
		}

		Expect(nodeName).NotTo(BeEmpty(), "No ready nodes found")
	})

	Context("tcpdump", func() {
		It("should capture network packets with minimal parameters", func() {
			generateTraffic()
			By("Running tcpdump with basic node capture")
			output, err := mcpInspector.
				MethodCall(tcpdumpToolName, map[string]any{
					"target_type":  "node",
					"node_name":    nodeName,
					"packet_count": 2,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains packet capture data")
			result := utils.UnmarshalCallToolResult[map[string]any](output)
			outputStr, exists := result["output"].(string)
			Expect(exists).To(BeTrue(), "Expected output field to exist")
			Expect(outputStr).NotTo(BeEmpty())

			By("Verifying tcpdump output contains network traffic")
			GinkgoWriter.Print("\n=== tcpdump Basic Capture ===\n")
			GinkgoWriter.Print("Node: ", nodeName, "\n")
			GinkgoWriter.Print("Captured packets:\n", outputStr, "\n")
			GinkgoWriter.Print("=============================\n\n")

			// tcpdump output should contain IP traffic indicators
			Expect(outputStr).To(Or(
				ContainSubstring("IP "),
				ContainSubstring("TCP "),
				ContainSubstring("UDP "),
				ContainSubstring(" > "),
			))
		})

		It("should capture network packets with interface 'any'", func() {
			generateTraffic()
			By("Running tcpdump with any interface")
			output, err := mcpInspector.
				MethodCall(tcpdumpToolName, map[string]any{
					"target_type":  "node",
					"node_name":    nodeName,
					"packet_count": 3,
					"interface":    "any",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains packet capture data")
			result := utils.UnmarshalCallToolResult[map[string]any](output)
			outputStr, exists := result["output"].(string)
			Expect(exists).To(BeTrue(), "Expected output field to exist")
			Expect(outputStr).NotTo(BeEmpty())

			By("Verifying tcpdump output with any interface")
			GinkgoWriter.Print("\n=== tcpdump Any Interface ===\n")
			GinkgoWriter.Print("Node: ", nodeName, "\n")
			GinkgoWriter.Print("Interface: any\n")
			GinkgoWriter.Print("Captured packets:\n", outputStr, "\n")
			GinkgoWriter.Print("=============================\n\n")

			// Should capture traffic from any interface
			Expect(outputStr).To(Or(
				ContainSubstring("IP "),
				ContainSubstring("In "),
				ContainSubstring("Out "),
				ContainSubstring("lo "),
			))
		})

		It("should capture network packets with BPF filter", func() {
			generateTraffic()
			By("Running tcpdump with TCP filter")
			output, err := mcpInspector.
				MethodCall(tcpdumpToolName, map[string]any{
					"target_type":  "node",
					"node_name":    nodeName,
					"packet_count": 5,
					"interface":    "any",
					"bpf_filter":   "tcp",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains filtered packet data")
			result := utils.UnmarshalCallToolResult[map[string]any](output)
			outputStr, exists := result["output"].(string)
			Expect(exists).To(BeTrue(), "Expected output field to exist")

			By("Verifying tcpdump BPF filter output")
			GinkgoWriter.Print("\n=== tcpdump TCP Filter ===\n")
			GinkgoWriter.Print("Node: ", nodeName, "\n")
			GinkgoWriter.Print("Filter: tcp\n")
			GinkgoWriter.Print("Captured packets:\n", outputStr, "\n")
			GinkgoWriter.Print("==========================\n\n")

			// With TCP filter, output should either be empty or contain TCP traffic
			if outputStr != "" {
				Expect(strings.ToLower(outputStr)).To(ContainSubstring("tcp"))
			}
		})

		It("should handle custom snapshot length", func() {
			generateTraffic()
			By("Running tcpdump with custom snaplen")
			output, err := mcpInspector.
				MethodCall(tcpdumpToolName, map[string]any{
					"target_type":  "node",
					"node_name":    nodeName,
					"packet_count": 2,
					"interface":    "any",
					"snaplen":      128,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains packet data with custom snaplen")
			result := utils.UnmarshalCallToolResult[map[string]any](output)
			outputStr, exists := result["output"].(string)
			Expect(exists).To(BeTrue(), "Expected output field to exist")

			By("Verifying tcpdump custom snaplen output")
			GinkgoWriter.Print("\n=== tcpdump Custom Snaplen ===\n")
			GinkgoWriter.Print("Node: ", nodeName, "\n")
			GinkgoWriter.Print("Snaplen: 128 bytes\n")
			GinkgoWriter.Print("Captured packets:\n", outputStr, "\n")
			GinkgoWriter.Print("==============================\n\n")

			// Should successfully capture with custom snaplen
			if outputStr != "" {
				Expect(outputStr).To(ContainSubstring("IP "))
			}
		})
	})

	Context("pwru", func() {
		It("should trace kernel network stack", func() {
			// Skip pwru tests in Kind clusters as they require full kernel capabilities
			if os.Getenv("CI") == "true" {
				Skip("pwru tests are skipped in CI/Kind environments due to limited kernel capabilities")
			}
			
			generateTraffic()
			By("Running pwru to trace kernel networking")
			output, err := mcpInspector.
				MethodCall(pwruToolName, map[string]any{
					"node_name":          nodeName,
					"output_limit_lines": 10,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains kernel trace data")
			result := utils.UnmarshalCallToolResult[map[string]any](output)
			outputStr, exists := result["output"].(string)
			Expect(exists).To(BeTrue(), "Expected output field to exist")
			Expect(outputStr).NotTo(BeEmpty())

			By("Verifying pwru kernel trace output")
			GinkgoWriter.Print("\n=== pwru Kernel Trace ===\n")
			GinkgoWriter.Print("Node: ", nodeName, "\n")
			GinkgoWriter.Print("Kernel trace:\n", outputStr, "\n")
			GinkgoWriter.Print("=========================\n\n")

			// pwru output should contain kernel function names and SKB traces
			Expect(outputStr).To(Or(
				ContainSubstring("SKB"),
				ContainSubstring("CPU"),
				ContainSubstring("PROCESS"),
				ContainSubstring("NETNS"),
				ContainSubstring("FUNC"),
			))
		})

		It("should trace with BPF filter", func() {
			// Skip pwru tests in Kind clusters as they require full kernel capabilities
			if os.Getenv("CI") == "true" {
				Skip("pwru tests are skipped in CI/Kind environments due to limited kernel capabilities")
			}
			
			generateTraffic()
			By("Running pwru with host filter")
			output, err := mcpInspector.
				MethodCall(pwruToolName, map[string]any{
					"node_name":          nodeName,
					"bpf_filter":         "tcp",
					"output_limit_lines": 8,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains filtered trace data")
			result := utils.UnmarshalCallToolResult[map[string]any](output)
			outputStr, exists := result["output"].(string)
			Expect(exists).To(BeTrue(), "Expected output field to exist")

			By("Verifying pwru filtered trace output")
			GinkgoWriter.Print("\n=== pwru TCP Filter ===\n")
			GinkgoWriter.Print("Node: ", nodeName, "\n")
			GinkgoWriter.Print("Filter: tcp\n")
			GinkgoWriter.Print("Kernel trace:\n", outputStr, "\n")
			GinkgoWriter.Print("=======================\n\n")

			// With filter, should still show kernel tracing structure
			if outputStr != "" {
				Expect(outputStr).To(Or(
					ContainSubstring("SKB"),
					ContainSubstring("FUNC"),
					ContainSubstring("CPU"),
				))
			}
		})

		It("should limit output lines correctly", func() {
			// Skip pwru tests in Kind clusters as they require full kernel capabilities
			if os.Getenv("CI") == "true" {
				Skip("pwru tests are skipped in CI/Kind environments due to limited kernel capabilities")
			}
			
			generateTraffic()
			By("Running pwru with small line limit")
			output, err := mcpInspector.
				MethodCall(pwruToolName, map[string]any{
					"node_name":          nodeName,
					"output_limit_lines": 3,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result respects line limits")
			result := utils.UnmarshalCallToolResult[map[string]any](output)
			outputStr, exists := result["output"].(string)
			Expect(exists).To(BeTrue(), "Expected output field to exist")

			By("Verifying pwru line limit output")
			GinkgoWriter.Print("\n=== pwru Line Limit ===\n")
			GinkgoWriter.Print("Node: ", nodeName, "\n")
			GinkgoWriter.Print("Limit: 3 lines\n")
			GinkgoWriter.Print("Kernel trace:\n", outputStr, "\n")
			GinkgoWriter.Print("=======================\n\n")

			// Should respect the requested line limit of 3
			if outputStr != "" {
				lines := strings.Split(strings.TrimSpace(outputStr), "\n")
				// We requested output_limit_lines: 3
				// pwru output format: 1 header line (SKB, CPU, PROCESS, ...) + N data lines
				// So output_limit_lines=3 should produce: 1 header + 3 data = 4 total lines
				// Changed from <=5 to <=4 per review feedback to tighten validation
				Expect(len(lines)).To(BeNumerically(">=", 1), "Should have at least header line")
				Expect(len(lines)).To(BeNumerically("<=", 4), "Should respect line limit parameter")
			}
		})
	})

})
