# OVN tools

Live-cluster tools that exec into an OVN / ovnkube pod and query the OVN Northbound or Southbound databases. Available with `--mode live-cluster` or `--mode dual`.

See also: [User guide](user-guide.md) (including [common parameters](user-guide.md#common-parameters))

| Tool | Description |
|------|-------------|
| [`ovn-show`](#ovn-show) | Display a comprehensive overview of OVN configuration from either the Northbound or Southbound database |
| [`ovn-get`](#ovn-get) | Query records from an OVN database table with flexible filtering |
| [`ovn-lflow-list`](#ovn-lflow-list) | List logical flows from the OVN Southbound database |
| [`ovn-trace`](#ovn-trace) | Trace a packet through the OVN logical network |

---

## ovn-show

For Northbound (`nbdb`): Runs `ovn-nbctl show` and displays logical switches, logical routers, their ports, and connections between them.

For Southbound (`sbdb`): Runs `ovn-sbctl show` and displays chassis information, port bindings, and their relationships.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `namespace` | string | no | — | Kubernetes namespace of the OVN pod (e.g., `"ovn-kubernetes"`) |
| `name` | string | **yes** | — | Name of the pod running OVN (e.g., `"ovnkube-node-xxxxx"`) |
| `database` | string | **yes** | — | OVN database to query - `"nbdb"` for Northbound or `"sbdb"` for Southbound |

Also accepts common [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting).

### Examples

```json
{
  "name": "ovnkube-node-xxxxx",
  "namespace": "ovn-kubernetes",
  "database": "nbdb"
}
```

```json
{
  "name": "ovnkube-node-xxxxx",
  "namespace": "ovn-kubernetes",
  "database": "sbdb",
  "tail": 50
}
```

---

## ovn-get

This is a versatile command that can:

1. List all records in a table (when no record specified)
2. Get a specific record (when record specified)

Common Northbound tables: `Logical_Switch`, `Logical_Router`, `Logical_Switch_Port`, `Logical_Router_Port`, `ACL`, `Address_Set`, `Port_Group`, `Load_Balancer`, `NAT`

Common Southbound tables: `Chassis`, `Port_Binding`, `Datapath_Binding`, `Logical_Flow`, `MAC_Binding`, `Multicast_Group`, `SB_Global`

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `namespace` | string | no | — | Kubernetes namespace of the OVN pod |
| `name` | string | **yes** | — | Name of the pod running OVN |
| `database` | string | **yes** | — | OVN database to query - `"nbdb"` for Northbound or `"sbdb"` for Southbound |
| `table` | string | **yes** | — | Name of the table (e.g., `"Logical_Switch"`, `"Port_Binding"`) |
| `record` | string | no | lists all records | Record identifier (UUID or name). If not specified, lists all records |
| `columns` | string | no | — | Comma-separated list of columns to display (e.g., `"name,_uuid,ports"`) |

Also accepts common [`pattern`](user-guide.md#pattern-filtering) (only when listing all records) and [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting).

### Examples

```json
{
  "name": "ovnkube-node-xxxxx",
  "namespace": "ovn-kubernetes",
  "database": "nbdb",
  "table": "Port_Group"
}
```

```json
{
  "name": "ovnkube-node-xxxxx",
  "namespace": "ovn-kubernetes",
  "database": "nbdb",
  "table": "Logical_Router",
  "record": "ovn_cluster_router"
}
```

```json
{
  "name": "ovnkube-node-xxxxx",
  "namespace": "ovn-kubernetes",
  "database": "nbdb",
  "table": "Logical_Switch",
  "columns": "name,ports"
}
```

---

## ovn-lflow-list

Runs `ovn-sbctl lflow-list` to retrieve logical flows which represent the compiled logical network pipeline. This is essential for debugging packet forwarding.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `namespace` | string | no | — | Kubernetes namespace of the OVN pod |
| `name` | string | **yes** | — | Name of the pod running OVN |
| `datapath` | string | no | — | Datapath name or UUID to filter flows for a specific logical switch/router |

Also accepts common [`pattern`](user-guide.md#pattern-filtering) and [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting).

### Examples

```json
{
  "name": "ovnkube-node-xxxxx",
  "namespace": "ovn-kubernetes"
}
```

```json
{
  "name": "ovnkube-node-xxxxx",
  "namespace": "ovn-kubernetes",
  "datapath": "node1",
  "pattern": "inport.*pod1",
  "head": 200
}
```

---

## ovn-trace

Runs `ovn-trace` to simulate packet processing through the logical network pipeline. This shows which logical flows match, what actions are taken, and the final disposition.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `namespace` | string | no | — | Kubernetes namespace of the OVN pod |
| `name` | string | **yes** | — | Name of the pod running OVN |
| `datapath` | string | **yes** | — | Name of the logical switch or router to start the trace |
| `microflow` | string | **yes** | — | Microflow specification describing the packet (e.g., `"inport==\"pod1\" && eth.src==00:00:00:00:00:01 && ip4.src==10.244.0.5 && ip4.dst==10.244.1.5"`) |
| `mode` | string | no | `"detailed"` | Output verbosity mode - `"detailed"` (default), `"summary"`, or `"minimal"` |

Also accepts common [`pattern`](user-guide.md#pattern-filtering) and [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting).

### Examples

```json
{
  "name": "ovnkube-node-xxxxx",
  "namespace": "ovn-kubernetes",
  "datapath": "node1",
  "microflow": "inport==\"pod1\" && eth.src==00:00:00:00:00:01 && ip4.src==10.244.0.5 && ip4.dst==10.244.1.5"
}
```

```json
{
  "name": "ovnkube-node-xxxxx",
  "namespace": "ovn-kubernetes",
  "datapath": "node1",
  "microflow": "inport==\"pod1\" && eth.src==00:00:00:00:00:01 && icmp && ip4.src==10.244.0.5 && ip4.dst==8.8.8.8",
  "mode": "summary"
}
```
