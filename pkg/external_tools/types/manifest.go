package types

type TargetParams struct {
	TargetType    string `json:"target_type" jsonschema:"Target type: node or pod (required)"`
	NodeName      string `json:"node_name,omitempty" jsonschema:"Node name (required if target_type is node)"`
	NodeImage     string `json:"node_image,omitempty" jsonschema:"Node image (required if target_type is node)"`
	PodName       string `json:"pod_name,omitempty" jsonschema:"Pod name (required if target_type is pod)"`
	PodNamespace  string `json:"pod_namespace,omitempty" jsonschema:"Pod namespace (default: default, used when target_type is pod)"`
	ContainerName string `json:"container_name,omitempty" jsonschema:"Container name (optional, uses first container if not specified)"`
}

type IPAddrShowParams struct {
	TargetParams
	Interface string `json:"interface,omitempty" jsonschema:"Network interface name (optional)"`
}

type IPRouteShowParams struct {
	TargetParams
	Destination string `json:"destination,omitempty" jsonschema:"Destination network or IP (optional)"`
	Table       string `json:"table,omitempty" jsonschema:"Routing table name or number (optional)"`
}

type IPLinkShowParams struct {
	TargetParams
	Interface string `json:"interface,omitempty" jsonschema:"Network interface name (optional)"`
}

type IPNeighShowParams struct {
	TargetParams
	Interface string `json:"interface,omitempty" jsonschema:"Network interface name (optional)"`
	Address   string `json:"address,omitempty" jsonschema:"IP address filter (optional)"`
}

type IPRuleShowParams struct {
	TargetParams
}

type SSParams struct {
	TargetParams
	Protocol   string `json:"protocol,omitempty" jsonschema:"Protocol filter: tcp, udp, or all"`
	State      string `json:"state,omitempty" jsonschema:"Connection state filter: listening, established, or all"`
	Process    bool   `json:"process,omitempty" jsonschema:"Show process information"`
	Numeric    bool   `json:"numeric,omitempty" jsonschema:"Don't resolve service names"`
	PortFilter string `json:"port_filter,omitempty" jsonschema:"Filter by port (e.g., :8080 or sport = :8080)"`
}

type NetstatParams struct {
	TargetParams
	Protocol  string `json:"protocol,omitempty" jsonschema:"Protocol filter: tcp, udp, or all"`
	Listening bool   `json:"listening,omitempty" jsonschema:"Show only listening sockets"`
	Numeric   bool   `json:"numeric,omitempty" jsonschema:"Don't resolve names"`
}

type PingParams struct {
	TargetParams
	Target   string `json:"target" jsonschema:"Target hostname or IP address (required)"`
	Count    int    `json:"count,omitempty" jsonschema:"Number of packets to send (default: 4, max: 10)"`
	Timeout  int    `json:"timeout,omitempty" jsonschema:"Timeout in seconds (default: 5, max: 10)"`
	Interval int    `json:"interval,omitempty" jsonschema:"Interval between packets in seconds (default: 1, min: 1)"`
	IPv6     bool   `json:"ipv6,omitempty" jsonschema:"Use IPv6"`
}

type TracerouteParams struct {
	TargetParams
	Target  string `json:"target" jsonschema:"Target hostname or IP address (required)"`
	MaxHops int    `json:"max_hops,omitempty" jsonschema:"Maximum number of hops (default: 30, max: 30)"`
	Timeout int    `json:"timeout,omitempty" jsonschema:"Timeout per hop in seconds (default: 5)"`
	IPv6    bool   `json:"ipv6,omitempty" jsonschema:"Use IPv6"`
}

type DigParams struct {
	TargetParams
	Hostname   string `json:"hostname" jsonschema:"Hostname to lookup (required)"`
	Server     string `json:"server,omitempty" jsonschema:"DNS server to query (optional)"`
	RecordType string `json:"record_type,omitempty" jsonschema:"DNS record type: A, AAAA, MX, NS, TXT, CNAME, SOA, or PTR (default: A)"`
	Short      bool   `json:"short,omitempty" jsonschema:"Short output format"`
}

