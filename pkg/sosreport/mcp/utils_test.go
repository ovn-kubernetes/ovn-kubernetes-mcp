// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package sosreport

import (
	"os"
	"testing"
)

func TestReadLines(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test-readlines-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.Remove(tmpfile.Name())
		if err != nil {
			t.Fatalf("failed to remove temporary file %s: %v", tmpfile.Name(), err)
		}
	}()

	testContent := `line 1: ERROR occurred
line 2: INFO message
line 3: ERROR again
line 4: DEBUG info
line 5: ERROR third time
line 6: WARN warning
line 7: ERROR fourth
line 8: INFO another
line 9: ERROR fifth
line 10: DEBUG more
`
	if _, err := tmpfile.Write([]byte(testContent)); err != nil {
		t.Fatal(err)
	}
	err = tmpfile.Close()
	if err != nil {
		t.Fatalf("failed to close temporary file %s: %v", tmpfile.Name(), err)
	}

	tests := []struct {
		name      string
		wantLines int
		wantFirst string
		wantLast  string
	}{
		{
			name:      "reads all lines",
			wantLines: 10,
			wantFirst: "line 1: ERROR occurred",
			wantLast:  "line 10: DEBUG more",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.Open(tmpfile.Name())
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				err := file.Close()
				if err != nil {
					t.Errorf("failed to close file %s: %v", file.Name(), err)
				}
			}()

			lines, err := readLines(file)
			if err != nil {
				t.Errorf("readLines() unexpected error = %v", err)
				return
			}

			if len(lines) != tt.wantLines {
				t.Errorf("readLines() got %d lines, want %d", len(lines), tt.wantLines)
			}
			if len(lines) > 0 {
				if lines[0] != tt.wantFirst {
					t.Errorf("readLines() first line = %q, want %q", lines[0], tt.wantFirst)
				}
				if lines[len(lines)-1] != tt.wantLast {
					t.Errorf("readLines() last line = %q, want %q", lines[len(lines)-1], tt.wantLast)
				}
			}
		})
	}
}

func TestValidateRelativePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid simple path",
			path:    "sos_commands/openvswitch/ovs-vsctl_-t_5_show",
			wantErr: false,
		},
		{
			name:    "valid nested path",
			path:    "var/log/pods/namespace_pod/container/0.log",
			wantErr: false,
		},
		{
			name:    "invalid traversal with ..",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "invalid traversal in middle",
			path:    "sos_commands/../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "invalid absolute path",
			path:    "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "valid path with . current dir",
			path:    "./sos_commands/openvswitch/ovs-vsctl_-t_5_show",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRelativePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRelativePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
