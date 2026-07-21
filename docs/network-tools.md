# Network tools

Live-cluster packet capture and kernel stack tracing. Available with `--mode live-cluster` or `--mode dual`.

See also: [User guide](user-guide.md) (including [common parameters](user-guide.md#common-parameters))

| Tool | Description |
|------|-------------|
| [`tcpdump`](#tcpdump) | Capture network packets on a node or inside a pod |
| [`pwru`](#pwru) | Trace packets through the Linux kernel networking stack using eBPF |

---

## tcpdump

Captures packets with BPF filtering under fixed caps so runs stay bounded.

Safety / targeting limits:

- Max packets: `packet_count` default **100**, hard max **1000**
- Max snaplen: default **96** bytes, hard max **1500** bytes
- Timeout: common [`timeout_seconds`](user-guide.md#per-call-timeout) (server default **120s**, per-call max **300s**)
- Target rules: `target_type` must be `node` or `pod`; `name` is required; `namespace` is required for `pod` (defaults to `default` for the node debug pod); `container_name` is optional for `pod`

Node captures run in a debug pod on the named node; pod captures exec into the existing pod.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `target_type` | string | **yes** | — | `"node"` or `"pod"` |
| `name` | string | **yes** | — | Name of the target (node or pod) |
| `namespace` | string | required when `target_type` is `"pod"`; optional when `target_type` is `"node"` | `"default"` when `target_type` is `"node"` | Namespace of the target (node or pod) |
| `container_name` | string | no | — (default container) | Name of the container in the pod when `target_type` is `"pod"` |
| `interface` | string | no | — (tool default) | Network interface name or `"any"` |
| `packet_count` | integer | no | `100` (max: `1000`) | Number of packets to capture |
| `bpf_filter` | string | no | — | BPF filter expression to match packets (e.g., `"tcp and dst port 8080"`, `"host 10.0.0.1"`) |
| `snaplen` | integer | no | `96` (max: `1500`) | Snapshot length in bytes |

Also accepts common [`timeout_seconds`](user-guide.md#per-call-timeout).

### Examples

```json
{
  "target_type": "node",
  "name": "worker-1",
  "interface": "eth0",
  "packet_count": 100,
  "bpf_filter": "tcp port 80"
}
```

```json
{
  "target_type": "pod",
  "name": "my-pod",
  "namespace": "default",
  "interface": "eth0",
  "packet_count": 100,
  "bpf_filter": "host 10.0.0.1"
}
```

```json
{
  "target_type": "node",
  "name": "worker-1",
  "interface": "any",
  "packet_count": 50,
  "bpf_filter": "port 53"
}
```

---

## pwru

pwru (packet, where are you?) shows which kernel functions process a packet, helping debug packet drops, routing issues, and understanding the kernel's packet processing path.

This tool creates a specialized debug pod on the specified node with necessary eBPF capabilities to trace packets through kernel networking functions.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `node_name` | string | **yes** | — | Name of the node to run pwru on |
| `node_pod_namespace` | string | no | `"default"` | Namespace of the debug pod on which the command is expected to be executed |
| `bpf_filter` | string | no | — | BPF filter expression to match packets (e.g., `"tcp and dst port 8080"`, `"host 10.0.0.1"`) |
| `output_limit_lines` | integer | no | `100` (max: `1000`) | Maximum number of trace events to capture |

Also accepts common [`timeout_seconds`](user-guide.md#per-call-timeout).

### Examples

```json
{
  "node_name": "worker-1",
  "bpf_filter": "host 10.244.0.5",
  "output_limit_lines": 100
}
```

```json
{
  "node_name": "worker-1",
  "bpf_filter": "tcp and dst port 8080",
  "output_limit_lines": 50
}
```

```json
{
  "node_name": "worker-1",
  "bpf_filter": "icmp",
  "output_limit_lines": 100
}
```
