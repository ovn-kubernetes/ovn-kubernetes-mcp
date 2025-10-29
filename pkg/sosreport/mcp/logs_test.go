package sosreport

import (
	"strings"
	"testing"
)

func TestSearchPodLogs(t *testing.T) {
	tests := []struct {
		name           string
		sosreport      string
		pattern        string
		podFilter      string
		maxResults     int
		wantError      bool
		errorMsg       string
		wantContains   string
		wantMinMatches int
		wantMaxMatches int
	}{
		{
			name:           "search for ERROR in logs",
			sosreport:      sosreportTestData,
			pattern:        "ERROR",
			podFilter:      "",
			maxResults:     100,
			wantError:      false,
			wantContains:   "ERROR: Failed to connect",
			wantMinMatches: 1,
		},
		{
			name:           "filter by pod name",
			sosreport:      sosreportTestData,
			pattern:        ".*",
			podFilter:      "ovnkube-node",
			maxResults:     100,
			wantError:      false,
			wantContains:   "ovnkube-node",
			wantMinMatches: 1,
		},
		{
			name:           "filter by pod with no match",
			sosreport:      sosreportTestData,
			pattern:        ".*",
			podFilter:      "non-existent-pod",
			maxResults:     100,
			wantError:      false,
			wantContains:   "No matches found",
			wantMinMatches: 0,
		},
		{
			name:           "search pattern with no matches",
			sosreport:      sosreportTestData,
			pattern:        "NOTFOUND",
			podFilter:      "",
			maxResults:     100,
			wantError:      false,
			wantContains:   "No matches found",
			wantMinMatches: 0,
		},
		{
			name:           "limit max results",
			sosreport:      sosreportTestData,
			pattern:        ".*",
			podFilter:      "",
			maxResults:     2,
			wantError:      false,
			wantContains:   "search truncated",
			wantMinMatches: 1,
			wantMaxMatches: 2,
		},
		{
			name:       "invalid regex pattern",
			sosreport:  sosreportTestData,
			pattern:    "[invalid(",
			podFilter:  "",
			maxResults: 100,
			wantError:  true,
			errorMsg:   "invalid search pattern",
		},
		{
			name:       "invalid sosreport path",
			sosreport:  "testdata/non-existent",
			pattern:    "ERROR",
			podFilter:  "",
			maxResults: 100,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := searchPodLogs(tt.sosreport, tt.pattern, tt.podFilter, tt.maxResults)
			if tt.wantError {
				if err == nil {
					t.Errorf("searchPodLogs() expected error but got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("searchPodLogs() error = %v, want error containing %q", err, tt.errorMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("searchPodLogs() unexpected error = %v", err)
				return
			}

			// Check for expected content
			if tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
				t.Errorf("searchPodLogs() result does not contain %q, got:\n%s", tt.wantContains, result)
			}

			// For cases with no matches, we should get the "No matches found" message
			if tt.wantMinMatches == 0 && !strings.Contains(result, "No matches found") {
				t.Errorf("searchPodLogs() expected 'No matches found' message but didn't get it, got:\n%s", result)
			}

			// For cases with matches, we shouldn't have the "No matches found" message
			if tt.wantMinMatches > 0 && strings.Contains(result, "No matches found") {
				t.Errorf("searchPodLogs() unexpected 'No matches found' message when matches were expected")
			}

			// Check for truncation message when max results is set
			if tt.maxResults > 0 && tt.wantMaxMatches > 0 && tt.wantMinMatches > 0 {
				if !strings.Contains(result, "search truncated") && len(strings.Split(result, "\n")) > tt.maxResults {
					t.Errorf("searchPodLogs() expected truncation message when max results exceeded")
				}
			}
		})
	}
}
