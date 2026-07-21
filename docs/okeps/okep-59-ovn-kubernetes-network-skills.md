# OKEP-59: Network Troubleshooting Skills for OVN-Kubernetes

* Issue: [#59](https://github.com/ovn-kubernetes/ovn-kubernetes-mcp/issues/59)

# Problem Statement

The `ovn-kubernetes-mcp` server today exposes ~20 read-only debugging
tools across 5 layers - Kubernetes API (`resource-get`, `resource-list`,
`pod-logs`), OVN NB/SB (`ovn-show`, `ovn-get`, `ovn-lflow-list`,
`ovn-trace`), OVS (`ovs-vsctl-show`, `ovs-list-br`, `ovs-list-ports`,
`ovs-list-ifaces`, `ovs-ofctl-dump-flows`, `ovs-appctl-dump-conntrack`,
`ovs-appctl-ofproto-trace`), kernel networking (`get-conntrack`,
`get-iptables`, `get-nft`, `get-ip`), and packet capture (`tcpdump`).
Those tools are the *inputs* to a triage; they are not the triage
itself. What is missing today is any guidance for the LLM on how to
compose them into a deterministic, layered troubleshooting workflow.

Without that guidance, LLMs misuse the tool surface in predictable
ways:

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
  Anthropic Skills API, custom agents) without forking
  per-client.
* Ship an **initial set of few skills** covering the most frequent
  Phase-1 triage paths: `pod-to-pod-connectivity`,
  `service-connectivity`, `network-policy-debugging`,
  `external-connectivity`, `node-networking`, `ovn-topology-discovery`,
  `ovs-flow-analysis`, `conntrack-debugging`, `packet-drop-investigation`.
* Skills MUST only call **named MCP tools** from OKEP-5494 (no free-form
  bash), MUST be **read-only**, and MUST encode layered, evidence-driven
  workflows.
* Skills MUST live in-tree and be code-reviewed; a Phase-2 CI lint against
  the MCP tool registry adds automated drift protection.

# Future Goals

* **Feature-specific skills**: ovn-k8s skills related to features in the main product `egressip`,
  `egress-firewall`, `route-advertisements`, `ipsec`, `bgp`,
  `udn-primary-network`, `multus-secondary-networks`, etc.
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
4. **Context economics.** A monolithic system prompt covering all N
   scenarios is in context for every conversation. Agent Skills load
   only the matching skill body via *progressive disclosure*: the
   `description` is always loaded; the body is loaded only when the
   user's symptom matches.

## Two-layer contract

```text
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
| `external-connectivity`     | North-south egress/ingress, SNAT/DNAT issues | ovn-get NAT/Logical_Router/Static_Route, ovn-trace, ovs-vsctl-show, get-ip route, get-iptables nat, get-nft, tcpdump br-ex |
| `node-networking`           | NetworkUnavailable, tunnel down, missing routes | resource-get Node, get-ip address/link/route/rule/neigh, ovs-vsctl-show, ovn-get Chassis, get-iptables, get-nft, pod-logs ovn-controller |
| `ovn-topology-discovery`    | Map / understand the logical network | ovn-show NB/SB, ovn-get LS/LR/LSP/LRP/Chassis/Port_Binding/NAT/Static_Route/Load_Balancer, ovs-vsctl-show |
| `ovs-flow-analysis`         | Logical vs OpenFlow mismatch, stale flows | ovn-lflow-list, ovs-ofctl-dump-flows, ovs-appctl-ofproto-trace, ovn-trace, pod-logs ovn-controller |
| `conntrack-debugging`       | NAT broken, stale entries, table-full, zone conflicts | ovs-appctl-dump-conntrack, get-conntrack -L/-S/-C, ovs-ofctl-dump-flows ct, ovn-lflow-list ct |
| `packet-drop-investigation` | Silent drops, unknown drop point | tcpdump, ovs-ofctl-dump-flows drop, ovs-appctl-ofproto-trace, ovn-trace, ovs-appctl-dump-conntrack, get-iptables, get-nft, pwru |

Skills are independent (pick one by symptom) and **composable** -
`external-connectivity` calls into `node-networking`, and
`packet-drop-investigation` calls into `ovs-flow-analysis` and
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

**As a layered-feature owner** (EgressIP, IPSec, BGP, RouteAdvertisements,
UDN, ...), I want to publish a skill for my feature's failure modes so the
LLM does not have to rediscover them.

# Proposed Solution

## Repository layout

```text
ovn-kubernetes-mcp/
  skills/                              <- canonical, vendor-neutral
    README.md                          <- human-readable index + format
                                          reference (agents load per-skill
                                          SKILL.md, not this file)
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

