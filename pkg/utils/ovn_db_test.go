package utils

import "testing"

// TestValidateOVNTableName tests OVN table name validation.
func TestValidateOVNTableName(t *testing.T) {
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
			err := ValidateOVNTableName(tt.table)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOVNTableName(%q) error = %v, wantErr %v", tt.table, err, tt.wantErr)
			}
		})
	}
}
