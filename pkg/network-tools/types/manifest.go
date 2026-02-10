package types

// BaseNetworkDiagParams contains common parameters shared across network diagnostic tools.
type BaseNetworkDiagParams struct {
	BPFFilter string `json:"bpf_filter,omitempty"`
}

// TcpdumpParams contains parameters for running tcpdump packet capture.
// Supports both node-level and pod-level packet capture with BPF filtering.
type TcpdumpParams struct {
	BaseNetworkDiagParams

	TargetType string `json:"target_type"`

	// Node-specific fields
	NodeName string `json:"node_name,omitempty"`

	// Pod-specific fields
	PodName       string `json:"pod_name,omitempty"`
	PodNamespace  string `json:"pod_namespace,omitempty"`
	ContainerName string `json:"container_name,omitempty"`

	Interface   string `json:"interface,omitempty"`
	PacketCount int    `json:"packet_count,omitempty"`
	Snaplen     int    `json:"snaplen,omitempty"`
}

// PwruParams contains parameters for running pwru (packet, where are you?) eBPF-based
// kernel packet tracing. This tool traces packets through the Linux kernel networking stack.
type PwruParams struct {
	BaseNetworkDiagParams

	NodeName         string `json:"node_name"`
	OutputLimitLines int    `json:"output_limit_lines,omitempty"`
}

// CommandResult represents the output and status of an executed command.
type CommandResult struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code,omitempty"`
	Error    string `json:"error,omitempty"`
}
