package types

type DebugNodeParams struct {
	NodeName  string `json:"node_name" jsonschema:"Node name"`
	NodeImage string `json:"node_image" jsonschema:"Node Image"`
}

type ExecPodParams struct {
	PodName       string `json:"pod_name" jsonschema:"Pod name"`
	PodNamespace  string `json:"pod_namespace" jsonschema:"Pod namespace (use default namespace if not specified)"`
	ContainerName string `json:"container_name,omitempty" jsonschema:"Container name (optional)"`
}

type TcpdumpParams struct {
	TargetType string `json:"target_type" jsonschema:"required,Target type: node or pod"`

	// Node-specific fields
	NodeName  string `json:"node_name,omitempty" jsonschema:"Node name (required if target_type=node)"`
	NodeImage string `json:"node_image,omitempty" jsonschema:"Debug pod image (required if target_type=node)"`

	// Pod-specific fields
	PodName       string `json:"pod_name,omitempty" jsonschema:"Pod name (required if target_type=pod)"`
	PodNamespace  string `json:"pod_namespace,omitempty" jsonschema:"Pod namespace (use default namespace if not specified)"`
	ContainerName string `json:"container_name,omitempty" jsonschema:"Container name (optional)"`

	Command      []string `json:"command,omitempty" jsonschema:"Command to execute"`
	Interface    string   `json:"interface" jsonschema:"required,Network interface name or 'any'"`
	Duration     int      `json:"duration,omitempty" jsonschema:"Capture duration in seconds (max: 30)"`
	PacketCount  int      `json:"packet_count,omitempty" jsonschema:"Number of packets to capture (max: 1000)"`
	BPFFilter    string   `json:"bpf_filter,omitempty" jsonschema:"BPF filter expression (e.g. 'host 10.0.0.1 and port 80')"`
	Snaplen      int      `json:"snaplen,omitempty" jsonschema:"Snapshot length in bytes (default: 96, max: 262)"`
	OutputFormat string   `json:"output_format,omitempty" jsonschema:"Output format: text or pcap (default: text)"`
}

type CommandResult struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code,omitempty"`
	Error    string `json:"error,omitempty"`
}
