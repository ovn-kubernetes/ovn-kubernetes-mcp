# Changelog

## v0.1.0 — Initial Release

The first release of the OVN-Kubernetes MCP Server — a Model Context Protocol server
for troubleshooting OVN-Kubernetes clusters with AI assistants.

### Highlights

- **Three operating modes**: `live-cluster` for real-time cluster debugging,
  `offline` for analyzing sosreports and must-gather archives, and `dual` for both.
  See the [User Guide](docs/user-guide.md) for full details.

- **Two transport modes**: `stdio` for direct IDE integration (Cursor, Claude, etc.)
  and `http` (Streamable HTTP) for remote/containerized deployments.
  See [connection setup](README.md#how-to-connect-to-the-mcp-server).

- **Live cluster tools**:
  - **[Kubernetes](docs/kubernetes.md)**: pod logs, resource get/list with jsonpath support.
  - **[OVN](docs/ovn.md)**: show, get, lflow-list, trace across Northbound/Southbound databases.
  - **[OVS](docs/ovs.md)**: bridge/port/interface listing, OpenFlow dump, conntrack dump, ofproto-trace.
  - **[Kernel](docs/kernel.md)**: conntrack, iptables, nft, ip route/device inspection via debug pods.
  - **[Network tools](docs/network-tools.md)**: tcpdump packet capture and pwru eBPF kernel packet tracing.

- **Offline analysis tools**:
  - **[sosreport](docs/sosreport.md)**: plugin listing, command search, pod log search.
  - **[must-gather](docs/must-gather.md)**: resource get/list, pod logs, OVN-K info, OVN database queries
    via ovsdb-tool.

- **[Kubernetes deployment](README.md#kubernetes-deployment)**: Dockerfile, kustomize manifests, RBAC, NetworkPolicy,
  and Service for running the server in-cluster with ServiceAccount credentials.

- **Tool timeout enforcement**: configurable per-tool timeout (default 120s) applied
  at the protocol middleware layer.

- **[Selective tool exposure](README.md#selectively-exposing-tools)**: `--disable-categories` and `--disable-tools` flags to
  hide tools/categories from clients without code changes.

- **E2E test suite**: Ginkgo-based end-to-end tests covering live-cluster, offline,
  and dual modes with CI integration.
