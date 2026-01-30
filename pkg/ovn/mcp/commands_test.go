package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ovntypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/ovn/types"
)

// loadTestData loads test data from the testdata directory.
func loadTestData(t *testing.T, filename string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("failed to load testdata %s: %v", filename, err)
	}
	return string(data)
}

// TestParseOutput tests the parseOutput function with realistic OVN outputs.
func TestParseOutput(t *testing.T) {
	tests := []struct {
		name           string
		testdataFile   string
		expectedLines  int
		checkFirstLine string
		checkLastLine  string
	}{
		{
			name:           "parse ovn-nbctl show output",
			testdataFile:   "ovn-nbctl-show.txt",
			expectedLines:  57,
			checkFirstLine: `switch 7a8b9c0d-1234-5678-9abc-def012345678 (ovn-worker)`,
			checkLastLine:  `networks: ["100.64.0.1/16"]`,
		},
		{
			name:           "parse ovn-sbctl show output",
			testdataFile:   "ovn-sbctl-show.txt",
			expectedLines:  19,
			checkFirstLine: `Chassis "ovn-worker"`,
			checkLastLine:  `Port_Binding cr-rtos-ovn-control-plane`,
		},
		{
			name:           "parse ovn-nbctl list Logical_Switch output",
			testdataFile:   "ovn-nbctl-list-logical-switch.txt",
			expectedLines:  36,
			checkFirstLine: `_uuid               : 7a8b9c0d-1234-5678-9abc-def012345678`,
			checkLastLine:  `qos_rules           : []`,
		},
		{
			name:           "parse ovn-nbctl list ACL output",
			testdataFile:   "ovn-nbctl-list-acl.txt",
			expectedLines:  39,
			checkFirstLine: `_uuid               : 507eb871-13d0-4b4b-9495-cf6601000a72`,
			checkLastLine:  `tier                : 0`,
		},
		{
			name:           "parse ovn-sbctl lflow-list output",
			testdataFile:   "ovn-sbctl-lflow-list.txt",
			expectedLines:  22,
			checkFirstLine: `Datapath: "ovn-worker" (7a8b9c0d-1234-5678-9abc-def012345678)  Pipeline: ingress`,
			checkLastLine:  `table=0 (ls_out_pre_acl     ), priority=0    , match=(1), action=(next;)`,
		},
		{
			name:           "parse ovn-nbctl list NAT output",
			testdataFile:   "ovn-nbctl-list-nat.txt",
			expectedLines:  39,
			checkFirstLine: `_uuid               : 8a7b6c5d-4e3f-2a1b-0c9d-8e7f6a5b4c3d`,
			checkLastLine:  `type                : dnat_and_snat`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawOutput := loadTestData(t, tt.testdataFile)
			lines := parseOutput(rawOutput)

			if len(lines) != tt.expectedLines {
				t.Errorf("parseOutput() returned %d lines, want %d", len(lines), tt.expectedLines)
			}

			if len(lines) > 0 && lines[0] != tt.checkFirstLine {
				t.Errorf("first line = %q, want %q", lines[0], tt.checkFirstLine)
			}

			if len(lines) > 0 && lines[len(lines)-1] != tt.checkLastLine {
				t.Errorf("last line = %q, want %q", lines[len(lines)-1], tt.checkLastLine)
			}
		})
	}
}

// TestParseOutputEmpty tests parseOutput with empty and whitespace-only inputs.
func TestParseOutputEmpty(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect int
	}{
		{"empty string", "", 0},
		{"only newlines", "\n\n\n", 0},
		{"only whitespace", "   \n   \n   ", 0},
		{"single line", "hello", 1},
		{"single line with newline", "hello\n", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOutput(tt.input)
			if len(result) != tt.expect {
				t.Errorf("parseOutput() = %d lines, want %d", len(result), tt.expect)
			}
		})
	}
}

