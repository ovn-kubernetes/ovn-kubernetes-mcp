package types

import (
	k8stypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
)

// Database represents an OVN database type.
type Database string

const (
	// NorthboundDB is the OVN Northbound database.
	NorthboundDB Database = "nbdb"
	// SouthboundDB is the OVN Southbound database.
	SouthboundDB Database = "sbdb"
)

// ShowParams are the parameters for ovn-nbctl/ovn-sbctl show command.
type ShowParams struct {
	k8stypes.NamespacedNameParams
	Database Database `json:"database"`
	MaxLines int      `json:"max_lines,omitempty"`
}

// ShowResult contains the output of ovn-nbctl/ovn-sbctl show command.
type ShowResult struct {
	Database Database `json:"database"`
	Output   string   `json:"output"`
}

// ListTableParams are the parameters for listing records in an OVN table.
type ListTableParams struct {
	k8stypes.NamespacedNameParams
	Database Database `json:"database"`
	Table    string   `json:"table"`
	Columns  string   `json:"columns,omitempty"`
	Filter   string   `json:"filter,omitempty"`
	MaxLines int      `json:"max_lines,omitempty"`
}

// ListTableResult contains records from an OVN table.
type ListTableResult struct {
	Database Database `json:"database"`
	Table    string   `json:"table"`
	Records  []string `json:"records"`
}

// LogicalSwitchListParams are the parameters for listing logical switches.
type LogicalSwitchListParams struct {
	k8stypes.NamespacedNameParams
	Filter   string `json:"filter,omitempty"`
	MaxLines int    `json:"max_lines,omitempty"`
}

// LogicalSwitchListResult contains the list of logical switches.
type LogicalSwitchListResult struct {
	Switches []string `json:"switches"`
}

// LogicalRouterListParams are the parameters for listing logical routers.
type LogicalRouterListParams struct {
	k8stypes.NamespacedNameParams
	Filter   string `json:"filter,omitempty"`
	MaxLines int    `json:"max_lines,omitempty"`
}

// LogicalRouterListResult contains the list of logical routers.
type LogicalRouterListResult struct {
	Routers []string `json:"routers"`
}

// LogicalSwitchPortListParams are the parameters for listing logical switch ports.
type LogicalSwitchPortListParams struct {
	k8stypes.NamespacedNameParams
	Switch   string `json:"switch,omitempty"`
	Filter   string `json:"filter,omitempty"`
	MaxLines int    `json:"max_lines,omitempty"`
}

// LogicalSwitchPortListResult contains the list of logical switch ports.
type LogicalSwitchPortListResult struct {
	Switch string   `json:"switch,omitempty"`
	Ports  []string `json:"ports"`
}

// ACLListParams are the parameters for listing ACLs.
type ACLListParams struct {
	k8stypes.NamespacedNameParams
	Entity   string `json:"entity"` // Logical switch or port group name
	Filter   string `json:"filter,omitempty"`
	MaxLines int    `json:"max_lines,omitempty"`
}

// ACLListResult contains the list of ACLs.
type ACLListResult struct {
	Entity string   `json:"entity,omitempty"` // Logical switch or port group name
	ACLs   []string `json:"acls"`
}

// LogicalFlowListParams are the parameters for listing logical flows from SBDB.
type LogicalFlowListParams struct {
	k8stypes.NamespacedNameParams
	Datapath string `json:"datapath,omitempty"`
	Filter   string `json:"filter,omitempty"`
	MaxLines int    `json:"max_lines,omitempty"`
}

// LogicalFlowListResult contains the list of logical flows.
type LogicalFlowListResult struct {
	Datapath string   `json:"datapath,omitempty"`
	Flows    []string `json:"flows"`
}

// TraceMode represents the output verbosity mode for ovn-trace.
type TraceMode string

const (
	// TraceModeDetailed shows detailed trace output (default).
	TraceModeDetailed TraceMode = "detailed"
	// TraceModeSummary shows summary output only.
	TraceModeSummary TraceMode = "summary"
	// TraceModeMinimal shows minimal output.
	TraceModeMinimal TraceMode = "minimal"
)

// OVNTraceParams are the parameters for ovn-trace command.
type OVNTraceParams struct {
	k8stypes.NamespacedNameParams
	Datapath  string    `json:"datapath"`
	Microflow string    `json:"microflow"`
	Mode      TraceMode `json:"mode,omitempty"` // Output mode: detailed (default), summary, or minimal
	Filter    string    `json:"filter,omitempty"`
	MaxLines  int       `json:"max_lines,omitempty"`
}

// OVNTraceResult contains the output of ovn-trace command.
type OVNTraceResult struct {
	Datapath  string `json:"datapath"`
	Microflow string `json:"microflow"`
	Output    string `json:"output"`
}

// ChassisListParams are the parameters for listing chassis from SBDB.
type ChassisListParams struct {
	k8stypes.NamespacedNameParams
	Filter   string `json:"filter,omitempty"`
	MaxLines int    `json:"max_lines,omitempty"`
}

// ChassisListResult contains the list of chassis.
type ChassisListResult struct {
	Chassis []string `json:"chassis"`
}

// PortBindingListParams are the parameters for listing port bindings from SBDB.
type PortBindingListParams struct {
	k8stypes.NamespacedNameParams
	Filter   string `json:"filter,omitempty"`
	MaxLines int    `json:"max_lines,omitempty"`
}

// PortBindingListResult contains the list of port bindings.
type PortBindingListResult struct {
	Bindings []string `json:"bindings"`
}

// GetParams are the parameters for getting a specific record from an OVN table.
type GetParams struct {
	k8stypes.NamespacedNameParams
	Database Database `json:"database"`
	Table    string   `json:"table"`
	Record   string   `json:"record"`
	Column   string   `json:"column,omitempty"`
}

// GetResult contains the output of ovn-nbctl/ovn-sbctl get command.
type GetResult struct {
	Database Database `json:"database"`
	Table    string   `json:"table"`
	Record   string   `json:"record"`
	Output   string   `json:"output"`
}

// PortGroupListParams are the parameters for listing port groups.
type PortGroupListParams struct {
	k8stypes.NamespacedNameParams
	Filter   string `json:"filter,omitempty"`
	MaxLines int    `json:"max_lines,omitempty"`
}

// PortGroupListResult contains the list of port groups.
type PortGroupListResult struct {
	PortGroups []string `json:"port_groups"`
}

// NATListParams are the parameters for listing NAT rules.
type NATListParams struct {
	k8stypes.NamespacedNameParams
	Router   string `json:"router,omitempty"`
	Filter   string `json:"filter,omitempty"`
	MaxLines int    `json:"max_lines,omitempty"`
}

// NATListResult contains the list of NAT rules.
type NATListResult struct {
	Router string   `json:"router,omitempty"`
	NATs   []string `json:"nats"`
}

// AddressSetListParams are the parameters for listing address sets.
type AddressSetListParams struct {
	k8stypes.NamespacedNameParams
	Filter   string `json:"filter,omitempty"`
	MaxLines int    `json:"max_lines,omitempty"`
}

// AddressSetListResult contains the list of address sets.
type AddressSetListResult struct {
	AddressSets []string `json:"address_sets"`
}
