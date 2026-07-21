# User Guide

This guide documents every tool exposed by the **ovn-kubernetes-mcp** server. Use it alongside the [README](../README.md) for installation, operating modes, and connection setup.

## Operating modes and tool availability

| Mode | Available categories |
|------|----------------------|
| `live-cluster` (default) | [kubernetes](kubernetes.md), [ovn](ovn.md), [ovs](ovs.md), [kernel](kernel.md), [network-tools](network-tools.md) |
| `offline` | [sosreport](sosreport.md), [must-gather](must-gather.md) |
| `dual` | All categories above (live-cluster tools still need a kubeconfig or in-cluster credentials) |

You can hide categories or individual tools with `--disable-categories` and `--disable-tools`. See [Selectively exposing tools](../README.md#selectively-exposing-tools) in the README.

## Common parameters

Some tools reuse the same optional fields for output limiting, filtering, or per-call timeouts. **Support is not universal** — see the matrix below. Category pages list only tool-specific parameters and link here for shared fields.

| Category | `head` / `tail` / `apply_tail_first` | `pattern` | `timeout_seconds` | Notes |
|----------|--------------------------------------|-----------|-------------------|-------|
| [kubernetes](kubernetes.md) | `pod-logs` only | `pod-logs` only | — | `resource-get` / `resource-list` use neither |
| [ovn](ovn.md) | all tools | `ovn-get`, `ovn-lflow-list`, `ovn-trace` | — | `ovn-show` has head/tail only |
| [ovs](ovs.md) | most tools (`ovs-vsctl` only when `action=show`) | `ovs-ofctl`, `ovs-appctl` | — | |
| [kernel](kernel.md) | all tools | — | all tools | |
| [network-tools](network-tools.md) | — | — | all tools | Packet matching uses `bpf_filter`, not `pattern` |
| [sosreport](sosreport.md) | `sos-get-command`, `sos-get-pod-logs` | `sos-get-command`, `sos-get-pod-logs`; `sos-search-commands` for exec/filepath search | — | `sos-search-commands` also has `max_results` |
| [must-gather](must-gather.md) | `must-gather-pod-logs` only | `must-gather-pod-logs` only | — | Other must-gather tools use neither |

### Head / tail line limiting

Used by tools marked in the matrix above.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `head` | integer | `100` lines if `tail` is not specified | Return only first N lines |
| `tail` | integer | — | Return only last N lines |
| `apply_tail_first` | boolean | `false` | If both head and tail are set and apply_tail_first is true, apply tail before head |

### Pattern filtering

Used by tools marked in the matrix above (grep-style line filtering unless a tool defines different semantics).

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `pattern` | string | — | Regex pattern to filter log/output lines |

### Per-call timeout

Used by [kernel](kernel.md) and [network-tools](network-tools.md) only. Other categories rely on the server `--tool-timeout` without a per-call override field.

The server default is **120 seconds** via `--tool-timeout` (`0` disables the global timeout); per-call `timeout_seconds` overrides that value and is capped at **300** seconds.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `timeout_seconds` | integer | server `--tool-timeout` (**120**) | Timeout in seconds for the command execution. If not specified, server default timeout is used. The maximum value is 300 seconds |

## Tool categories

### Live Cluster Mode

Available with `--mode live-cluster` or `--mode dual`.

| Category | Description |
|----------|-------------|
| [kubernetes](kubernetes.md) | Pod logs and Kubernetes resource get/list |
| [ovn](ovn.md) | OVN Northbound/Southbound introspection and packet tracing |
| [ovs](ovs.md) | Open vSwitch configuration, OpenFlow flows, and datapath debugging |
| [kernel](kernel.md) | Node-level conntrack, iptables, nftables, and `ip` commands |
| [network-tools](network-tools.md) | Packet capture (`tcpdump`) and kernel stack tracing (`pwru`) |

### Offline Mode

Available with `--mode offline` or `--mode dual`. No cluster access required.

| Category | Description |
|----------|-------------|
| [sosreport](sosreport.md) | Browse and search extracted sosreport archives |
| [must-gather](must-gather.md) | Query must-gather archives (Kubernetes resources, logs, OVN databases) |

## Prerequisites by category

| Category | Requirements |
|----------|--------------|
| Live-cluster tools | Valid kubeconfig, or in-cluster ServiceAccount credentials |
| [must-gather](must-gather.md) | [`omc`](https://github.com/gmeghnag/omc) on `PATH`. Database list/query tools also need [`ovsdb-tool`](https://www.openvswitch.org/support/dist-docs/ovsdb-tool.1.txt) |
| [sosreport](sosreport.md) | Extracted sosreport directory containing `sos_commands/` and `sos_reports/manifest.json`. `sos-get-pod-logs` also needs the `container_log` plugin (`<pod>_<namespace>_<container>-<id>.log`) |
