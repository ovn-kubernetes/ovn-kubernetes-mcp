// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package sosreport

import (
	"strings"
	"testing"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils/headtail"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils/pattern"
)

func TestGetPodLogs(t *testing.T) {
	tests := []struct {
		name           string
		sosreport      string
		namespace      string
		podName        string
		container      string
		pattern        string
		head           int
		tail           int
		wantError      bool
		errorMsg       string
		wantContains   string
		wantNotContain string
		wantMinMatches int
		wantMaxLines   int
	}{
		{
			name:           "get first container when container omitted",
			sosreport:      sosreportTestData,
			namespace:      "openshift-ovn-kubernetes",
			podName:        "ovnkube-node-abc",
			wantContains:   "Starting ovnkube-controller",
			wantNotContain: "northd",
			wantMinMatches: 1,
		},
		{
			name:           "get specific container",
			sosreport:      sosreportTestData,
			namespace:      "openshift-ovn-kubernetes",
			podName:        "ovnkube-node-abc",
			container:      "northd",
			wantContains:   "Starting northd container",
			wantNotContain: "Starting ovnkube-controller",
			wantMinMatches: 1,
		},
		{
			name:           "filter with pattern",
			sosreport:      sosreportTestData,
			namespace:      "openshift-ovn-kubernetes",
			podName:        "ovnkube-node-abc",
			container:      "ovnkube-controller",
			pattern:        "ERROR",
			wantContains:   "ERROR: Failed to connect",
			wantMinMatches: 1,
		},
		{
			name:           "pod not found",
			sosreport:      sosreportTestData,
			namespace:      "openshift-ovn-kubernetes",
			podName:        "non-existent-pod",
			wantContains:   "No pod logs found for pod",
			wantMinMatches: 0,
		},
		{
			name:           "container not found",
			sosreport:      sosreportTestData,
			namespace:      "openshift-ovn-kubernetes",
			podName:        "ovnkube-node-abc",
			container:      "missing-container",
			wantContains:   "No pod logs found for pod",
			wantMinMatches: 0,
		},
		{
			name:           "pattern with no matches",
			sosreport:      sosreportTestData,
			namespace:      "openshift-ovn-kubernetes",
			podName:        "ovnkube-node-abc",
			container:      "ovnkube-controller",
			pattern:        "NOTFOUND",
			wantContains:   "No matches found",
			wantMinMatches: 0,
		},
		{
			name:           "does not match other pods",
			sosreport:      sosreportTestData,
			namespace:      "openshift-ovn-kubernetes",
			podName:        "ovnkube-node-abc",
			pattern:        "ERROR",
			wantContains:   "ERROR: Failed to connect",
			wantNotContain: "Master failed cluster sync",
			wantMinMatches: 1,
		},
		{
			name:           "limit with head",
			sosreport:      sosreportTestData,
			namespace:      "openshift-ovn-kubernetes",
			podName:        "ovnkube-node-abc",
			head:           2,
			wantMinMatches: 1,
			wantMaxLines:   2,
		},
		{
			name:      "missing namespace",
			sosreport: sosreportTestData,
			namespace: "",
			podName:   "ovnkube-node-abc",
			wantError: true,
			errorMsg:  "namespace is required",
		},
		{
			name:      "missing name",
			sosreport: sosreportTestData,
			namespace: "openshift-ovn-kubernetes",
			podName:   "",
			wantError: true,
			errorMsg:  "name is required",
		},
		{
			name:      "invalid regex pattern",
			sosreport: sosreportTestData,
			namespace: "openshift-ovn-kubernetes",
			podName:   "ovnkube-node-abc",
			pattern:   "[invalid(",
			wantError: true,
			errorMsg:  "invalid search pattern",
		},
		{
			name:      "invalid sosreport path",
			sosreport: "testdata/non-existent",
			namespace: "openshift-ovn-kubernetes",
			podName:   "ovnkube-node-abc",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getPodLogs(tt.sosreport, tt.namespace, tt.podName, tt.container,
				pattern.PatternParams{Pattern: tt.pattern},
				headtail.HeadTailParams{Head: tt.head, Tail: tt.tail})
			if tt.wantError {
				if err == nil {
					t.Errorf("getPodLogs() expected error but got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("getPodLogs() error = %v, want error containing %q", err, tt.errorMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("getPodLogs() unexpected error = %v", err)
				return
			}

			if tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
				t.Errorf("getPodLogs() result does not contain %q, got:\n%s", tt.wantContains, result)
			}

			if tt.wantNotContain != "" && strings.Contains(result, tt.wantNotContain) {
				t.Errorf("getPodLogs() result unexpectedly contains %q, got:\n%s", tt.wantNotContain, result)
			}

			if tt.wantMinMatches == 0 && !strings.Contains(result, "No matches found") && !strings.Contains(result, "No pod logs found for pod") {
				t.Errorf("getPodLogs() expected no-match message but didn't get it, got:\n%s", result)
			}

			if tt.wantMinMatches > 0 && (strings.Contains(result, "No matches found") || strings.Contains(result, "No pod logs found for pod")) {
				t.Errorf("getPodLogs() unexpected no-match message when matches were expected")
			}

			if tt.wantMaxLines > 0 {
				lines := strings.Split(result, "\n")
				lineCount := 0
				for _, line := range lines {
					if line != "" {
						lineCount++
					}
				}
				if lineCount > tt.wantMaxLines {
					t.Errorf("getPodLogs() got %d lines, want at most %d. Result:\n%s", lineCount, tt.wantMaxLines, result)
				}
			}
		})
	}
}

