package types

import (
	k8stypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
)

// BridgeResult contains the list of OVS bridges found on a node.
type BridgeResult struct {
	Bridges []string `json:"bridges"`
}

// ShowParams are the parameters for ovs-vsctl show command.
type ShowParams struct {
	k8stypes.NamespacedNameParams
	MaxLines int `json:"max_lines,omitempty"`
}

// ShowResult contains the output of ovs-vsctl show command.
type ShowResult struct {
	Output string `json:"output"`
}

// PortResult contains the list of ports attached to an OVS bridge.
type PortResult struct {
	Ports []string `json:"ports"`
}

// InterfaceResult contains the list of interfaces attached to an OVS bridge.
type InterfaceResult struct {
	Interfaces []string `json:"interfaces"`
}

// FlowsResult contains the OpenFlow flows from a specific OVS bridge.
type FlowsResult struct {
	Flows  []string `json:"flows"`
	Bridge string   `json:"bridge"`
}

// ConntrackResult contains connection tracking entries from the OVS datapath.
type ConntrackResult struct {
	Entries []string `json:"entries"`
}

// GetOVSCommandParams are the parameters for OVS related commands.
type GetOVSCommandParams struct {
	k8stypes.NamespacedNameParams
	Bridge   string `json:"bridge"`
	Filter   string `json:"filter,omitempty"`
	MaxLines int    `json:"max_lines,omitempty"`
}

// DumpConntrackParams are the parameters for dump-conntrack command.
type DumpConntrackParams struct {
	k8stypes.NamespacedNameParams
	Filter           string   `json:"filter,omitempty"`
	MaxLines         int      `json:"max_lines,omitempty"`
	AdditionalParams []string `json:"additional_params,omitempty"`
}

// OfprotoTraceParams are the parameters for ofproto/trace command.
type OfprotoTraceParams struct {
	k8stypes.NamespacedNameParams
	Bridge   string `json:"bridge"`
	Flow     string `json:"flow"`
	Filter   string `json:"filter,omitempty"`
	MaxLines int    `json:"max_lines,omitempty"`
}

// OfprotoTraceResult returns the complete trace output.
type OfprotoTraceResult struct {
	Bridge string `json:"bridge"`
	Flow   string `json:"flow"`
	Output string `json:"output"`
}