type CurlParams struct {
	TargetParams
	URL            string            `json:"url" jsonschema:"URL to fetch (required)"`
	Method         string            `json:"method,omitempty" jsonschema:"HTTP method: GET, POST, PUT, DELETE, or HEAD (default: GET)"`
	Headers        map[string]string `json:"headers,omitempty" jsonschema:"HTTP headers"`
	Timeout        int               `json:"timeout,omitempty" jsonschema:"Timeout in seconds (default: 10, max: 30)"`
	FollowRedirect bool              `json:"follow_redirect,omitempty" jsonschema:"Follow HTTP redirects"`
	MaxRedirects   int               `json:"max_redirects,omitempty" jsonschema:"Maximum redirects to follow (default: 5, max: 10)"`
	InsecureSSL    bool              `json:"insecure_ssl,omitempty" jsonschema:"Skip SSL certificate verification"`
}

type NetcatParams struct {
	TargetParams
	Host    string `json:"host" jsonschema:"Target host (required)"`
	Port    int    `json:"port" jsonschema:"Target port (required)"`
	UDP     bool   `json:"udp,omitempty" jsonschema:"Use UDP instead of TCP"`
	Timeout int    `json:"timeout,omitempty" jsonschema:"Timeout in seconds (default: 5, max: 10)"`
}

type IPTablesListParams struct {
	TargetParams
	Table       string `json:"table,omitempty" jsonschema:"Table name: filter, nat, mangle, or raw (default: filter)"`
	Chain       string `json:"chain,omitempty" jsonschema:"Chain name (optional)"`
	LineNumbers bool   `json:"line_numbers,omitempty" jsonschema:"Show rule line numbers"`
	Verbose     bool   `json:"verbose,omitempty" jsonschema:"Verbose output"`
	IPv6        bool   `json:"ipv6,omitempty" jsonschema:"Use ip6tables instead of iptables"`
}

type NFTListParams struct {
	TargetParams
	Table string `json:"table,omitempty" jsonschema:"Table name (optional)"`
	Chain string `json:"chain,omitempty" jsonschema:"Chain name (optional)"`
}

type ConntrackListParams struct {
	TargetParams
	Protocol   string `json:"protocol,omitempty" jsonschema:"Protocol filter: tcp, udp, or icmp"`
	SourceIP   string `json:"source_ip,omitempty" jsonschema:"Source IP filter"`
	DestIP     string `json:"dest_ip,omitempty" jsonschema:"Destination IP filter"`
	SourcePort int    `json:"source_port,omitempty" jsonschema:"Source port filter"`
	DestPort   int    `json:"dest_port,omitempty" jsonschema:"Destination port filter"`
}

type ConntrackStatsParams struct {
	TargetParams
}

type EthtoolParams struct {
	TargetParams
	Interface string `json:"interface" jsonschema:"Network interface name (required)"`
	Operation string `json:"operation,omitempty" jsonschema:"Operation type: info, stats, or features (default: info)"`
}

type SysctlNetParams struct {
	TargetParams
	Pattern string `json:"pattern,omitempty" jsonschema:"Pattern to filter parameters (optional)"`
}

type TcpdumpParams struct {
	TargetParams
	Interface    string `json:"interface" jsonschema:"Network interface name or 'any' (required)"`
	Duration     int    `json:"duration,omitempty" jsonschema:"Capture duration in seconds (max: 30)"`
	PacketCount  int    `json:"packet_count,omitempty" jsonschema:"Number of packets to capture (max: 1000)"`
	BPFFilter    string `json:"bpf_filter,omitempty" jsonschema:"BPF filter expression (e.g., 'host 10.0.0.1 and port 80')"`
	Snaplen      int    `json:"snaplen,omitempty" jsonschema:"Snapshot length in bytes (default: 96, max: 262)"`
	OutputFormat string `json:"output_format,omitempty" jsonschema:"Output format: text or pcap (default: text)"`
}

type CommandResult struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code,omitempty"`
	Error    string `json:"error,omitempty"`
}