```text
User: "curl-front-7c4 (ns shop) and api-back-9d2 (ns shop) can't talk
       on TCP/8080. Used to work."

Agent:
  - matches symptom -> loads pod-to-pod-connectivity
  - resource-get both pods                -> podIP, nodeName, Running
  - resource-list ovnkube pods            -> picks src/dst ovnkube-node
                                             (each hosts its own NB/SB/OVS
                                             in OVN-IC mode)
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

1. Author the initial set of `SKILL.md` files under canonical `skills/`
   (one directory per skill from the catalogue above). Add
   `skills/README.md` as the human-readable index + format reference.
2. Add `.cursor/skills` and `.claude/skills` symlinks.
3. Add a short docs page on pointing Cursor / Claude Code / `@cursor/sdk`
   at the server and the available skills.

In Phase-1, skill correctness is enforced by PR review (well-formedness)
plus the contributor's eval run (behaviour, see Testing Strategy). There
is no automated skill gate yet.

Phase-2 deliverables:

1. **`skills-lint` (`make skills-lint`, wired into CI).** A static checker
   that parses each `SKILL.md`, validates frontmatter, resolves every tool
   reference against the MCP server's tool registry, rejects free-form
   bash and write verbs, and enforces the reporting block and line cap.
   Implement it in Go by importing the server's `RegisteredToolNames()` so
   the lint and the server share one source of truth - renaming a tool
   then fails CI on every skill that references it, instead of breaking at
   runtime.
2. **Skills as MCP resources.** Add a resource provider exposing each skill
   at e.g. `mcp://ovn-kubernetes/skills/<name>` with `mimeType:
   text/markdown`. Additive; does not change the file format.

# Security Model

Skills inherit OKEP-5494's model unchanged. They neither expand nor relax
it:

* Skills only call **named MCP tools**. They cannot escape the tool
  sandbox.
* Skills are **read-only by construction** - PR review (and the Phase-2
  `skills-lint`) reject write verbs.
* Skills are **public**, reviewed in the open. No secrets, no IPs,
  no customer data; placeholders only (`<src-pod>`, `<node>`, `<ovn-pod>`).
* Skills carry no RBAC of their own. They run with whatever ServiceAccount
  / kubeconfig the MCP server uses.
* **Remediation** appears only as text in the Reporting block, addressed
  to the human. The agent does not execute it under this OKEP.

| Threat | Mitigation |
|---|---|
| Malicious skill PR adds bash/exec | PR review (Phase-1); Phase-2 `skills-lint` forbids free-form shell |
| Stale skill references removed tool | PR review (Phase-1); Phase-2 `skills-lint` cross-checks the live tool registry |
| Prompt-injection via tool output | Out of skill / lint scope. Repo-loading controls how skills are *sourced*; it does not stop tool output from influencing the agent context. That trust boundary belongs to the MCP client / agent runtime, which must treat tool output as untrusted data. |
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
  the destination node - in OVN-IC mode (what OKEP-5494 targets) each
  of these pods hosts its own local NB/SB databases and OVS instance,
  so every OVN/OVS query for that node routes through the same pod,
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

**Configuration.** In Phase-2, operators selectively disable individual
skills via a `skills.disabled` list in the MCP server config (per
OKEP-5494's config file), which hides matching skills from the resource
provider. Phase-1 has no config-driven toggle: a client is pointed at the
`skills/` directory as a whole (directory-level, all-or-nothing), so
selectively disabling one skill means omitting its directory from
`skills/`, and disabling everything means not pointing the client at
`skills/` at all.

# Testing Strategy

A skill is fundamentally a prompt for an agent, so it is validated the
way prompts are - with **evals**, not conventional unit tests. This
follows the emerging practice for agent skills in the wider ecosystem
(see OpenAI's
[Testing Agent Skills Systematically with Evals](https://developers.openai.com/blog/eval-skills)).
An eval is a simple loop:

`symptom prompt → captured agent run (tool-call trace + final RCA) → checks → score`

## Who runs evals, and when

Evals are **run by the contributor** who authors or changes a skill, not
by CI. Standing up an OVN-Kubernetes cluster and driving live-cluster
tools (`ovn-trace`, `ovs-appctl-ofproto-trace`, `tcpdump`, `pwru`, ...)
is impractical to run as an automated gate, and the result depends on the
backing LLM, which we do not pin. Instead:

* The author runs the skill's eval prompts against a cluster they have
  access to (a disposable kind cluster is sufficient or a production
  cluster) and **pastes the eval summary into the PR**: which prompts
  triggered the skill, the tool-call sequence, the resulting RCA, and
  pass/fail per check.
* Reviewers use that summary as evidence, the same way they would review
  test output, alongside reading the skill itself.

In Phase-1 there is no automated skill gate - well-formedness is checked
in PR review and behaviour by the contributor's eval run. Phase-2 adds the
static `skills-lint` (Implementation Details) as an automated CI gate for
well-formedness (frontmatter, tool-reference resolution, no free-form
bash, reporting block, line cap). Either way, review / lint validates that
a skill is *well-formed*; the contributor's eval run validates that it
*works*.

## What an eval checks

Following the standard eval taxonomy, a skill is graded on four axes:

* **Triggering** - given a symptom prompt, did the correct skill get
  selected from its `description`? Include **negative controls** (prompts
  that must NOT trigger it) to catch over-eager matching.
* **Process** - did the agent follow the skill's layered sequence,
  calling the expected tools on the expected pods/nodes?
* **Outcome** - did the final RCA name the correct root cause and cite the
  tool output that proves it?
* **Efficiency** - did it get there without thrashing (tool-call count,
  token use) versus running without the skill?

## Eval prompt set

Each skill ships with a small (~10-20 prompt) set of symptom phrasings -
explicit ("use the pod-to-pod-connectivity skill"), implicit ("pods on
node A can't reach node B"), and negative controls - kept next to the
skill. It grows over time: every real miss found in use becomes a new
prompt, so the set becomes a living regression record the next
contributor re-runs.

## Optional: offline replay

Where a scenario has a captured `must-gather` / `sos-report` bundle, a
contributor may run the same eval against OKEP-5494's offline mode with no
cluster - a cheaper way to re-check the read-only steps of a skill. This
is a convenience, not a requirement.

# Documentation

OKEP on [https://ovn-kubernetes.io/](https://ovn-kubernetes.io/). End-user
docs in `docs/skills/`:

* **Getting started** - how to point Cursor / Claude Code / Cursor CLI /
  `@cursor/sdk` at the server and the skills folder.
* **Skill reference** - rendered list with frontmatter + tools used.
* **Authoring guide** - format, lint rules, review.
* **Cookbook** - 5-10 worked examples (paired with the eval prompt sets
  so users can replay).
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
  Mitigation: PR review catches semantic drift; the contributor's eval
  run catches behavioural drift; the Phase-2 `skills-lint` catches
  mechanical breakage (stale tool references, bad frontmatter).
* **Over-trust.** Engineers may stop verifying RCAs because "the skill
  ran". Skills are first-pass triage, not authoritative; docs and the
  reporting block frame them that way.
* **Skill bias.** A skill encodes one path; real bugs occasionally need
  another. Mitigation: each skill ends with a decision tree pointing at
  sibling skills; the evidence-citation requirement makes a wrong-path
  conclusion detectable.
* **Maintenance cost.** Skills are documents reviewed per PR like code.
  Mitigation: PR review in Phase-1, the Phase-2 `skills-lint` for
  mechanical breakage, and layered-feature ownership to spread the load.
* **Security carry-over.** All caveats from OKEP-5494 apply unchanged.
  Skills do not improve the underlying read-only guarantees; they
  orchestrate within them.
