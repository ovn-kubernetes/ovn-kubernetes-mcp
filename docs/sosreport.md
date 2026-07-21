# Sosreport tools

Offline tools for browsing extracted sosreport archives. Available with `--mode offline` or `--mode dual`.

`sosreport_path` must point at an **extracted** directory on disk (not a `.tar.xz` / archive file); it needs `sos_commands/` and `sos_reports/manifest.json`. `sos-get-pod-logs` additionally requires the sosreport `container_log` plugin with copied logs named `<pod>_<namespace>_<container>-<id>.log`.

See also: [User guide](user-guide.md) (including [common parameters](user-guide.md#common-parameters))

| Tool | Description |
|------|-------------|
| [`sos-list-plugins`](#sos-list-plugins) | List enabled sosreport plugins with their command counts |
| [`sos-list-commands`](#sos-list-commands) | List all commands collected by a specific sosreport plugin |
| [`sos-search-commands`](#sos-search-commands) | Search for commands across all plugins by pattern |
| [`sos-get-command`](#sos-get-command) | Get command output using filepath from manifest |
| [`sos-get-pod-logs`](#sos-get-pod-logs) | Get Kubernetes pod container logs from a sosreport |

---

## sos-list-plugins

Returns a list of plugins with the number of commands collected by each plugin.

Use this to discover which plugins are available, then use `sos-list-commands` to see what commands are available within a specific plugin.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `sosreport_path` | string | **yes** | — | Path to extracted sosreport directory |

### Examples

```json
{"sosreport_path": "/path/to/sosreport-hostname-20240101"}
```

---

## sos-list-commands

Returns all commands executed by the plugin with their filepaths. Use the filepath with `sos-get-command` to retrieve the actual command output.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `sosreport_path` | string | **yes** | — | Path to extracted sosreport directory |
| `plugin` | string | **yes** | — | Plugin name (e.g. `"openvswitch"`, `"networking"`, `"kubernetes"`) |

### Examples

```json
{
  "sosreport_path": "/path/to/sosreport-hostname-20240101",
  "plugin": "openvswitch"
}
```

```json
{
  "sosreport_path": "/path/to/sosreport-hostname-20240101",
  "plugin": "networking"
}
```

---

## sos-search-commands

Searches command names and filepaths across all plugins. Returns matching commands with their plugin, exec string, and filepath. Does NOT return file contents.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `sosreport_path` | string | **yes** | — | Path to extracted sosreport directory |
| `pattern` | string | **yes** | — | Regex pattern to search in command exec and filepath |
| `max_results` | integer | no | `100` | Maximum results to return |

### Examples

```json
{
  "sosreport_path": "/path/to/sosreport-hostname-20240101",
  "pattern": "iptables"
}
```

```json
{
  "sosreport_path": "/path/to/sosreport-hostname-20240101",
  "pattern": "ovn.*show"
}
```

```json
{
  "sosreport_path": "/path/to/sosreport-hostname-20240101",
  "pattern": "journalctl.*kubelet"
}
```

---

## sos-get-command

Use the filepath returned by `sos-list-commands` or `sos-search-commands` to retrieve the actual command output. Supports optional grep-style filtering.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `sosreport_path` | string | **yes** | — | Path to extracted sosreport directory |
| `filepath` | string | **yes** | — | Relative filepath from `sos-list-commands` or `sos-search-commands` |

Also accepts common [`pattern`](user-guide.md#pattern-filtering) and [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting).

### Examples

```json
{
  "sosreport_path": "/path/to/sosreport-hostname-20240101",
  "filepath": "sos_commands/openvswitch/ovs-vsctl_-t_5_show"
}
```

```json
{
  "sosreport_path": "/path/to/sosreport-hostname-20240101",
  "filepath": "sos_commands/firewall_tables/iptables_-vnxL",
  "pattern": "KUBE-"
}
```

```json
{
  "sosreport_path": "/path/to/sosreport-hostname-20240101",
  "filepath": "sos_commands/openvswitch/ovs-vsctl_-t_5_show",
  "tail": 50
}
```

---

## sos-get-pod-logs

Reads container logs collected from the sosreport for the specified pod. Returns log lines. If `container` is omitted, returns the first matching container log for the pod.

Requires the sosreport’s `container_log` plugin (missing or empty yields “No pod logs found in sosreport”) with copied logs under the usual `<pod>_<namespace>_<container>-<id>.log` layout.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `sosreport_path` | string | **yes** | — | Path to extracted sosreport directory |
| `namespace` | string | **yes** | — | Pod namespace |
| `name` | string | **yes** | — | Pod name |
| `container` | string | no | — | Container name. If omitted, returns the first matching container log for the pod |

Also accepts common [`pattern`](user-guide.md#pattern-filtering) and [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting).

### Examples

```json
{
  "sosreport_path": "/path/to/sosreport-hostname-20240101",
  "namespace": "openshift-ovn-kubernetes",
  "name": "ovnkube-node-abc"
}
```

```json
{
  "sosreport_path": "/path/to/sosreport-hostname-20240101",
  "namespace": "openshift-ovn-kubernetes",
  "name": "ovnkube-node-abc",
  "container": "ovnkube-controller"
}
```

```json
{
  "sosreport_path": "/path/to/sosreport-hostname-20240101",
  "namespace": "openshift-ovn-kubernetes",
  "name": "ovnkube-node-abc",
  "pattern": "ERROR.*"
}
```
