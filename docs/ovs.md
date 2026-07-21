# OVS tools

Live-cluster tools that exec into an ovnkube-node pod to inspect Open vSwitch. Available with `--mode live-cluster` or `--mode dual`.

See also: [User guide](user-guide.md) (including [common parameters](user-guide.md#common-parameters))

| Tool | Description |
|------|-------------|
| [`ovs-vsctl`](#ovs-vsctl) | Run an ovs-vsctl command against an ovnkube-node pod to inspect the OVS switch configuration |
| [`ovs-ofctl`](#ovs-ofctl) | Run an ovs-ofctl command against an ovnkube-node pod to inspect the OpenFlow state of an OVS bridge |
| [`ovs-appctl`](#ovs-appctl) | Run an ovs-appctl command against an ovnkube-node pod to interact with the OVS daemons for datapath and OpenFlow debugging |

---

## ovs-vsctl

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `namespace` | string | **yes** | — | Kubernetes namespace of the OVS pod |
| `name` | string | **yes** | — | Name of the pod running OVS |
| `action` | string | **yes** | — | The ovs-vsctl subcommand to run. `show`: Display a comprehensive overview of OVS configuration in a hierarchical format (bridges, ports, interfaces, controllers). `list-br`: List all OVS bridges on the pod. `list-ports`: List all ports on a specific OVS bridge (requires bridge). `list-ifaces`: List all interfaces on a specific OVS bridge (requires bridge) |
| `bridge` | string | required for `list-ports` and `list-ifaces` | — | Name of the OVS bridge (e.g., `"br-int"`) |

Also accepts common [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting) when `action` is `show`.

### Examples

```json
{"namespace": "ovn-kubernetes", "name": "ovnkube-node-xxxxx", "action": "show"}
```

```json
{"namespace": "ovn-kubernetes", "name": "ovnkube-node-xxxxx", "action": "list-br"}
```

```json
{
  "namespace": "ovn-kubernetes",
  "name": "ovnkube-node-xxxxx",
  "action": "list-ports",
  "bridge": "br-int"
}
```

```json
{
  "namespace": "ovn-kubernetes",
  "name": "ovnkube-node-xxxxx",
  "action": "list-ifaces",
  "bridge": "br-int"
}
```

---

## ovs-ofctl

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `namespace` | string | **yes** | — | Kubernetes namespace of the OVS pod |
| `name` | string | **yes** | — | Name of the pod running OVS |
| `action` | string | **yes** | — | The ovs-ofctl subcommand to run. `dump-flows`: Dump the OpenFlow flow entries programmed on the specified bridge |
| `bridge` | string | **yes** | — | Name of the OVS bridge (e.g., `"br-int"`) |

Also accepts common [`pattern`](user-guide.md#pattern-filtering) and [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting).

### Examples

```json
{
  "namespace": "ovn-kubernetes",
  "name": "ovnkube-node-xxxxx",
  "action": "dump-flows",
  "bridge": "br-int"
}
```

```json
{
  "namespace": "ovn-kubernetes",
  "name": "ovnkube-node-xxxxx",
  "action": "dump-flows",
  "bridge": "br-int",
  "pattern": "table=0",
  "head": 50
}
```

---

## ovs-appctl

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `namespace` | string | **yes** | — | Kubernetes namespace of the OVS pod |
| `name` | string | **yes** | — | Name of the pod running OVS |
| `action` | string | **yes** | — | The ovs-appctl subcommand to run. `dpctl/dump-conntrack`: Dump connection tracking entries from the OVS datapath. `ofproto/trace`: Simulate packet processing through the OpenFlow pipeline (requires bridge and flow) |
| `bridge` | string | required for `ofproto/trace` | — | Name of the OVS bridge (e.g., `"br-int"`) |
| `flow` | string | required for `ofproto/trace` | — | Flow specification describing the packet to trace (e.g., `"in_port=1,ip,nw_src=10.244.0.5,nw_dst=10.96.0.1"`) |
| `additional_params` | string[] | no (only used when action is `"dpctl/dump-conntrack"`) | — | Additional CLI arguments to pass to dpctl/dump-conntrack (e.g., `["zone=5"]`) |

Also accepts common [`pattern`](user-guide.md#pattern-filtering) and [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting).

### Examples

```json
{
  "namespace": "ovn-kubernetes",
  "name": "ovnkube-node-xxxxx",
  "action": "dpctl/dump-conntrack"
}
```

```json
{
  "namespace": "ovn-kubernetes",
  "name": "ovnkube-node-xxxxx",
  "action": "dpctl/dump-conntrack",
  "additional_params": ["zone=5"]
}
```

```json
{
  "namespace": "ovn-kubernetes",
  "name": "ovnkube-node-xxxxx",
  "action": "ofproto/trace",
  "bridge": "br-int",
  "flow": "in_port=1,ip,nw_src=10.244.0.5,nw_dst=10.96.0.1"
}
```
