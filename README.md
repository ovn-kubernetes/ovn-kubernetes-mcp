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
  - [Kubernetes deployment](#kubernetes-deployment)
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

| Option | Default                         | Description |
|--------|---------------------------------|-------------|
| `--mode` | `live-cluster`                  | Server mode: `live-cluster`, `offline`, or `dual`. |
| `--transport` | `stdio`                         | Transport: `stdio` or `http`. |
| `--host` | `localhost`                     | Address the HTTP server binds to (`http` only). Use `0.0.0.0` in a container so clients can reach the listener. |
| `--port` | `8080`                          | Port for HTTP transport. |
| `--kubeconfig` | (none)                          | Path to kubeconfig file. Omit when using in-cluster **ServiceAccount** credentials (for example the pod deployment); otherwise set for `live-cluster` and `dual`. |
| `--pwru-image` | `docker.io/cilium/pwru:v1.0.10` | Container image for the **pwru** network tool (kernel packet tracing). |
| `--tcpdump-image` | `nicolaka/netshoot:v0.15`       | Container image for the **tcpdump** network tool (packet capture). |
| `--kernel-image` | `nicolaka/netshoot:v0.15`       | Container image for kernel tools (conntrack, ip, iptables, nft). |
| `--tool-timeout` | `120`                           | Timeout in seconds for tool operations. Set to `0` to disable. |

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

### Kubernetes deployment

The manifests under [`config/`](config/) run the server **in the cluster** as HTTP ([`config/deployment.yaml`](config/deployment.yaml)): `--transport=http`, `--host=0.0.0.0`, `--port=8080`, and `--mode=live-cluster`. There is **no** `--kubeconfig` argument in that deployment; the pod uses its **ServiceAccount** credentials to connect to the Kubernetes API server.

Apply the manifests with:

```shell
make deploy-ovnk-mcp-k8s IMAGE=<your-registry>/ovnk-mcp-server:<tag>
```

`IMAGE` is optional if the image already matches what the manifests expect. Remove the stack with `make undeploy-ovnk-mcp-k8s`.

**Service:** a `ClusterIP` Service `ovnk-mcp-server` in namespace `ovn-kubernetes-mcp` exposes port **8080** to the pod ([`config/service.yaml`](config/service.yaml)).

**NetworkPolicy:** [`config/networkpolicy.yaml`](config/networkpolicy.yaml) selects the MCP server pod and sets **`ingress: []`**, which denies **all** ingress to that pod from the cluster network. So other workloads cannot reach the MCP HTTP endpoint through the Service, and there is no in-manifest Ingress or LoadBalancer for public access.

**Connecting an LLM client (Cursor, Claude, and so on) from your machine:** the usual pattern is to use **`kubectl port-forward`** from a kube context that can access the cluster. That forwards a local port to the Service without relying on in-cluster ingress to the pod (the forward is set up by the kubelet). In one terminal:

```shell
kubectl port-forward -n ovn-kubernetes-mcp svc/ovnk-mcp-server 8080:8080
```

Then configure the host’s MCP client for **streamable HTTP** against the forwarded port (same shape as [HTTP live-cluster](#live-cluster-mode) above):

```json
{
  "mcpServers": {
    "ovn-kubernetes": {
      "url": "http://127.0.0.1:8080"
    }
  }
}
```

Keep the port-forward process running while you use the tools. The server URL is the **base** of the HTTP transport; clients that implement [MCP Streamable HTTP](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports) negotiate on top of that.

**Security note:** the shipped manifests do not add TLS or application-level auth on the MCP HTTP listener. Treat network access as sensitive: use port-forward or private networking, and rely on Kubernetes RBAC (who may port-forward or change `NetworkPolicy`) to limit who can reach the server.

**Allowing only a trusted in-cluster pod:** leave [`config/networkpolicy.yaml`](config/networkpolicy.yaml) in place. For a pod that is allowed to talk to the MCP server (for example a single well-known automation or gateway workload), add a **second** `NetworkPolicy` in namespace `ovn-kubernetes-mcp`. Policies that select the same pod are [additive](https://kubernetes.io/docs/concepts/services-networking/network-policies/): allowed ingress is the **union** of every matching policy’s `ingress` rules, so the deny-all policy keeps every other source blocked while your new policy explicitly permits the trusted peer on port **8080** only.

1. Give the trusted client pod a **narrow, unique** label (example below uses `app.kubernetes.io/name: ovnk-mcp-trusted-client`).
2. Apply a policy that uses the **same** `podSelector` as the shipped policy (MCP server labels: `app.kubernetes.io/name: ovn-kubernetes-mcp` and `app.kubernetes.io/component: mcp-server`) and an `ingress` rule whose `from` matches only that client, with `ports` limited to `8080` / `TCP`.

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: ovnk-mcp-server-allow-trusted-client
  namespace: ovn-kubernetes-mcp
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: ovn-kubernetes-mcp
      app.kubernetes.io/component: mcp-server
  policyTypes:
    - Ingress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app.kubernetes.io/name: ovnk-mcp-trusted-client
      ports:
        - port: 8080
          protocol: TCP
```

From that client pod, call the Service at `http://ovnk-mcp-server.ovn-kubernetes-mcp.svc:8080` (or the fully qualified `*.svc.cluster.local` name). If the client runs in **another** namespace, use a `from` entry with both `namespaceSelector` and `podSelector` as described under [Targeting multiple namespaces by label](https://kubernetes.io/docs/concepts/services-networking/network-policies/#targeting-multiple-namespaces-by-label). Prefer explicit labels over wide selectors, and keep trusting this path only for workloads you control—there is still no MCP-level authentication on the wire.

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
