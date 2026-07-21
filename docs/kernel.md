# Kernel tools

Live-cluster tools that run on a Kubernetes node via an ephemeral debug pod. Available with `--mode live-cluster` or `--mode dual`.

See also: [User guide](user-guide.md) (including [common parameters](user-guide.md#common-parameters))

The debug pod uses the image from `--kernel-image` (default `nicolaka/netshoot:v0.15`) and mounts the host filesystem.

| Tool | Description |
|------|-------------|
| [`get-conntrack`](#get-conntrack) | Interact with the connection tracking system of a Kubernetes node |
| [`get-iptables`](#get-iptables) | List packet filter rules (iptables / ip6tables) |
| [`get-nft`](#get-nft) | List packet filtering and classification rules (nftables) |
| [`get-ip`](#get-ip) | List routing, network devices, and interfaces (`ip`) |

---

## get-conntrack

Use this command to discover a list of all (or a filtered selection of) currently tracked connections.

The tool prefers the `conntrack` CLI in the configured `--kernel-image`. If that binary is missing, it falls back to reading `/proc/net/nf_conntrack`, which only supports dump/list (`-L`/`--dump`) with limited filtering. Count (`-C`/`--count`) and stats (`-S`/`--stats`) require the `conntrack` CLI.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `node` | string | **yes** | — | Name of the node from where conntrack entries are expected to be extracted |
| `namespace` | string | no | `"default"` | Namespace of the debug pod from where conntrack entries are expected to be extracted |
| `command` | string | no | `"-L"` | These options specify the particular operation to perform. These options can only be used if configured image has `conntrack` utility available. If omitted or empty, defaults to `-L`. `-L`/`--dump`: List connection tracking table. `-C`/`--count`: Show the table counter. `-S`/`--stats`: Show the in-kernel connection tracking system statistics |
| `filter_parameters` | string | no | — | These parameters are useful to filter certain entries from the whole table: `-s`/`--src`/`--orig-src IP_ADDRESS`: Match only entries whose source address in the original direction equals to mentioned IP. `-d`/`--dst`/`--orig-dst IP_ADDRESS`: Match only entries whose destination address in the original direction equals to mentioned IP. `-p`/`--proto PROTO`: Specify layer four (TCP, UDP, ...) protocol. `--sport`/`--orig-port-src PORT`: Source port in original direction. `--dport`/`--orig-port-dst PORT`: Destination port in original direction |

Also accepts common [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting) and [`timeout_seconds`](user-guide.md#per-call-timeout).

### Examples

```json
{"node": "ovn-control-plane", "command": "-L"}
```

```json
{"node": "ovn-worker", "namespace": "ovn-kubernetes", "command": "-L"}
```

```json
{
  "node": "ovn-worker",
  "filter_parameters": "-s 1.2.3.4 -d 5.6.7.8 -p tcp --sport 32000 --dport 10250"
}
```

---

## get-iptables

Iptables and ip6tables are used to inspect the tables of IPv4 and IPv6 packet filter rules in the Linux kernel.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `node` | string | **yes** | — | Name of the node from where packet filter rules are expected to be extracted |
| `namespace` | string | no | `"default"` | Namespace of the debug pod from where packet filter rules are expected to be extracted |
| `table` | string | no | `"filter"` | There are currently five independent tables (which tables are present at any time depends on the kernel configuration options and which modules are present). `filter`: This is the default table. `nat`: This table is consulted when a packet that creates a new connection is encountered. `mangle`: This table is used for specialized packet alteration. `raw`: This table is used mainly for configuring exemptions from connection tracking in combination with the NOTRACK target. `security`: This table is used for Mandatory Access Control (MAC) networking rules |
| `command` | string | no | `"-L"` | These options specify the desired action to perform. Only one of them can be specified on the command line unless otherwise stated below. If omitted or empty, defaults to `-L`. `-L`/`--list [chain]`: List all rules in the selected chain. If no chain is selected, all chains are listed. `-S`/`--list-rules [chain]`: Print all rules in the selected chain. If no chain is selected, all chains are printed like iptables-save |
| `filter_parameters` | string | no | — | These parameters are useful to filter certain entries from the whole table: `-s`/`--source address[/mask]`: Source specification. Address can be either a network name, a hostname, a network IP address (with /mask), or a plain IP address. `-d`/`--destination address[/mask]`: Destination specification. `-v`/`--verbose`: Verbose output. `-n`/`--numeric`: Numeric output. IP addresses and port numbers will be printed in numeric format. `-p`/`--protocol protocol`: The protocol of the rule or of the packet to check. `-4`/`--ipv4`: IPv4. `-6`/`--ipv6`: IPv6 |

Also accepts common [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting) and [`timeout_seconds`](user-guide.md#per-call-timeout).

### Examples

```json
{
  "node": "ovn-control-plane",
  "table": "nat",
  "filter_parameters": "-nv4"
}
```

```json
{
  "node": "ovn-control-plane",
  "table": "nat",
  "command": "-L"
}
```

---

## get-nft

Lists packet filtering and classification rules via `nft`.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `node` | string | **yes** | — | Name of the node from where packet filtering and classification rules are expected to be extracted |
| `namespace` | string | no | `"default"` | Namespace of the debug pod from where packet filtering and classification rules are expected to be extracted |
| `command` | string | **yes** | — | These options specify the desired action to perform. Only one of them can be specified on the command line unless otherwise stated below. `list ruleset`: The ruleset keyword is used to identify the whole set of tables, chains, etc. Print the ruleset in human-readable format. `list tables`: List all chains and rules of the specified table. `list chains`: List all rules of the specified chain. `list sets`: Display the elements in the specified set. `list maps`: Display the elements in the specified map. `list flowtables`: List all flowtables |
| `address_families` | string | no | — | Address families determine the type of packets which are processed. For each address family, the kernel contains so-called hooks at specific stages of the packet processing paths, which invoke nftables if rules for these hooks exist. `ip`: IPv4 address family. `ip6`: IPv6 address family. `inet`: Internet (IPv4/IPv6) address family. `arp`: ARP address family, handling IPv4 ARP packets. `bridge`: Bridge address family, handling packets which traverse a bridge device. `netdev`: Netdev address family, handling packets on ingress and egress |

Also accepts common [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting) and [`timeout_seconds`](user-guide.md#per-call-timeout).

### Examples

```json
{
  "node": "ovn-control-plane",
  "command": "list tables",
  "address_families": "inet"
}
```

---

## get-ip

Lists routing, network devices, and interfaces via `ip`.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `node` | string | **yes** | — | Name of the node on which ip command is expected to be executed |
| `namespace` | string | no | `"default"` | Namespace of the debug pod on which ip command is expected to be executed |
| `options` | string | no | — | These options helps in providing more details or formatting output data. `-d`/`-details`: Output more detailed information. `-4`: shortcut for `-family inet`. `-6`: shortcut for `-family inet6`. `-r`/`-resolve`: use the system's name resolver to print DNS names instead of host addresses. `-n`/`-netns <NETNS>`: switches ip to the specified network namespace NETNS. `-a`/`-all`: executes specified command over all objects, it depends if command supports this option |
| `command` | string | **yes** | — | These options specify the desired action to perform. Only one of them can be specified on the command line unless otherwise stated below. `address show`: protocol (IP or IPv6) address on a device. `link show`: network device. `neighbour show`: manage ARP or NDISC cache entries. `netns show`: manage network namespaces. `route show`: routing table entry. `rule show`: rule in routing policy database. `vrf show`: manage virtual routing and forwarding devices. `xfrm state list`: show Security Association Database. `xfrm policy list`: show Security Policy Database |
| `filter_parameters` | string | no | — | This allows to mention sub command to get more filtered data. Available sub command varies and supportability depends on what is already supported with `ip` utility |

Also accepts common [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting) and [`timeout_seconds`](user-guide.md#per-call-timeout).

### Examples

```json
{
  "node": "ovn-control-plane",
  "options": "-4",
  "command": "route show",
  "filter_parameters": "table all"
}
```
