package types

// DebugNodeParams contains parameters for running debug commands on a Kubernetes node.
type DebugNodeParams struct {
	NodeName  string `json:"node_name" jsonschema:"Node name"`
	NodeImage string `json:"node_image" jsonschema:"Node Image"`
	HostPath  string `json:"-"`
	MountPath string `json:"-"`
}

// ExecPodParams contains parameters for executing commands inside a pod container.
type ExecPodParams struct {
	PodName       string `json:"pod_name" jsonschema:"Pod name"`
	PodNamespace  string `json:"pod_namespace" jsonschema:"Pod namespace (use default namespace if not specified)"`
	ContainerName string `json:"container_name,omitempty" jsonschema:"Container name (optional)"`
}

// TcpdumpParams contains parameters for running tcpdump packet capture.
// Supports both node-level and pod-level packet capture with BPF filtering.
type TcpdumpParams struct {
	TargetType string `json:"target_type" jsonschema:"required,Target type: node or pod"`

	// Node-specific fields
	NodeName  string `json:"node_name,omitempty" jsonschema:"Node name (required if target_type=node)"`
	NodeImage string `json:"node_image,omitempty" jsonschema:"Debug pod image (required if target_type=node)"`

	// Pod-specific fields
	PodName       string `json:"pod_name,omitempty" jsonschema:"Pod name (required if target_type=pod)"`
	PodNamespace  string `json:"pod_namespace,omitempty" jsonschema:"Pod namespace (use default namespace if not specified)"`
	ContainerName string `json:"container_name,omitempty" jsonschema:"Container name (optional)"`

	Interface    string `json:"interface,omitempty" jsonschema:"Network interface name or 'any' (optional, tcpdump will use default if not specified)"`
	PacketCount  int    `json:"packet_count,omitempty" jsonschema:"Number of packets to capture (default: 1000, max: 1000)"`
	BPFFilter    string `json:"bpf_filter,omitempty" jsonschema:"BPF filter expression (e.g. 'host 10.0.0.1 and port 80')"`
	Snaplen      int    `json:"snaplen,omitempty" jsonschema:"Snapshot length in bytes (default: 96, max: 262)"`
	OutputFormat string `json:"output_format,omitempty" jsonschema:"Output format: text or pcap (default: text)"`
}

// PwruParams contains parameters for running pwru (packet, where are you?) eBPF-based
// kernel packet tracing. This tool traces packets through the Linux kernel networking stack.
type PwruParams struct {
	NodeName string `json:"node_name" jsonschema:"Node name (required)"`

	// Pwru-specific fields
	PcapFilter       string `json:"pcap_filter,omitempty" jsonschema:"pcap-filter expression (e.g. 'tcp and dst port 8080')"`
	OutputLimitLines int    `json:"output_limit_lines,omitempty" jsonschema:"Exit after number of events (default: 1000, max: 1000)"`
}

// CommandResult represents the output and status of an executed command.
type CommandResult struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code,omitempty"`
	Error    string `json:"error,omitempty"`
}
