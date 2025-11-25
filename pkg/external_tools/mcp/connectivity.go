package external_tools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/client"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/types"
)

func (s *MCPServer) Ping(ctx context.Context, req *mcp.CallToolRequest, in types.PingParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := client.ValidateHostname(in.Target); err != nil {
		return nil, types.CommandResult{}, err
	}

	count, err := validateIntWithDefault(in.Count, 4, 1, 10, "count")
	if err != nil {
		return nil, types.CommandResult{}, err
	}

	timeout, err := validateIntWithDefault(in.Timeout, 5, 1, 10, "timeout")
	if err != nil {
		return nil, types.CommandResult{}, err
	}

	interval, err := validateIntWithDefault(in.Interval, 1, 1, 5, "interval")
	if err != nil {
		return nil, types.CommandResult{}, err
	}

	cmdName := "ping"
	if in.IPv6 {
		cmdName = "ping6"
	}
	command := newCommand(cmdName,
		"-c", strconv.Itoa(count),
		"-W", strconv.Itoa(timeout),
		"-i", strconv.Itoa(interval),
		in.Target).build()
	return s.executeAndWrapResult(ctx, in.TargetParams, command)
}

func (s *MCPServer) Traceroute(ctx context.Context, req *mcp.CallToolRequest, in types.TracerouteParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := client.ValidateHostname(in.Target); err != nil {
		return nil, types.CommandResult{}, err
	}
	_, err := validateIntWithDefault(in.MaxHops, 30, 1, 30, "max_hops")
	if err != nil {
		return nil, types.CommandResult{}, err
	}

	cmdName := "tracepath"
	if in.IPv6 {
		cmdName = "tracepath6"
	}
	command := newCommand(cmdName, "-n", in.Target).build()
	return s.executeAndWrapResult(ctx, in.TargetParams, command)
}

func (s *MCPServer) Dig(ctx context.Context, req *mcp.CallToolRequest, in types.DigParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if in.Hostname == "" {
		return nil, types.CommandResult{}, fmt.Errorf("hostname is required")
	}
	cmd := newCommand("dig")

	if in.Server != "" {
		if err := client.ValidateIP(in.Server); err != nil {
			return nil, types.CommandResult{}, fmt.Errorf("invalid DNS server: %w", err)
		}
		cmd.add("@" + in.Server)
	}

	recordType := stringWithDefault(in.RecordType, "A")

	cmd.add(in.Hostname, recordType).
		addIf(in.Short, "+short")

	return s.executeAndWrapResult(ctx, in.TargetParams, cmd.build())
}

func (s *MCPServer) Curl(ctx context.Context, req *mcp.CallToolRequest, in types.CurlParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := client.ValidateURL(in.URL); err != nil {
		return nil, types.CommandResult{}, err
	}

	method := stringWithDefault(in.Method, "GET")

	timeout, err := validateIntWithDefault(in.Timeout, 10, 1, 30, "timeout")
	if err != nil {
		return nil, types.CommandResult{}, err
	}

	cmd := newCommand("curl", "-X", method)
	for key, value := range in.Headers {
		cmd.add("-H", fmt.Sprintf("%s: %s", key, value))
	}
	cmd.add("--max-time", strconv.Itoa(timeout))

	if in.FollowRedirect {
		maxRedirects, err := validateIntWithDefault(in.MaxRedirects, 5, 1, 10, "max_redirects")
		if err != nil {
			return nil, types.CommandResult{}, err
		}
		cmd.add("-L", "--max-redirs", strconv.Itoa(maxRedirects))
	}

	cmd.addIf(in.InsecureSSL, "-k").
		add("-i", in.URL)
	return s.executeAndWrapResult(ctx, in.TargetParams, cmd.build())
}

func (s *MCPServer) Netcat(ctx context.Context, req *mcp.CallToolRequest, in types.NetcatParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := client.ValidateHostname(in.Host); err != nil {
		return nil, types.CommandResult{}, err
	}
	if err := client.ValidatePort(in.Port); err != nil {
		return nil, types.CommandResult{}, err
	}

	timeout, err := validateIntWithDefault(in.Timeout, 5, 1, 10, "timeout")
	if err != nil {
		return nil, types.CommandResult{}, err
	}

	cmd := newCommand("nc", "-v", "-z", "-w", strconv.Itoa(timeout)).
		addIf(in.UDP, "-u").
		add(in.Host, strconv.Itoa(in.Port))
	output, _ := s.executeCommand(ctx, in.TargetParams, cmd.build())
	return nil, types.CommandResult{Output: output}, nil
}

// registerConnectivityTools registers connectivity testing tools
func (s *MCPServer) registerConnectivityTools(server *mcp.Server) {
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ping",
			Description: `Test ICMP connectivity to a host.

Examples:
- Basic ping: {"target": "8.8.8.8"}
- Custom count: {"target": "google.com", "count": 5}
- IPv6: {"target": "2001:4860:4860::8888", "ipv6": true}
- With interval: {"target": "192.168.1.1", "count": 10, "interval": 2}`,
		}, s.Ping)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "traceroute",
			Description: `Trace the network path to a host.

Examples:
- Basic traceroute: {"target": "8.8.8.8"}
- Custom max hops: {"target": "google.com", "max_hops": 20}
- IPv6: {"target": "2001:4860:4860::8888", "ipv6": true}`,
		}, s.Traceroute)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "dig",
			Description: `Perform DNS lookup.

Examples:
- Basic lookup: {"hostname": "google.com"}
- Custom server: {"hostname": "example.com", "server": "8.8.8.8"}
- Specific record: {"hostname": "example.com", "record_type": "MX"}
- Short output: {"hostname": "example.com", "short": true}`,
		}, s.Dig)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "curl",
			Description: `Perform HTTP/HTTPS request.

Examples:
- Basic GET: {"url": "https://example.com"}
- POST request: {"url": "https://api.example.com/data", "method": "POST"}
- With headers: {"url": "https://api.example.com", "headers": {"Authorization": "Bearer token"}}
- Follow redirects: {"url": "https://example.com", "follow_redirect": true}`,
		}, s.Curl)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "netcat",
			Description: `Test TCP/UDP port connectivity.

Examples:
- TCP port check: {"host": "example.com", "port": 443}
- UDP port check: {"host": "8.8.8.8", "port": 53, "udp": true}
- With timeout: {"host": "192.168.1.1", "port": 22, "timeout": 3}`,
		}, s.Netcat)
}
