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

The tool prefers the `conntrack` CLI in the configured `--kernel-image`. If that binary is missing, it falls back to reading `/proc/net/nf_conntrack`, which only supports dump/list (`-L`/`--dump`) and does not apply filters. Count (`-C`/`--count`) and stats (`-S`/`--stats`) require the `conntrack` CLI.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `node` | string | **yes** | — | Name of the node from where conntrack entries are expected to be extracted |
| `namespace` | string | no | `"default"` | Namespace of the debug pod from where conntrack entries are expected to be extracted |
| `command` | string | no | `"-L"` | These options specify the particular operation to perform. These options can only be used if configured image has `conntrack` utility available. If omitted or empty, defaults to `-L`. `-L`/`--dump`: List connection tracking table. `-C`/`--count`: Show the table counter. `-S`/`--stats`: Show the in-kernel connection tracking system statistics |
| `filter_parameters` | string | no | — | Read-only dump filters and output options (see below) |

Accepted filters follow the [conntrack manual](https://www.netfilter.org/projects/conntrack-tools/conntrack-manpage.html) and its dump option tables:

| Options | Purpose |
|---------|---------|
| `-s`/`--src`/`--orig-src`, `-d`/`--dst`/`--orig-dst`, `-r`/`--reply-src`, `-q`/`--reply-dst` | Original and reply addresses |
| `--mask-src`, `--mask-dst` | Address masks |
| `-p`/`--proto`/`--protonum`, `-f`/`--family` | Layer four protocol and IPv4/IPv6 family |
| `--sport`/`--orig-port-src`, `--dport`/`--orig-port-dst`, `--reply-port-src`, `--reply-port-dst`, `--state` | Protocol ports and state |
| `-m`/`--mark`, `-c`/`--secmark`, `-u`/`--status`, `-l`/`--label` | Marks, status, and labels |
| `-w`/`--zone`, `--orig-zone`, `--reply-zone` | Connection tracking zones |
| `--icmp-type`, `--icmp-code`, `--icmp-id`, `--icmpv6-type`, `--icmpv6-code`, `--icmpv6-id` | ICMP fields |
| `--srckey`/`--orig-key-src`, `--dstkey`/`--orig-key-dst`, `--reply-key-src`, `--reply-key-dst` | GRE keys |
| `-o`/`--output` | Output format, such as `extended,id` |
| `-n`/`--src-nat`, `-g`/`--dst-nat`, `-j`/`--any-nat` | NAT filters, optionally followed by an `ADDRESS[:PORT]` value |

All options require a non-empty value except the NAT filters. Values may be separate (`-p tcp`), attached to short options (`-ptcp`), or use `--name=value`. Supply optional NAT values separately. Conntrack validates protocol-specific values and option combinations. An optional full table name (`conntrack`, `expect`, `dying`, or `unconfirmed`) may precede the options.

Counter clearing (`-z`/`--zero`), alternate operations, and unlisted options or abbreviations are rejected before creating a debug pod. Filters apply only to dump/list; count, stats, and the file fallback do not apply them.

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
| `filter_parameters` | string | no | — | Optional chain and rule number, followed by read-only listing options (see below) |

Accepted options are `-v`/`--verbose`, `-n`/`--numeric`, `-x`/`--exact`, `--line-numbers`, `-w`/`--wait [seconds]`, `-4`/`--ipv4`, and `-6`/`--ipv6`. A chain name and optional positive rule number must come before the options (for example, `INPUT 3 -nv4`). Wait seconds must be a non-negative integer.

Short flags may be combined (`-nv4`); wait values may be separate (`-w 5`), attached (`-w5`), or use `--wait=5`. IPv6 is selected only by an explicit `-6`/`--ipv6` flag, including `-6` in a short-flag bundle.

Other options and abbreviated long option names are rejected; select the table through `table`. Source (`-s`), destination (`-d`), and protocol (`-p`) selectors define or check rules; iptables does not support them with listing commands.

Chain selection works with both `-L`/`--list` and `-S`/`--list-rules`. For example, `command: "-S"` with `filter_parameters: "INPUT"` lists only the INPUT chain. With `-S`/`--list-rules`, `-n`/`--numeric`, `-x`/`--exact`, and `--line-numbers` are rejected before creating a debug pod, including when short flags are combined. IPv4/IPv6 selection, verbose output, and wait options remain supported.

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
| `options` | string | no | — | Diagnostic global options for details, output formatting, protocol family, and network namespace (see below) |
| `command` | string | **yes** | — | These options specify the desired action to perform. Only one of them can be specified on the command line unless otherwise stated below. `address show`: protocol (IP or IPv6) address on a device. `link show`: network device. `neighbour show`: manage ARP or NDISC cache entries. `netns show`: manage network namespaces. `route show`: routing table entry. `rule show`: rule in routing policy database. `vrf show`: manage virtual routing and forwarding devices. `xfrm state list`: show Security Association Database. `xfrm policy list`: show Security Policy Database |
| `filter_parameters` | string | no | — | This allows to mention sub command to get more filtered data. Available sub command varies and supportability depends on what is already supported with `ip` utility |

Accepted global options follow the [ip(8) manual](https://man7.org/linux/man-pages/man8/ip.8.html):

| Options | Purpose |
|---------|---------|
| `-d`/`-details`, `-s`/`-stats`/`-statistics` | Details and statistics; repeat `-s` for more statistics |
| `-o`/`-oneline`, `-j`/`-json`, `-p`/`-pretty`, `-br`/`-brief` | Output formatting |
| `-h`/`-human`/`-human-readable`, `-iec`, `-N`/`-Numeric`, `-r`/`-resolve` | Units and name resolution |
| `-a`/`-all`, `-t`/`-timestamp`, `-ts`/`-tshort` | All objects and timestamps, where supported by the command |
| `-c`/`-color[=always\|auto\|never]` | Color output |
| `-4`, `-6`, `-0`, `-B`, `-M` | IPv4, IPv6, link, bridge, or MPLS family |
| `-f`/`-family FAMILY` | Family: `inet`, `inet6`, `link`, `bridge`, or `mpls` |
| `-n`/`-netns NETNS` | Network namespace |
| `-rc`/`-rcvbuf SIZE` | Netlink receive buffer size (unsigned 32-bit integer) |

Use these exact names with one or two leading hyphens. Namespace, family, and buffer options require a separate value; only color accepts `=value`. Short flags must be separate (`-s -d`, not `-sd`). Batch mode, other options, and unlisted abbreviations are rejected before creating a debug pod.

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