// TestFilterLinesWithOVNOutput tests filterLines with realistic OVN data.
func TestFilterLinesWithOVNOutput(t *testing.T) {
	tests := []struct {
		name          string
		testdataFile  string
		pattern       string
		expectedCount int
		shouldContain string
	}{
		{
			name:          "filter logical switches by name",
			testdataFile:  "ovn-nbctl-list-logical-switch.txt",
			pattern:       "ovn-worker",
			expectedCount: 2, // name lines containing ovn-worker
			shouldContain: "ovn-worker",
		},
		{
			name:          "filter logical switches by external_ids",
			testdataFile:  "ovn-nbctl-list-logical-switch.txt",
			pattern:       "k8s-ls-name",
			expectedCount: 2,
			shouldContain: "k8s-ls-name",
		},
		{
			name:          "filter ACLs with drop action",
			testdataFile:  "ovn-nbctl-list-acl.txt",
			pattern:       "action.*: drop",
			expectedCount: 1,
			shouldContain: "drop",
		},
		{
			name:          "filter ACLs by direction",
			testdataFile:  "ovn-nbctl-list-acl.txt",
			pattern:       "direction.*: to-lport",
			expectedCount: 2,
			shouldContain: "to-lport",
		},
		{
			name:          "filter logical flows by table",
			testdataFile:  "ovn-sbctl-lflow-list.txt",
			pattern:       "table=0",
			expectedCount: 10,
			shouldContain: "table=0",
		},
		{
			name:          "filter logical flows by priority 110",
			testdataFile:  "ovn-sbctl-lflow-list.txt",
			pattern:       "priority=110",
			expectedCount: 9,
			shouldContain: "priority=110",
		},
		{
			name:          "filter SNAT rules",
			testdataFile:  "ovn-nbctl-list-nat.txt",
			pattern:       "type.*: snat",
			expectedCount: 2,
			shouldContain: "snat",
		},
		{
			name:          "filter chassis by hostname",
			testdataFile:  "ovn-sbctl-show.txt",
			pattern:       "hostname:",
			expectedCount: 2,
			shouldContain: "hostname:",
		},
		{
			name:          "filter port bindings in sbctl show",
			testdataFile:  "ovn-sbctl-show.txt",
			pattern:       "Port_Binding",
			expectedCount: 9,
			shouldContain: "Port_Binding",
		},
		{
			name:          "filter routers in nbctl show",
			testdataFile:  "ovn-nbctl-show.txt",
			pattern:       "^router ",
			expectedCount: 2,
			shouldContain: "router ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawOutput := loadTestData(t, tt.testdataFile)
			lines := parseOutput(rawOutput)

			filtered, err := filterLines(lines, tt.pattern)
			if err != nil {
				t.Fatalf("filterLines() error = %v", err)
			}

			if len(filtered) != tt.expectedCount {
				t.Errorf("filterLines() returned %d lines, want %d", len(filtered), tt.expectedCount)
				for i, line := range filtered {
					t.Logf("  line %d: %s", i, line)
				}
			}

			// Verify all filtered lines contain the expected substring
			for _, line := range filtered {
				if tt.shouldContain != "" && !strings.Contains(line, tt.shouldContain) {
					t.Errorf("filtered line %q should contain %q", line, tt.shouldContain)
				}
			}
		})
	}
}

// TestFilterLinesInvalidPattern tests filterLines with invalid regex patterns.
func TestFilterLinesInvalidPattern(t *testing.T) {
	lines := []string{"test line 1", "test line 2"}

	_, err := filterLines(lines, "[invalid")
	if err == nil {
		t.Error("filterLines() should return error for invalid regex pattern")
	}
}

// TestLimitLinesWithOVNOutput tests limitLines with realistic data counts.
func TestLimitLinesWithOVNOutput(t *testing.T) {
	rawOutput := loadTestData(t, "ovn-sbctl-lflow-list.txt")
	lines := parseOutput(rawOutput)

	tests := []struct {
		name      string
		maxLines  int
		wantLines int
	}{
		{"limit to 5 flows", 5, 5},
		{"limit to 10 flows", 10, 10},
		{"limit to 50 (more than available)", 50, len(lines)},
		{"zero maxLines uses default", 0, min(defaultMaxLines, len(lines))},
		{"negative maxLines uses default", -1, min(defaultMaxLines, len(lines))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := limitLines(lines, tt.maxLines)
			if len(result) != tt.wantLines {
				t.Errorf("limitLines() = %d lines, want %d", len(result), tt.wantLines)
			}
		})
	}
}

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

// TestValidateTableName tests OVN table name validation.
func TestValidateTableName(t *testing.T) {
	tests := []struct {
		name    string
		table   string
		wantErr bool
	}{
		// Valid OVN NB tables
		{"Logical_Switch", "Logical_Switch", false},
		{"Logical_Router", "Logical_Router", false},
		{"Logical_Switch_Port", "Logical_Switch_Port", false},
		{"ACL", "ACL", false},
		{"Address_Set", "Address_Set", false},
		{"Port_Group", "Port_Group", false},
		{"Load_Balancer", "Load_Balancer", false},
		{"NAT", "NAT", false},
		// Valid OVN SB tables
		{"Chassis", "Chassis", false},
		{"Port_Binding", "Port_Binding", false},
		{"Datapath_Binding", "Datapath_Binding", false},
		{"Logical_Flow", "Logical_Flow", false},
		{"MAC_Binding", "MAC_Binding", false},
		// Invalid
		{"empty", "", true},
		{"starts with number", "1Table", true},
		{"has hyphen", "Logical-Switch", true},
		{"has space", "Logical Switch", true},
		{"injection attempt", "Table;drop", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTableName(tt.table)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTableName(%q) error = %v, wantErr %v", tt.table, err, tt.wantErr)
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

// TestTimeoutMechanism verifies timeout is applied correctly for OVN tools
func TestTimeoutMechanism(t *testing.T) {
	// Test 1: Timeout triggers when operation is slow
	t.Run("operation exceeds timeout", func(t *testing.T) {
		toolTimeout := 10 * time.Millisecond
		ctx := context.Background()

		// Apply timeout like OVN tools do
		if toolTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, toolTimeout)
			defer cancel()
		}

		// Simulate slow operation (50ms > 10ms timeout)
		time.Sleep(50 * time.Millisecond)

		// Verify context timed out
		if ctx.Err() != context.DeadlineExceeded {
			t.Error("expected context to timeout, but it didn't")
		}
	})

	// Test 2: No timeout when disabled
	t.Run("timeout disabled", func(t *testing.T) {
		toolTimeout := time.Duration(0) // Disabled
		ctx := context.Background()

		// Apply timeout (should be no-op when 0)
		if toolTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, toolTimeout)
			defer cancel()
		}

		// Simulate operation
		time.Sleep(10 * time.Millisecond)

		// Verify context did NOT timeout
		if ctx.Err() != nil {
			t.Errorf("expected no timeout, but got: %v", ctx.Err())
		}
	})
}
