package sosreport

import (
	"strings"
	"testing"
)

const sosreportTestData = "testdata/sosreport"

func TestGetCommandOutput(t *testing.T) {
	tests := []struct {
		name         string
		sosreport    string
		filepath     string
		pattern      string
		maxLines     int
		wantError    bool
		errorMsg     string
		wantContains string
		wantMinLines int
		wantMaxLines int
	}{
		{
			name:         "get ovs-vsctl output without pattern",
			sosreport:    sosreportTestData,
			filepath:     "sos_commands/openvswitch/ovs-vsctl_-t_5_show",
			pattern:      "",
			maxLines:     0,
			wantError:    false,
			wantContains: "Bridge br-int",
			wantMinLines: 1,
		},
		{
			name:         "get ovs-ofctl output without pattern",
			sosreport:    sosreportTestData,
			filepath:     "sos_commands/openvswitch/ovs-ofctl_dump-flows_br-int",
			pattern:      "",
			maxLines:     0,
			wantError:    false,
			wantContains: "cookie=0x0",
			wantMinLines: 1,
		},
		{
			name:         "get ip addr show output",
			sosreport:    sosreportTestData,
			filepath:     "sos_commands/networking/ip_addr_show",
			pattern:      "",
			maxLines:     0,
			wantError:    false,
			wantContains: "lo:",
			wantMinLines: 1,
		},
		{
			name:         "filter with pattern no matches",
			sosreport:    sosreportTestData,
			filepath:     "sos_commands/networking/ip_addr_show",
			pattern:      "NOTFOUND",
			maxLines:     0,
			wantError:    false,
			wantContains: "No lines matching pattern",
			wantMinLines: 1,
		},
		{
			name:         "limit max lines",
			sosreport:    sosreportTestData,
			filepath:     "sos_commands/networking/ip_addr_show",
			pattern:      "",
			maxLines:     2,
			wantError:    false,
			wantContains: "output truncated",
			wantMaxLines: 2,
		},
		{
			name:      "non-existent file",
			sosreport: sosreportTestData,
			filepath:  "sos_commands/non-existent-file",
			pattern:   "",
			maxLines:  0,
			wantError: true,
			errorMsg:  "command output file not found",
		},
		{
			name:      "invalid sosreport path",
			sosreport: "testdata/non-existent",
			filepath:  "sos_commands/openvswitch/ovs-vsctl_-t_5_show",
			pattern:   "",
			maxLines:  0,
			wantError: true,
			errorMsg:  "sosreport path does not exist",
		},
		{
			name:      "invalid regex pattern",
			sosreport: sosreportTestData,
			filepath:  "sos_commands/networking/ip_addr_show",
			pattern:   "[invalid(",
			maxLines:  0,
			wantError: true,
			errorMsg:  "invalid pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := getCommandOutput(tt.sosreport, tt.filepath, tt.pattern, tt.maxLines)
			if tt.wantError {
				if err == nil {
					t.Errorf("getCommandOutput() expected error but got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("getCommandOutput() error = %v, want error containing %q", err, tt.errorMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("getCommandOutput() unexpected error = %v", err)
				return
			}

			// Check for expected content
			if tt.wantContains != "" && !strings.Contains(output, tt.wantContains) {
				t.Errorf("getCommandOutput() output does not contain %q, got:\n%s", tt.wantContains, output)
			}

			// Check line count if specified (excluding truncation message and empty lines)
			if tt.wantMinLines > 0 || tt.wantMaxLines > 0 {
				lines := strings.Split(output, "\n")
				lineCount := 0
				for _, line := range lines {
					if line != "" && !strings.Contains(line, "output truncated") && !strings.HasPrefix(line, "...") {
						lineCount++
					}
				}

				if tt.wantMinLines > 0 && lineCount < tt.wantMinLines {
					t.Errorf("getCommandOutput() got %d lines, want at least %d", lineCount, tt.wantMinLines)
				}

				if tt.wantMaxLines > 0 && lineCount > tt.wantMaxLines {
					t.Errorf("getCommandOutput() got %d lines, want at most %d. Output:\n%s", lineCount, tt.wantMaxLines, output)
				}
			}
		})
	}
}

func TestListPlugins(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		wantError       bool
		wantPluginCount int
		wantTotalCmds   int
	}{
		{
			name:            "valid sosreport",
			path:            sosreportTestData,
			wantError:       false,
			wantPluginCount: 3, // openvswitch, networking, container_log
			wantTotalCmds:   3, // 2 from openvswitch, 1 from networking, 0 from container_log
		},
		{
			name:      "non-existent path",
			path:      "testdata/non-existent",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := listPlugins(tt.path)
			if tt.wantError {
				if err == nil {
					t.Errorf("listPlugins() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("listPlugins() unexpected error = %v", err)
				return
			}

			if len(result.Plugins) != tt.wantPluginCount {
				t.Errorf("listPlugins() got %d plugins, want %d", len(result.Plugins), tt.wantPluginCount)
			}

			if result.TotalCommands != tt.wantTotalCmds {
				t.Errorf("listPlugins() got %d total commands, want %d", result.TotalCommands, tt.wantTotalCmds)
			}

			// Verify plugin names
			pluginNames := make(map[string]bool)
			for _, p := range result.Plugins {
				pluginNames[p.Name] = true
			}
			expectedPlugins := []string{"openvswitch", "networking", "container_log"}
			for _, expected := range expectedPlugins {
				if !pluginNames[expected] {
					t.Errorf("listPlugins() missing expected plugin %q", expected)
				}
			}
		})
	}
}

func TestListCommands(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		pluginName   string
		wantError    bool
		errorMsg     string
		wantCmdCount int
	}{
		{
			name:         "openvswitch plugin",
			path:         sosreportTestData,
			pluginName:   "openvswitch",
			wantError:    false,
			wantCmdCount: 2,
		},
		{
			name:         "networking plugin",
			path:         sosreportTestData,
			pluginName:   "networking",
			wantError:    false,
			wantCmdCount: 1,
		},
		{
			name:         "container_log plugin with no commands",
			path:         sosreportTestData,
			pluginName:   "container_log",
			wantError:    false,
			wantCmdCount: 0,
		},
		{
			name:       "non-existent plugin",
			path:       sosreportTestData,
			pluginName: "non-existent",
			wantError:  true,
			errorMsg:   "plugin \"non-existent\" not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := listCommands(tt.path, tt.pluginName)
			if tt.wantError {
				if err == nil {
					t.Errorf("listCommands() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("listCommands() unexpected error = %v", err)
				return
			}

			if result.Plugin != tt.pluginName {
				t.Errorf("listCommands() got plugin %q, want %q", result.Plugin, tt.pluginName)
			}

			if result.CommandCount != tt.wantCmdCount {
				t.Errorf("listCommands() got %d commands, want %d", result.CommandCount, tt.wantCmdCount)
			}

			if len(result.Commands) != tt.wantCmdCount {
				t.Errorf("listCommands() got %d command entries, want %d", len(result.Commands), tt.wantCmdCount)
			}

			// For plugins with commands, verify structure
			if tt.wantCmdCount > 0 {
				for _, cmd := range result.Commands {
					if cmd.Exec == "" {
						t.Errorf("listCommands() command has empty Exec field")
					}
					if cmd.Filepath == "" {
						t.Errorf("listCommands() command has empty Filepath field")
					}
				}
			}
		})
	}
}

func TestSearchCommands(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		pattern        string
		maxResults     int
		wantError      bool
		wantMinMatches int
		wantMaxMatches int
	}{
		{
			name:           "search for ovs commands",
			path:           sosreportTestData,
			pattern:        "ovs",
			maxResults:     100,
			wantError:      false,
			wantMinMatches: 2, // At least the 2 ovs commands
			wantMaxMatches: 100,
		},
		{
			name:           "search for ip command",
			path:           sosreportTestData,
			pattern:        "ip.*show",
			maxResults:     100,
			wantError:      false,
			wantMinMatches: 1, // ip addr show
			wantMaxMatches: 100,
		},
		{
			name:           "search with no matches",
			path:           sosreportTestData,
			pattern:        "nonexistent-pattern-xyz",
			maxResults:     100,
			wantError:      false,
			wantMinMatches: 0,
			wantMaxMatches: 0,
		},
		{
			name:       "invalid regex pattern",
			path:       sosreportTestData,
			pattern:    "[invalid(",
			maxResults: 100,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := searchCommands(tt.path, tt.pattern, tt.maxResults)
			if tt.wantError {
				if err == nil {
					t.Errorf("searchCommands() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("searchCommands() unexpected error = %v", err)
				return
			}

			if result.Total < tt.wantMinMatches {
				t.Errorf("searchCommands() got %d matches, want at least %d", result.Total, tt.wantMinMatches)
			}

			if result.Total > tt.wantMaxMatches {
				t.Errorf("searchCommands() got %d matches, want at most %d", result.Total, tt.wantMaxMatches)
			}

			if len(result.Matches) != result.Total {
				t.Errorf("searchCommands() matches slice length %d doesn't match Total %d", len(result.Matches), result.Total)
			}

			// Verify match structure
			for _, match := range result.Matches {
				if match.Plugin == "" {
					t.Errorf("searchCommands() match has empty Plugin field")
				}
				if match.Exec == "" {
					t.Errorf("searchCommands() match has empty Exec field")
				}
				if match.Filepath == "" {
					t.Errorf("searchCommands() match has empty Filepath field")
				}
			}
		})
	}
}
