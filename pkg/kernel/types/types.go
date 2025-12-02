package types

// CommonParams contains the common parameters required for executing kernel commands on a Kubernetes node.
// These parameters are embedded in other specific parameter types.
type CommonParams struct {
	Node     string `json:"node"`                // Node is the name of the Kubernetes node where the command will be executed
	Image    string `json:"image"`               // Image is the container image to use for running the command
	MaxLines int    `json:"max_lines,omitempty"` // Limit the number of lines in output
}

// ListConntrackParams contains parameters for listing connection tracking entries using conntrack command.
// Connection tracking entries show active network connections and their state.
type ListConntrackParams struct {
	CommonParams
	Command          string `json:"command,omitempty"`           // Command specifies the conntrack command to execute (e.g., "list", "dump")
	FilterParameters string `json:"filter_parameters,omitempty"` // FilterParameters specifies additional filter criteria for conntrack entries
}

// ListIPTablesParams contains parameters for inspecting iptables/ip6tables packet filter rules.
// Supports both IPv4 (iptables) and IPv6 (ip6tables) firewall rules.
type ListIPTablesParams struct {
	CommonParams
	Table            string `json:"table,omitempty"`             // Table specifies the iptables table to query (e.g., "filter", "nat", "mangle", "raw")
	Command          string `json:"command"`                     // Command specifies the iptables command to execute (e.g., "iptables", "ip6tables")
	FilterParameters string `json:"filter_parameters,omitempty"` // FilterParameters specifies additional filter criteria for iptables rules
}

// ListNFTParams contains parameters for inspecting nftables packet filtering and classification rules.
// nftables is the modern replacement for iptables in the Linux kernel.
type ListNFTParams struct {
	CommonParams
	Command         string `json:"command"`                    // Command specifies the nft command to execute
	AddressFamilies string `json:"address_families,omitempty"` // AddressFamilies specifies the address family to filter (e.g., "ip", "ip6", "inet", "arp", "bridge")
}

// ListIPParams contains parameters for inspecting routing, network devices, and network interfaces.
// Uses the iproute2 suite (ip command) for network configuration and inspection.
type ListIPParams struct {
	CommonParams
	Options          string `json:"options,omitempty"`           // Options specifies additional command-line options for the ip command
	Command          string `json:"command"`                     // Command specifies the ip subcommand to execute (e.g., "route", "link", "addr", "neigh")
	FilterParameters string `json:"filter_parameters,omitempty"` // FilterParameters specifies additional filter criteria for the output
}

// Result represents the output returned from executing a kernel command.
// The data contains the command's stdout/stderr output.
type Result struct {
	Data string `json:"data"` // Data contains the command execution output
}
