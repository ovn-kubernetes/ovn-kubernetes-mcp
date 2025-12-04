package sosreport

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestReadWithLimit(t *testing.T) {
	// Create a temporary test file
	tmpfile, err := os.CreateTemp("", "test-readwithlimit-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write test content
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
	tmpfile.Close()

	tests := []struct {
		name        string
		pattern     string
		maxLines    int
		wantLines   int
		wantPattern string
		wantTrunc   bool
	}{
		{
			name:      "no pattern, no limit",
			pattern:   "",
			maxLines:  0,
			wantLines: 10,
			wantTrunc: false,
		},
		{
			name:      "no pattern, with limit",
			pattern:   "",
			maxLines:  5,
			wantLines: 5,
			wantTrunc: true,
		},
		{
			name:        "with pattern, no limit",
			pattern:     "ERROR",
			maxLines:    0,
			wantLines:   5,
			wantPattern: "ERROR",
			wantTrunc:   false,
		},
		{
			name:        "with pattern, with limit",
			pattern:     "ERROR",
			maxLines:    3,
			wantLines:   3,
			wantPattern: "ERROR",
			wantTrunc:   true,
		},
		{
			name:        "pattern with no matches returns empty",
			pattern:     "NOTFOUND",
			maxLines:    0,
			wantLines:   0,
			wantPattern: "NOTFOUND",
			wantTrunc:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.Open(tmpfile.Name())
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			var searchPattern *regexp.Regexp
			if tt.pattern != "" {
				searchPattern = regexp.MustCompile(tt.pattern)
			}

			result, err := readWithLimit(file, searchPattern, tt.maxLines)
			if err != nil {
				t.Errorf("readWithLimit() unexpected error = %v", err)
				return
			}

			// Count lines (excluding truncation message and empty lines)
			lines := strings.Split(result, "\n")
			actualLines := 0
			for _, line := range lines {
				if line != "" && !strings.Contains(line, "output truncated") && !strings.HasPrefix(line, "...") {
					actualLines++
				}
			}

			// If wantLines is 0, we expect empty string
			if tt.wantLines == 0 && result != "" {
				t.Errorf("readWithLimit() expected empty result but got %q", result)
				return
			}

			if tt.wantLines > 0 && actualLines != tt.wantLines {
				t.Errorf("readWithLimit() got %d lines, want %d lines. Result:\n%s", actualLines, tt.wantLines, result)
			}

			// Check if all lines match the pattern (if pattern is specified)
			if tt.wantPattern != "" && tt.wantLines > 0 {
				for _, line := range lines {
					if line != "" && !strings.Contains(line, "output truncated") && !strings.Contains(line, tt.wantPattern) {
						t.Errorf("readWithLimit() line %q does not contain pattern %q", line, tt.wantPattern)
					}
				}
			}

			// Check for truncation message
			hasTruncMsg := strings.Contains(result, "output truncated")
			if tt.wantTrunc && !hasTruncMsg {
				t.Errorf("readWithLimit() expected truncation message but didn't find it")
			}
			if !tt.wantTrunc && hasTruncMsg {
				t.Errorf("readWithLimit() unexpected truncation message")
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