func TestMatchesContainerLogFile(t *testing.T) {
	tests := []struct {
		name      string
		logPath   string
		namespace string
		podName   string
		container string
		want      bool
	}{
		{
			name:      "matches first container without container filter",
			logPath:   "host/var/log/containers/ovnkube-node-abc_openshift-ovn-kubernetes_ovnkube-controller-deadbeef01.log",
			namespace: "openshift-ovn-kubernetes",
			podName:   "ovnkube-node-abc",
			want:      true,
		},
		{
			name:      "matches specific container",
			logPath:   "host/var/log/containers/ovnkube-node-abc_openshift-ovn-kubernetes_northd-cafebabe02.log",
			namespace: "openshift-ovn-kubernetes",
			podName:   "ovnkube-node-abc",
			container: "northd",
			want:      true,
		},
		{
			name:      "rejects wrong container",
			logPath:   "host/var/log/containers/ovnkube-node-abc_openshift-ovn-kubernetes_northd-cafebabe02.log",
			namespace: "openshift-ovn-kubernetes",
			podName:   "ovnkube-node-abc",
			container: "ovnkube-controller",
			want:      false,
		},
		{
			name:      "rejects container name prefix",
			logPath:   "host/var/log/containers/ovnkube-node-abc_openshift-ovn-kubernetes_ovnkube-controller-deadbeef01.log",
			namespace: "openshift-ovn-kubernetes",
			podName:   "ovnkube-node-abc",
			container: "ovnkube",
			want:      false,
		},
		{
			name:      "rejects different pod",
			logPath:   "host/var/log/containers/ovnkube-control-plane-xyz_openshift-ovn-kubernetes_ovnkube-cluster-manager-feedface03.log",
			namespace: "openshift-ovn-kubernetes",
			podName:   "ovnkube-node-abc",
			want:      false,
		},
		{
			name:      "supports gzip suffix",
			logPath:   "var/log/containers/ovnkube-node-abc_openshift-ovn-kubernetes_ovnkube-controller-deadbeef01.log.gz",
			namespace: "openshift-ovn-kubernetes",
			podName:   "ovnkube-node-abc",
			container: "ovnkube-controller",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesContainerLogFile(tt.logPath, tt.namespace, tt.podName, tt.container)
			if got != tt.want {
				t.Errorf("matchesContainerLogFile(...) = %v, want %v", got, tt.want)
			}
		})
	}
}
