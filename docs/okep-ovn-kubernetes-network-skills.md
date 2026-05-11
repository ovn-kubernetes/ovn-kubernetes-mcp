# OKEP: Network Troubleshooting Skills for OVN-Kubernetes

> Companion to [OKEP-5494: Model Context Protocol for Troubleshooting OVN-Kubernetes](https://github.com/ovn-kubernetes/ovn-kubernetes/blob/master/docs/okeps/okep-5494-ovn-kubernetes-mcp-server.md).
> OKEP-5494 gives the LLM safe, layered *tools*. This OKEP gives it expert
> *guidance* on how to use them.

# Problem Statement

OKEP-5494 exposes 20+ debugging tools across 5 layers (K8s, OVN, OVS, kernel,
capture) over MCP. Tools alone are not enough: LLMs misuse them in
unpredictable ways that OKEP-5494 itself flagged as risks - poor tool
selection, parameter hallucination, attention spreading, and
misinterpretation. In practice this looks like:

* Jumping to `tcpdump` before checking pod state or running `ovn-trace`.
* Inventing parameter values - container names (`nb-ovsdb` vs
  `sb-ovsdb`), datapath / logical-switch names, table names, conntrack
  zones, microflow expressions - instead of carrying them forward from
  earlier tool output.
* Skipping layers - one big call, no methodology.
* Missing tribal knowledge ("a NetworkPolicy that selects a pod denies
  *all* traffic in that direction unless explicitly allowed").

This is not a model problem alone; it is a guidance problem. Engineers who
debug OVN-Kubernetes follow a layered, evidence-driven approach that
nobody has codified. New team members and the LLM both have to rediscover
it.

This OKEP introduces **Network Troubleshooting Skills**: short,
version-controlled markdown playbooks that the agent loads
on demand and follows step-by-step, calling MCP tools in a deterministic
layered order.

Skills add a layer of **determinism** to an otherwise non-deterministic
agent loop: the *which tools, in what order, with which parameters
threaded between them* part of triage is fixed by the skill, leaving the
LLM responsible only for symptom matching, parameter extraction from
user prompts, and final analysis. The same bug investigated twice
should produce substantially the same transcript, which makes
reproducibility, review, and regression testing tractable.

# Goals

* Define a **skill format** and **skill catalogue** for OVN-Kubernetes
  network triage shipped alongside the MCP server.
* Adopt the cross-vendor [Agent Skills](https://www.anthropic.com/news/skills)
  convention (`SKILL.md` with `name` + `description` frontmatter) so skills
  work in any compatible runtime (Cursor, Cursor CLI, Claude Code,
  Anthropic Skills API, custom `@cursor/sdk` agents) without forking
  per-client.
* Ship an **initial set of few skills** covering the most frequent
  Phase-1 triage paths: `pod-to-pod-connectivity`,
  `service-connectivity`, `network-policy-debugging`, `dns-resolution`,
  `external-connectivity`, `node-networking`, `ovn-topology-discovery`,
  `ovs-flow-analysis`, `conntrack-debugging`, `packet-drop-investigation`.
* Skills MUST only call **named MCP tools** from OKEP-5494 (no free-form
  bash), MUST be **read-only**, and MUST encode layered, evidence-driven
  workflows.
* Skills MUST live in-tree, be code-reviewed, and be CI-linted against the
  MCP tool registry so they cannot drift.

# Future Goals

* **Feature-specific skills**: ovn-k8s skills related to features in the main product `egressip`,
  `egress-firewall`, `route-advertisements`, `ipsec`, `bgp`,
  `udn-primary-network`, `staleSNAT`, `multus-secondary-networks`, etc.
* **Skills as MCP resources** - serve `SKILL.md` over the wire so remote
  MCP clients can fetch them from the server without local file access.
* **Skill benchmark suite** - per-skill fixtures (must-gather /
  sos-report bundle + expected RCA) feeding the LLM Capability
  Assessment in OKEP-5494.

# Future-Stretch-Goals

* Skills used in production troubleshooting by end-users running
  OVN-Kubernetes clusters (rather than only by community developers).
  Same gating as OKEP-5494's Phase-3 stretch goal: requires solving
  where/how to run the model, an end-to-end agentic-AI architecture
  on the cluster, and an air-tight security review. Out of scope until
  those land.

# Non-Goals

* **No new format.** We adopt Agent Skills exactly. Inventing our own
  defeats portability across MCP/agent runtimes.
* **We do not own client loading.** How Cursor or Claude Code finds a
  `SKILL.md` is the client's concern.
* **Not a tool replacement.** Skills are guidance; tools live in
  OKEP-5494. Without the server, a skill is just a runbook.
* **No write / remediation.** Phase 1 stays read-only.
* **Not solving LLM quality.** A bad model + a good skill still fails;
  skills bound the failure mode, they do not eliminate it.

# Introduction

## Why skills, not just tool descriptions or a system prompt?

Tool descriptions cannot encode:

1. **Methodology.** "Always check the pod is Running before running
   `ovn-trace`" is a workflow rule, not a property of any single tool.
2. **Cross-tool wiring.** `addresses` from `ovn-get Logical_Switch_Port`
   becomes the `eth.src`/`ip4.src` of `ovn-trace`. That graph lives
   between tools, not inside any one of them.
3. **Decision trees.** "If `ovn-trace` allows but `ofproto-trace`
   drops, suspect flow-installation; restart `ovn-controller`" is
   exactly what skills are for.
4. **Context economics.** A monolithic system prompt covering all 10
   scenarios is in context for every conversation. Agent Skills load
   only the matching skill body via *progressive disclosure*: the
   `description` is always loaded; the body is loaded only when the
   user's symptom matches.

## Two-layer contract

```
              +-------------------------------------------+
              |          MCP Client + LLM Agent           |
              |  (Cursor / Claude Code / @cursor/sdk)     |
              +--------+---------------------+------------+
                       | loads               | calls
                       v                     v
              +-----------------+   +---------------------+
              |   SKILL.md      |   |   MCP Tools         |
              |   (this OKEP)   |   |   (OKEP-5494)       |
              | frontmatter +   |   | resource-get,       |
              | layered steps + |   | ovn-show, ovn-get,  |
              | decision tree   |   | ovn-trace,          |
              |                 |   | ovs-ofctl-...,      |
              | references --------->  tcpdump, pwru, ... |
              | tools by name   |   |                     |
              +-----------------+   +---------------------+
                       \                  /
                        v                v
                       ovnkube-node pods, NB/SB DBs,
                       OVS, kernel, NIC
```

A skill is a deterministic recipe over the MCP tool catalogue. If the
catalogue changes, the affected skills change in the same PR -
enforced by CI.

## Initial skill catalogue thoughts

| Skill | When to use | Primary MCP tools |
|---|---|---|
| `pod-to-pod-connectivity`   | Pod A can't reach Pod B | resource-get, ovn-show, ovn-get LSP/PB/Chassis, ovn-trace, ovs-appctl-ofproto-trace, get-conntrack, tcpdump, pwru |
| `service-connectivity`      | ClusterIP/NodePort/LB unreachable, missing endpoints | resource-get Service/EndpointSlice, ovn-get Load_Balancer, ovn-trace, ovs-appctl-dump-conntrack, get-iptables, get-nft |
| `network-policy-debugging`  | Traffic unexpectedly blocked/allowed | resource-list NetworkPolicy, ovn-get ACL/Address_Set/Port_Group, ovn-trace, ovn-lflow-list, ovs-appctl-ofproto-trace |
| `dns-resolution`            | NXDOMAIN, DNS timeout, kube-dns unreachable | resource-get Service/Pod/ConfigMap, pod-logs CoreDNS, tcpdump :53, ovn-get Load_Balancer, get-conntrack udp/53 |
| `external-connectivity`     | North-south egress/ingress, SNAT/DNAT issues | ovn-get NAT/Logical_Router/Static_Route, ovn-trace, ovs-vsctl-show, get-ip route, get-iptables nat, get-nft, tcpdump br-ex |
| `node-networking`           | NetworkUnavailable, tunnel down, missing routes | resource-get Node, get-ip address/link/route/rule/neigh, ovs-vsctl-show, ovn-get Chassis, get-iptables, get-nft, pod-logs ovn-controller |
| `ovn-topology-discovery`    | Map / understand the logical network | ovn-show NB/SB, ovn-get LS/LR/LSP/LRP/Chassis/Port_Binding/NAT/Static_Route/Load_Balancer, ovs-vsctl-show |
| `ovs-flow-analysis`         | Logical vs OpenFlow mismatch, stale flows | ovn-lflow-list, ovs-ofctl-dump-flows, ovs-appctl-ofproto-trace, ovn-trace, pod-logs ovn-controller |
| `conntrack-debugging`       | NAT broken, stale entries, table-full, zone conflicts | ovs-appctl-dump-conntrack, get-conntrack -L/-S/-C, ovs-ofctl-dump-flows ct, ovn-lflow-list ct |
| `packet-drop-investigation` | Silent drops, unknown drop point | tcpdump, ovs-ofctl-dump-flows drop, ovs-appctl-ofproto-trace, ovn-trace, ovs-appctl-dump-conntrack, get-iptables, get-nft, pwru |

Skills are independent (pick one by symptom) and **composable** -
`dns-resolution` calls into `service-connectivity`,
`external-connectivity` into `node-networking`,
`packet-drop-investigation` into `ovs-flow-analysis` and
`conntrack-debugging`.

# User Stories

**As an OVN-Kubernetes developer**, I want the AI assistant to follow the
layered triage methodology I would (pod state → NB → SB → OVS → kernel →
capture) so I get a consistent, auditable RCA instead of a model
improvising from training data.

**As a new engineer joining the team**, I want a published skill I can
read or run so I learn the canonical debug flow without shoulder-surfing
a senior engineer.

**As a triage engineer**, I want the agent to fetch parameters from earlier
tool output (MACs, port names, datapath names) instead of asking me to
type them.

**As a maintainer reviewing a bug**, I want the LLM transcript to map 1:1
to a published skill so I can audit each step and improve the skill in the
same PR that fixes the bug.

**As a layered-feature owner** (EgressIP, IPSec, BGP, RouteAdvertisements,
UDN, ...), I want to publish a skill for my feature's failure modes so the
LLM does not have to rediscover them.

# Proposed Solution

## Repository layout

```
ovn-kubernetes-mcp/
  skills/                              <- canonical, vendor-neutral
    SKILLS.md                          <- index + format reference
    pod-to-pod-connectivity/SKILL.md
    service-connectivity/SKILL.md
    ...
  .cursor/skills -> ../skills          <- symlink for Cursor / Cursor CLI
  .claude/skills -> ../skills          <- symlink for Claude Code
```

A single canonical `skills/` is the source of truth (and what we will
serve as MCP resources later). Per-client symlinks make it discoverable
in clients that look in `.cursor/skills` or `.claude/skills` without
duplicating files.

## Skill format

```markdown
---
name: <kebab-case-skill-name>             # matches directory
description: <symptom-oriented; must
  contain "Use when ..." trigger phrases>
---

# <Title>

## Prerequisites
- inputs to collect from the user

## Step 1 ... Step N
<imperative title>
<MCP tool call: tool name + parameters>
<what to extract; which fields feed which later step>

## Decision Tree
<branching summary; may hop to a sibling skill>

## Reporting Findings
1. Root cause   2. Evidence   3. Remediation   4. Verification
```

**Author rules** (CI-enforced):

1. Frontmatter present; `name` matches directory; `description` non-empty
   and contains "Use when ..." phrasing.
2. Every tool reference resolves to a registered MCP tool, every
   parameter is a real parameter of that tool.
3. No free-form bash, ssh, `kubectl exec`, or write verbs (`set`, `add`,
   `del`, `remove`, ...).
4. Ends with the four-bullet reporting block.
5. ≤ 300 lines. Cross-reference sibling skills instead of duplicating.

## How the agent uses a skill

1. Indexes all `SKILL.md`; only frontmatter is loaded by default.
2. On a user prompt, matches against `description` fields; loads the
   matching body.
3. Follows the steps, calling MCP tools with parameters extracted from
   the user prompt and prior tool output.
4. At each decision rule, continues, hops to a sibling skill, or
   returns the reporting block.

We do not extend MCP, do not require a new client feature, and do not
fork existing clients. Works today in Cursor, Cursor CLI, Claude Code,
Anthropic Skills API, and custom agents using `@cursor/sdk`.

## Worked example - "Pod A can't reach Pod B"

```
User: "curl-front-7c4 (ns shop) and api-back-9d2 (ns shop) can't talk
       on TCP/8080. Used to work."

Agent:
  - matches symptom -> loads pod-to-pod-connectivity
  - resource-get both pods                -> podIP, nodeName, Running
  - resource-list ovnkube pods            -> picks src/dst ovnkube-node
                                             and one ovnkube-control-plane
  - ovn-show NB; ovn-get LSP for both;
    ovn-get Port_Binding; ovn-get Chassis -> LSPs up=true, ports bound
  - resource-list NetworkPolicy ns=shop   -> finds default-deny-egress
  - ovn-trace with eth/ip from above      -> drop in ls_in_acl
  - decision tree -> hops to network-policy-debugging
                  -> identifies the offending ACL
  - Reports:
      Root cause   : NetworkPolicy 'default-deny-egress' selects
                     curl-front; matching allow-rule is missing.
      Evidence     : ovn-trace stage ls_in_acl, ACL UUID xxx,
                     external_ids k8s.ovn.org/name=default-deny-egress.
      Remediation  : Add an egress rule allowing TCP/8080 to api-back.
      Verification : Re-run ovn-trace; expect ls_in_acl allow-related.
```

The agent never invented a MAC, never opened tcpdump prematurely, and
produced a transcript a maintainer can read like a runbook execution log.

## Authoring and review process

Skills are code:

* In-tree under `skills/`, reviewed by an OVN-Kubernetes maintainer (or
  the layered-feature owner once feature skills land).
* MCP tool catalogue changes that affect skills must update both in the
  same PR; CI rejects PRs that leave a skill referencing a removed tool.

# Implementation Details

Phase-1 deliverables:

1. Move the drafts under
   [`hack/network_skills/.cursor/skills/`](../hack/network_skills/.cursor/skills/)
   to canonical `skills/`. Add `skills/SKILLS.md` (index + format
   reference).
2. Add `.cursor/skills` and `.claude/skills` symlinks.
3. Add `make skills-lint`: parses each `SKILL.md`, validates frontmatter,
   checks every tool reference against the in-process MCP tool registry
   (`RegisteredToolNames()` already exists per OKEP-5494), rejects
   free-form bash and write verbs, enforces the reporting block and the
   line-count cap.
4. Add a short docs page on pointing Cursor / Claude Code / `@cursor/sdk`
   at the server and the available skills.

**Tool registry coupling.** The MCP server and the skills lint share the
same `RegisteredToolNames()` source of truth. Renaming a tool is a CI
failure on every skill that references it - not a runtime regression.

**Skills as MCP resources (Phase 2).** When ready, add a resource provider
exposing each skill at e.g. `mcp://ovn-kubernetes/skills/<name>` with
`mimeType: text/markdown`. Additive; does not change the file format.

# Security Model

Skills inherit OKEP-5494's model unchanged. They neither expand nor relax
it:

* Skills only call **named MCP tools**. They cannot escape the tool
  sandbox.
* Skills are **read-only by construction** - lint rejects write verbs.
* Skills are **public**, reviewed in the open. No secrets, no IPs,
  no customer data; placeholders only (`<src-pod>`, `<node>`, `<ovn-pod>`).
* Skills carry no RBAC of their own. They run with whatever ServiceAccount
  / kubeconfig the MCP server uses.
* **Remediation** appears only as text in the Reporting block, addressed
  to the human. The agent does not execute it under this OKEP.

| Threat | Mitigation |
|---|---|
| Malicious skill PR adds bash/exec | CI lint forbids free-form shell; PR review |
| Stale skill references removed tool | CI lint cross-checks the live tool registry |
| Prompt-injection via tool output | Skills are repo-loaded, not output-loaded; standard MCP runtime guarantee |
| Skill leaks customer IPs/hostnames | PR review; placeholders only |
| Skill drifts from feature reality | Layered-feature ownership in Phase 2; PR review by maintainers |

# Targeting the Right Pod / Node

OVN-Kubernetes is a distributed system: each node has its own NB DB,
SB DB, OVS instance, and kernel state. Every MCP tool that touches
those layers therefore needs to know *which* `ovnkube-node` pod (or
which node) to run against. Without skills, the LLM has to recompute
this mapping for every individual tool call, which is where it
typically picks the wrong pod.

Skills solve this by making "find the right pods" an explicit early
step in every workflow. For example, Step 2 of `pod-to-pod-connectivity`
is:

* `resource-list` the ovnkube pods,
* identify the `ovnkube-node` pod on the source node and the one on
  the destination node, plus an `ovnkube-control-plane` pod for NB/SB
  queries,
* reuse those pod names as the `name=<pod>` parameter in every
  subsequent OVN/OVS tool call.

The MCP server itself stays stateless (per OKEP-5494); the skill is
where the "for this triage session, the source pod is X" decision is
recorded and reused.

# Deployment Strategy

1. **Bundled with the repo (Phase 1, default).** Clone the repo, point an
   Agent-Skills-compatible client at it, done.
2. **Per-client symlinks (Phase 1).** `.cursor/skills` and
   `.claude/skills` make discovery automatic.
3. **MCP resources (Phase 2).** Same skills, served over the wire to
   any MCP client - no local checkout required.

We deliberately do **not** ship a separate skills package or container
in Phase 1 - that is release-management cost without a real distribution
problem.

**Configuration.** Operators disable skills like they disable tools (per
OKEP-5494's config file): a `skills.disabled` list hides matching
`SKILL.md` from the resource provider in Phase 2; in Phase 1, simply do
not symlink them.

# Testing Strategy (TODO)

**Approach 1: replay against historical bundles (advisory).** Per-skill
fixture in `test/skills/<skill>/`: a curated must-gather / sos-report
bundle + an `expected.md` describing the correct RCA. A runner uses the
offline-execution mode from OKEP-5494 to point the MCP server at the
bundle, runs the agent loop, and diffs the transcript against
`expected.md`. Answers two questions: did the agent pick the right skill?
did it follow it?

**Approach 2: chaos / synthetic (Phase 1).** Per-skill chaos scenarios on a
kind / production cluster (delete an LSP, blackhole UDP/53, fill the
conntrack table, install a bad drop flow, ...). A job runs the
agent with skills enabled vs disabled and tracks the RCA-accuracy delta.
This is the benchmark dataset that backs OKEP-5494's "LLM Capability
Assessment".

# Documentation

OKEP on [https://ovn-kubernetes.io/](https://ovn-kubernetes.io/). End-user
docs in `docs/skills/`:

* **Getting started** - how to point Cursor / Claude Code / Cursor CLI /
  `@cursor/sdk` at the server and the skills folder.
* **Skill reference** - rendered list with frontmatter + tools used.
* **Authoring guide** - format, lint rules, review.
* **Cookbook** - 5-10 worked examples (paired with the Layer-2
  fixtures so users can replay).
* **FAQ / anti-patterns** - free-form bash, hardcoded parameters,
  cross-cluster assumptions, etc.

# Known Risks and Limitations

* **Client-coverage drift.** Agent Skills is real today in Cursor,
  Cursor CLI, Claude Code, and Anthropic's Skills API, but not yet
  universal. Clients that ignore `SKILL.md` get OKEP-5494's raw tool
  catalogue and the failure modes that motivated this OKEP. Mitigation:
  Phase 2 serves skills via MCP resources, the most portable channel.
* **LLM still in the loop.** Skills constrain the workflow but do not
  eliminate hallucination. The reporting block forces evidence
  citations, which makes review easier but does not prevent errors.
* **Skill staleness.** As OVN-Kubernetes evolves, skills can drift.
  Mitigation: Layer-1 lint catches mechanical breakage; PR review by
  maintainers catches semantic drift; Layer-2 replay catches
  behavioural drift.
* **Over-trust.** Engineers may stop verifying RCAs because "the skill
  ran". Skills are first-pass triage, not authoritative; docs and the
  reporting block frame them that way.
* **Skill bias.** A skill encodes one path; real bugs occasionally need
  another. Mitigation: each skill ends with a decision tree pointing at
  sibling skills; the evidence-citation requirement makes a wrong-path
  conclusion detectable.
* **Maintenance cost.** Skills are documents reviewed per PR like code.
  Mitigation: lint catches mechanical breakage; layered-feature
  ownership in Phase 2 spreads the load.
* **Security carry-over.** All caveats from OKEP-5494 apply unchanged.
  Skills do not improve the underlying read-only guarantees; they
  orchestrate within them.
