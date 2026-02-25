# ovn-kubernetes-mcp

Repo hosting the Model Context Protocol Server for troubleshooting OVN-Kubernetes.

## Table of Contents

- [Operating Modes](#operating-modes)
- [How to connect to the MCP Server](#how-to-connect-to-the-mcp-server)
  - [Command-line options](#command-line-options)
  - [Live Cluster Mode](#live-cluster-mode)
  - [Offline Mode](#offline-mode)
  - [Dual Mode](#dual-mode)
  - [Local development](#local-development)
- [Tools available in MCP Server](#tools-available-in-mcp-server)
  - [Live Cluster Mode](#live-cluster-mode-1)
  - [Offline Mode](#offline-mode-1)

---

## Operating Modes

| Mode                     | Description |
|--------------------------|-------------|
| `live-cluster` (default) | Connect to a live Kubernetes cluster for real-time debugging (requires kubeconfig). |
| `offline`                | Analyze sosreports and must-gathers without cluster access. |
| `dual`                   | Exposes tools from dual live-cluster and offline modes (requires kubeconfig for live-cluster tools). |

---

## How to connect to the MCP Server

To use the MCP server, ensure you have [Go](https://go.dev/) installed with a version greater than or equal to the version specified in the `go.mod` file.

The server currently supports 2 transport modes: `stdio` and `http`.

### Command-line options

| Option | Default | Description |
|--------|---------|-------------|
| `--mode` | `live-cluster` | Server mode: `live-cluster`, `offline`, or `dual`. |
| `--transport` | `stdio` | Transport: `stdio` or `http`. |
| `--port` | `8080` | Port for HTTP transport. |
| `--kubeconfig` | (none) | Path to kubeconfig file. Required for `live-cluster` and `dual`. |
| `--pwru-image` | `docker.io/cilium/pwru:v1.0.10` | Container image for the **pwru** network tool (kernel packet tracing). |
| `--tcpdump-image` | `nicolaka/netshoot:v0.13` | Container image for the **tcpdump** network tool (packet capture). |
| `--kernel-image` | `nicolaka/netshoot:v0.13` | Container image for kernel tools (conntrack, ip, iptables, nft). |

### Live Cluster Mode

For `stdio` mode, the server can be run and connected to by using the following configuration in an MCP host (Cursor, Claude, etc.):

```json
{
  "mcpServers": {
    "ovn-kubernetes": {
      "command": "go",
      "args": [
        "run",
        "github.com/ovn-kubernetes/ovn-kubernetes-mcp/cmd/ovnk-mcp-server@latest",
        "--kubeconfig",
        "/PATH-TO-THE-KUBECONFIG-FILE"
      ]
    }
  }
}
```

For `http` mode, the server should be started separately:

```shell
go run github.com/ovn-kubernetes/ovn-kubernetes-mcp/cmd/ovnk-mcp-server@latest --transport http --kubeconfig /PATH-TO-THE-KUBECONFIG-FILE
```

The following configuration should be used in an MCP host (Cursor, Claude, etc.) to connect to the server:

```json
{
  "mcpServers": {
    "ovn-kubernetes": {
      "url": "http://localhost:8080"
    }
  }
}
```

### Offline Mode

For offline troubleshooting (e.g., analyzing sosreports, must-gathers), use `--mode offline`:

For `stdio` mode:

```json
{
  "mcpServers": {
    "ovn-kubernetes": {
      "command": "go",
      "args": [
        "run",
        "github.com/ovn-kubernetes/ovn-kubernetes-mcp/cmd/ovnk-mcp-server@latest",
        "--mode",
        "offline"
      ]
    }
  }
}
```

For `http` mode:

```shell
go run github.com/ovn-kubernetes/ovn-kubernetes-mcp/cmd/ovnk-mcp-server@latest --transport http --mode offline
```

> **Note:** Must-gather tools require the [`omc`](https://github.com/gmeghnag/omc) binary. Some must-gather tools also require the [`ovsdb-tool`](https://www.openvswitch.org/support/dist-docs/ovsdb-tool.1.txt) binary; if it is not available, those tools are not registered.

### Dual Mode

For using dual [live-cluster](#live-cluster-mode) (needs kubeconfig) and [offline](#offline-mode) tools, use `--mode dual`:

For `stdio` mode:

```json
{
  "mcpServers": {
    "ovn-kubernetes": {
      "command": "go",
      "args": [
        "run",
        "github.com/ovn-kubernetes/ovn-kubernetes-mcp/cmd/ovnk-mcp-server@latest",
        "--mode",
        "dual",
        "--kubeconfig",
        "/PATH-TO-THE-KUBECONFIG-FILE"
      ]
    }
  }
}
```

For `http` mode:

```shell
go run github.com/ovn-kubernetes/ovn-kubernetes-mcp/cmd/ovnk-mcp-server@latest --transport http --mode dual --kubeconfig /PATH-TO-THE-KUBECONFIG-FILE
```

### Local development

When developing or building locally, run `make build` and use the binary path as the command.

- Use `--mode <live-cluster|offline|dual>` to choose the mode.
- For `live-cluster` or `dual`, add `--kubeconfig /path/to/kubeconfig`.
- For `offline`, omit `--kubeconfig`.

**stdio:**

```json
{
  "mcpServers": {
    "ovn-kubernetes": {
      "command": "/PATH-TO-THE-LOCAL-GIT-REPO/_output/ovnk-mcp-server",
      "args": [
        "--mode",
        "<live-cluster|offline|dual>",
        "--kubeconfig",
        "/PATH-TO-THE-KUBECONFIG-FILE"
      ]
    }
  }
}
```

**http:**

```shell
/PATH-TO-THE-LOCAL-GIT-REPO/_output/ovnk-mcp-server --transport http --mode <live-cluster|offline|dual> --kubeconfig /PATH-TO-THE-KUBECONFIG-FILE
```

For `offline` mode, omit the `--kubeconfig` argument in dual examples above.

---

<!-- TOOLS_SECTION_START -->
## Tools available in MCP Server

### Live Cluster Mode

Available when running with `--mode live-cluster` or `--mode dual` (and with a valid kubeconfig).

| Category      | Tool | Description |
|---------------|------|-------------|
| **kubernetes** | `pod-logs` | Get container logs from a pod in the Kubernetes cluster. |
| | `resource-get` | Get a specific Kubernetes resource by name. |
| | `resource-list` | List Kubernetes resources of a specific kind. |
| **ovn** | `ovn-show` | Display a comprehensive overview of OVN configuration from either the Northbound or Southbound database. |
| | `ovn-get` | Query records from an OVN database table with flexible filtering. |
| | `ovn-lflow-list` | List logical flows from the OVN Southbound database. |
| | `ovn-trace` | Trace a packet through the OVN logical network. |
| **ovs** | `ovs-list-br` | List all OVS bridges on a specific pod. |
| | `ovs-list-ports` | List all ports on a specific OVS bridge. |
| | `ovs-list-ifaces` | List all interfaces on a specific OVS bridge. |
| | `ovs-vsctl-show` | Display a comprehensive overview of OVS configuration. |
| | `ovs-ofctl-dump-flows` | Dump OpenFlow flows from a specific OVS bridge. |
| | `ovs-appctl-dump-conntrack` | Dump connection tracking entries from OVS datapath. |
| | `ovs-appctl-ofproto-trace` | Trace a packet through the OpenFlow pipeline. |
| **kernel** | `get-conntrack` | get-conntrack allows to interact with the connection tracking system of a Kubernetes node. |
| | `get-iptables` | get-iptables allows to interact with kernel to list packet filter rules. |
| | `get-nft` | get-nft allows to interact with kernel to list packet filtering and classification rules. |
| | `get-ip` | get-ip allows to interact with kernel to list routing, network devices, interfaces. |
| **network-tools** | `tcpdump` | Capture network packets on a node or inside a pod with strict safety controls. |
| | `pwru` | Trace packets through the Linux kernel networking stack using eBPF. |

### Offline Mode

Available when running with `--mode offline` or `--mode dual`. No cluster access required.

| Category       | Tool | Description |
|----------------|------|-------------|
| **sosreport** | `sos-list-plugins` | List enabled sosreport plugins with their command counts. |
| | `sos-list-commands` | List all commands collected by a specific sosreport plugin. |
| | `sos-search-commands` | Search for commands across all plugins by pattern. |
| | `sos-get-command` | Get command output using filepath from manifest. |
| | `sos-search-pod-logs` | Search Kubernetes pod container logs. |
| **must-gather** | `must-gather-get-resource` | Get a specific Kubernetes resource from a must-gather archive. |
| | `must-gather-list-resources` | List Kubernetes resources of a specific kind from a must-gather archive. |
| | `must-gather-pod-logs` | Get container logs from a pod in a must-gather archive. |
| | `must-gather-ovnk-info` | Get OVN-Kubernetes networking information from a must-gather archive. |
| | `must-gather-list-northbound-databases` | List OVN Northbound database files available in a must-gather archive. |
| | `must-gather-list-southbound-databases` | List OVN Southbound database files available in a must-gather archive. |
| | `must-gather-query-database` | Query an OVN database from a must-gather archive using ovsdb-tool. |

<!-- TOOLS_SECTION_END -->
