# OKEP-59: Network Troubleshooting Skills for OVN-Kubernetes

* Issue: [#59](https://github.com/ovn-kubernetes/ovn-kubernetes-mcp/issues/59)

# Problem Statement

The `ovn-kubernetes-mcp` server today exposes read-only debugging tools
across 5 layers - Kubernetes API (`resource-get`, `resource-list`,
`pod-logs`), OVN NB/SB (`ovn-show`, `ovn-get`, `ovn-lflow-list`,
`ovn-trace`), OVS (`ovs-vsctl`, `ovs-ofctl`, `ovs-appctl` - each takes an
`action` subcommand, e.g. `ovs-vsctl` action `show`/`list-br`/`list-ports`,
`ovs-ofctl` action `dump-flows`, `ovs-appctl` action
`dpctl/dump-conntrack`/`ofproto/trace`), kernel networking
(`get-conntrack`, `get-iptables`, `get-nft`, `get-ip`), and packet
capture / tracing (`tcpdump`, `pwru`). Those tools are the *inputs* to a
triage; they are not the triage itself. What is missing today is any
guidance for the LLM on how to compose them into a deterministic, layered
troubleshooting workflow.

Without that guidance, LLMs/Agents misuse the tool surface in unpredictable
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
* Adopt the cross-vendor
  [Agent Skills specification](https://agentskills.io/specification)
  (`SKILL.md` with `name` + `description` frontmatter) so skills work in any
  compatible runtime (Cursor, Cursor CLI, Claude Code, Anthropic Skills API,
  custom agents) without forking per-client.
* Ship an **initial set of skills** covering the most frequent
  Phase-1 triage paths: `pod-to-pod-connectivity`,
  `service-connectivity`, `external-connectivity`, `node-networking`.
* Skills MUST only call **named MCP tools** exposed by the server (no
  free-form bash), MUST be **read-only**, and MUST encode layered,
  evidence-driven workflows.
* Skills MUST live in-tree and be code-reviewed; a Phase-2 CI lint against
  the MCP tool registry adds automated drift protection.

# Future Goals

* **Feature-specific skills**: ovn-k8s skills related to features in the main product `egressip`,
  `egress-firewall`, `route-advertisements`, `ipsec`, `bgp`,
  `udn-primary-network`, `multus-secondary-networks`, etc.
* **Skills as MCP resources** - serve `SKILL.md` over the wire so remote
  MCP clients can fetch them from the server without local file access.
* **Skill benchmark suite** - per-skill evals (see Testing Strategy) run
  as a suite to measure troubleshooting quality over time.

# Future-Stretch-Goals

* Skills used in production troubleshooting by end-users running
  OVN-Kubernetes clusters (rather than only by community developers).
  This requires solving where/how to run the model, an end-to-end
  agentic-AI architecture on the cluster, and an air-tight security
  review. Out of scope until those land.

# Non-Goals

* **No new format.** We adopt Agent Skills exactly. Inventing our own
  defeats portability across MCP/agent runtimes.
* **We do not own client loading.** How Cursor or Claude Code finds a
  `SKILL.md` is the client's concern.
* **Not a tool replacement.** Skills are guidance; the tools live in the
  MCP server. Without the server, a skill is just a runbook.
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
3. **Decision trees.** "If `ovn-trace` allows but `ovs-appctl
   ofproto/trace` drops, suspect flow-installation and report restarting `ovn-controller` as remediation for an operator" is exactly what skills are for.
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
              | frontmatter +   |   | resource-get,       |
              | layered steps + |   | ovn-show, ovn-get,  |
              | decision tree   |   | ovn-trace,          |
              |                 |   | ovs-vsctl/ofctl,    |
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

## Initial skill catalogue

| Skill | When to use | Primary MCP tools |
|---|---|---|
| `pod-to-pod-connectivity`   | Pod A can't reach Pod B | resource-get, resource-list NetworkPolicy, ovn-show, ovn-get LSP/PB/Chassis/ACL/Port_Group, ovn-trace, ovn-lflow-list, ovs-appctl ofproto/trace, get-conntrack, tcpdump, pwru |
| `service-connectivity`      | ClusterIP/NodePort/LB unreachable, missing endpoints | resource-get Service/EndpointSlice, ovn-get Load_Balancer, ovn-trace, ovs-appctl dpctl/dump-conntrack, ovs-ofctl dump-flows, get-iptables, get-nft |
| `external-connectivity`     | North-south egress/ingress, SNAT/DNAT issues | ovn-get NAT/Logical_Router/Static_Route, ovn-trace, ovs-vsctl show, get-ip route, get-iptables nat, get-nft, tcpdump br-ex |
| `node-networking`           | NetworkUnavailable, tunnel down, missing routes | resource-get Node, get-ip address/link/route/rule/neigh, ovs-vsctl show, ovn-get Chassis, get-iptables, get-nft, pod-logs ovn-controller |

Skills are independent (pick one by symptom) and **composable** - a skill
**may** hand off to a sibling when the evidence points elsewhere. This
happens two ways: a skill can explicitly reference a sibling as a next
step, or, mid-execution, the agent may notice the symptoms now match a
different skill (for example, while running `service-connectivity` it
finds the node's tunnel to the backend is down, not a load-balancer
problem, and moves to `node-networking`). Which sibling, if any, is
loaded depends on what the evidence shows - it is not a fixed dependency
between skills.

# User Stories

**As an OVN-Kubernetes developer**, I want the AI assistant to follow the
layered triage methodology that I would follow (pod state → NB → SB → OVS → kernel →
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
```

A single canonical `skills/` is the source of truth (and what we will
serve as MCP resources later). Each client is pointed at this directory
through its own skills configuration.

## Skill format

```markdown
---
name: <kebab-case-skill-name>             # matches directory name
description: <symptom-oriented; must
  contain "Use when ..." trigger phrases>
---

# <Title>

## Prerequisites
- inputs to collect from the user

## Tools used
- list the MCP tools this skill calls

## Step 1 ... Step N
<imperative title>
<MCP tool call: tool name + parameters>
<what to extract; which fields feed which later step>

## Decision Tree
<branching summary; may hop to a sibling skill not already run>

## Reporting Findings
1. Root cause   2. Evidence   3. Remediation   4. Verification
```

**Author rules** (enforced by review in Phase-1, `skills-lint` in Phase-2):

1. Frontmatter present; `name` matches directory; `description` non-empty
   and contains "Use when ..." phrasing.
2. Every tool reference resolves to a registered MCP tool, every
   parameter is a real parameter of that tool.
3. In tool-call / executable steps: no free-form bash, ssh,
   `kubectl exec`, or write verbs (`set`, `add`, `del`, `remove`, ...).
   This applies only to executable steps; human-directed remediation
   prose (e.g. "Add an egress rule") is allowed.
4. Ends with the four-bullet reporting block.
5. ≤ 500 lines. Cross-reference sibling skills instead of duplicating.
6. Lists the MCP tools it requires (in a "Tools used" section) so the
   dependency on the server's tool surface is explicit.

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
  - decision tree -> ACL branch
  - ovn-get ACL/Port_Group                -> identifies the offending ACL
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

* In-tree under `skills/`, reviewed by an OVN-Kubernetes-mcp maintainer (or
  the layered-feature owner once feature skills land).
* MCP tool catalogue changes that affect skills must update both in the
  same PR; CI rejects PRs that leave a skill referencing a removed tool.

# Implementation Details

Phase-1 deliverables:

1. Author the initial set of `SKILL.md` files under canonical `skills/`
   (one directory per skill from the catalogue above). Add
   `skills/README.md` as the human-readable index + format reference.
2. Add a short docs page on pointing any Agent-Skills-compatible harness
   at the server and the `skills/` directory. Kept generic rather than
   per-agent; named harnesses (Cursor, Claude Code, ...) appear only as
   examples.
3. Add the eval prompts and breaking scenarios per skill (see Testing
   Strategy).
4. **Basic `skills-lint` (`make skills-lint`, wired into CI).**
   Spec-level well-formedness only: valid YAML frontmatter, `name`
   matching the directory, non-empty `description`, and a size cap.

In Phase-1 the CI gate therefore covers spec-level well-formedness only.
OVN-Kubernetes-specific correctness (right tools, right order) stays with
PR review, and behaviour with the contributor's eval run (see Testing
Strategy).

Phase-2 deliverables:

1. **Registry-backed `skills-lint`.** Extend the Phase-1 lint with the
   OVN-Kubernetes-specific checks: resolve every tool reference against
   the MCP server's tool registry, reject free-form bash and write verbs,
   and enforce the reporting block. Implement it in Go by reusing the
   server's own tool registrations as the source of truth so the lint and
   the server cannot disagree - renaming a tool then fails CI on every
   skill that references it, instead of breaking at runtime.
2. **Skills as MCP resources.** Add a resource provider exposing each skill
   at e.g. `mcp://ovn-kubernetes/skills/<name>` with `mimeType:
   text/markdown`. Additive; does not change the file format.

# Security Model

Skills inherit the MCP server's security model unchanged. They neither
expand nor relax it:

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

# Deployment Strategy

1. **Bundled with the repo (Phase 1, default).** Clone the repo and point
   an Agent-Skills-compatible client at the `skills/` directory via its
   skills configuration.
2. **MCP resources (Phase 2).** Same skills, served over the wire to
   any MCP client - no local checkout required.

We deliberately do **not** ship a separate skills package or container
in Phase 1 - that is release-management cost without a real distribution
problem.

**Configuration.** In Phase-2, operators selectively disable individual
skills via a `skills.disabled` list in the MCP server config, which hides
matching skills from the resource provider. Phase-1 has no config-driven toggle: a client is pointed at the
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
by CI. Creating an OVN-Kubernetes cluster and connecting the MCP server with the live-cluster tools (`ovn-trace`, `ovs-appctl ofproto/trace`, `tcpdump`, `pwru`, ...)
is impractical to run as an automated gate, and the result depends on the
backing LLM, which we do not pin. Instead:

* Each skill's eval prompts and the **reproducible breaking scenarios**
  they need are kept in-tree, so a reviewer can re-run them on their own
  cluster rather than take the author's word. This doubles as the
  regression suite as new breakages are found. Its exact shape is left to
  the implementation phase.
* The author runs the skill's eval prompts against a cluster they have
  access to (a disposable kind cluster is sufficient or a production
  cluster) and **pastes the eval summary into the PR**: which prompts
  triggered the skill, the tool-call sequence, the resulting RCA, and
  pass/fail per check.
* Reviewers use that summary as evidence, the same way they would review
  test output, alongside reading the skill itself.

Phase-1 gates spec-level well-formedness in CI via the basic
`skills-lint` (frontmatter, name/directory match, size cap); anything
OVN-Kubernetes-specific is checked in PR review, and behaviour by the
contributor's eval run. Phase-2 extends the lint with tool-reference
resolution against the registry, no free-form bash, and the reporting
block (Implementation Details). Either way, review / lint validates that
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

Each prompt records what it should produce - the skill that must trigger,
the tools that must appear, the expected root cause - and a run is graded
against that by a **judge**, either a reviewer or an LLM, checking whether
the triage report ticks every box.

# Documentation

End-user docs in `docs/skills/`:

* **Getting started** - how to point an Agent-Skills-compatible harness
  at the server and the skills folder, with one or two named examples.
* **Skill reference** - rendered list with frontmatter + tools used.
* **Authoring guide** - format, lint rules, review.
* **Cookbook** - 5-10 worked examples (paired with the eval prompt sets
  so users can replay).
* **FAQ / anti-patterns** - free-form bash, hardcoded parameters,
  cross-cluster assumptions, etc.

# Known Risks and Limitations

* **Client-coverage drift.** Agent Skills is real today in Cursor,
  Cursor CLI, Claude Code, and Anthropic's Skills API, but not yet
  universal. Clients that ignore `SKILL.md` get the server's raw tool
  catalogue and the failure modes that motivated this OKEP. Mitigation:
  Phase 2 serves skills via MCP resources, the most portable channel.
* **LLM still in the loop.** Skills constrain the workflow but do not
  eliminate hallucination. The reporting block forces evidence
  citations, which makes review easier but does not prevent errors.
* **Skill staleness.** As OVN-Kubernetes evolves, skills can drift.
  Mitigation: PR review catches semantic drift; the contributor's eval
  run catches behavioural drift; `skills-lint` catches mechanical
  breakage (bad frontmatter in Phase-1, stale tool references in Phase-2).
* **Over-trust.** Engineers may stop verifying RCAs because "the skill
  ran". Skills are first-pass triage, not authoritative; docs and the
  reporting block frame them that way.
* **Skill bias.** A skill encodes one path; real bugs occasionally need
  another. Mitigation: each skill ends with a decision tree pointing at
  sibling skills; the evidence-citation requirement makes a wrong-path
  conclusion detectable.
* **Maintenance cost.** Skills are documents reviewed per PR like code.
  Mitigation: PR review, `skills-lint` for mechanical breakage, and
  layered-feature ownership to spread the load.
* **Security carry-over.** All caveats of the underlying MCP server apply
  unchanged. Skills do not improve the underlying read-only guarantees;
  they orchestrate within them.
