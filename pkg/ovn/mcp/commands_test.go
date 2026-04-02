package mcp

import (
	"testing"

	ovntypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/ovn/types"
)

// TestValidateDatabase tests database validation.
func TestValidateDatabase(t *testing.T) {
	tests := []struct {
		name    string
		db      ovntypes.Database
		wantErr bool
	}{
		{"valid nbdb", ovntypes.NorthboundDB, false},
		{"valid sbdb", ovntypes.SouthboundDB, false},
		{"invalid empty", ovntypes.Database(""), true},
		{"invalid unknown", ovntypes.Database("unknown"), true},
		{"invalid NBDB uppercase", ovntypes.Database("NBDB"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDatabase(tt.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDatabase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateRecordName tests OVN record identifier validation.
func TestValidateRecordName(t *testing.T) {
	tests := []struct {
		name    string
		record  string
		wantErr bool
	}{
		// Valid UUIDs (common in OVN)
		{"UUID format", "7a8b9c0d-1234-5678-9abc-def012345678", false},
		{"short UUID", "bee65224-bc1c-43fe-8afd-05f62ff150cc", false},
		// Valid names
		{"logical switch name", "ovn-worker", false},
		{"logical router name", "GR_ovn-worker", false},
		{"cluster router", "ovn_cluster_router", false},
		{"port name", "default_nginx-deployment-5d7f8b9c7d-abc12", false},
		// Invalid
		{"empty", "", true},
		{"semicolon injection", "record;drop", true},
		{"pipe injection", "record|cat", true},
		{"backtick injection", "record`whoami`", true},
		{"dollar injection", "record$HOME", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRecordName(tt.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRecordName(%q) error = %v, wantErr %v", tt.record, err, tt.wantErr)
			}
		})
	}
}

// TestValidateDatapath tests OVN datapath validation.
func TestValidateDatapath(t *testing.T) {
	tests := []struct {
		name     string
		datapath string
		wantErr  bool
	}{
		{"logical switch name", "ovn-worker", false},
		{"join switch", "join", false},
		{"external switch", "ext_ovn-worker", false},
		{"UUID", "7a8b9c0d-1234-5678-9abc-def012345678", false},
		{"empty", "", true},
		{"injection", "dp;rm -rf", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDatapath(tt.datapath)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDatapath(%q) error = %v, wantErr %v", tt.datapath, err, tt.wantErr)
			}
		})
	}
}

// TestValidateMicroflow tests OVN trace microflow validation.
func TestValidateMicroflow(t *testing.T) {
	tests := []struct {
		name      string
		microflow string
		wantErr   bool
	}{
		// Valid microflow specifications (from real ovn-trace usage)
		{
			name:      "basic inport with IP",
			microflow: `inport=="pod1" && eth.src==00:00:00:00:00:01 && ip4.src==10.244.0.5 && ip4.dst==10.244.1.5`,
			wantErr:   false,
		},
		{
			name:      "icmp trace",
			microflow: `inport=="pod1" && eth.src==00:00:00:00:00:01 && icmp && ip4.src==10.244.0.5 && ip4.dst==8.8.8.8`,
			wantErr:   false,
		},
		{
			name:      "tcp trace with port",
			microflow: `inport=="web-port" && eth.src==0a:58:0a:f4:01:05 && ip4.src==10.244.1.5 && ip4.dst==10.244.0.10 && tcp.dst==80`,
			wantErr:   false,
		},
		{
			name:      "egress traffic",
			microflow: `inport=="k8s-ovn-worker" && eth.dst==0a:58:0a:f4:01:01 && ip4.src==10.244.1.5 && ip4.dst==192.168.1.1`,
			wantErr:   false,
		},
		// Invalid
		{"empty", "", true},
		{"semicolon injection", `inport=="pod1";rm -rf /`, true},
		{"pipe injection", `inport=="pod1" | cat /etc/passwd`, true},
		{"backtick injection", "inport==`whoami`", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMicroflow(tt.microflow)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMicroflow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateColumnSpec tests column specification validation.
func TestValidateColumnSpec(t *testing.T) {
	tests := []struct {
		name    string
		columns string
		wantErr bool
	}{
		{"empty allowed", "", false},
		{"single column", "name", false},
		{"multiple columns", "name,_uuid,addresses", false},
		{"with underscores", "external_ids", false},
		{"injection attempt", "name;drop", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateColumnSpec(tt.columns)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateColumnSpec(%q) error = %v, wantErr %v", tt.columns, err, tt.wantErr)
			}
		})
	}
}

// TestGetDBCommand tests the database command selection.
func TestGetDBCommand(t *testing.T) {
	tests := []struct {
		name string
		db   ovntypes.Database
		want string
	}{
		{"nbdb returns ovn-nbctl", ovntypes.NorthboundDB, "ovn-nbctl"},
		{"sbdb returns ovn-sbctl", ovntypes.SouthboundDB, "ovn-sbctl"},
		{"unknown defaults to ovn-nbctl", ovntypes.Database("unknown"), "ovn-nbctl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDBCommand(tt.db)
			if got != tt.want {
				t.Errorf("getDBCommand(%q) = %q, want %q", tt.db, got, tt.want)
			}
		})
	}
}
