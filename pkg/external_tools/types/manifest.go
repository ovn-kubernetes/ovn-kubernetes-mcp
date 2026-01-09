package types

// TcpdumpParams contains parameters for running tcpdump packet capture.
// Supports both node-level and pod-level packet capture with BPF filtering.
// For node-level captures, a hardcoded secure image (nicolaka/netshoot:v0.13) is used.
type TcpdumpParams struct {
	TargetType string `json:"target_type"`

	// Node-specific fields
	NodeName string `json:"node_name,omitempty"`

	// Pod-specific fields
	PodName       string `json:"pod_name,omitempty"`
	PodNamespace  string `json:"pod_namespace,omitempty"`
	ContainerName string `json:"container_name,omitempty"`

	Interface    string `json:"interface,omitempty"`
	PacketCount  int    `json:"packet_count,omitempty"`
	BPFFilter    string `json:"bpf_filter,omitempty"`
	Snaplen      int    `json:"snaplen,omitempty"`
	OutputFormat string `json:"output_format,omitempty"`
}

// PwruParams contains parameters for running pwru (packet, where are you?) eBPF-based
// kernel packet tracing. This tool traces packets through the Linux kernel networking stack.
type PwruParams struct {
	NodeName string `json:"node_name"`

	// Pwru-specific fields
	PcapFilter       string `json:"pcap_filter,omitempty"`
	OutputLimitLines int    `json:"output_limit_lines,omitempty"`
}

// CommandResult represents the output and status of an executed command.
type CommandResult struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code,omitempty"`
	Error    string `json:"error,omitempty"`
}
