// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	ovntypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/ovn/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils/headtail"
)

func TestShowFailureOutputLimits(t *testing.T) {
	cause := errors.New("command terminated with exit code 1")
	stdout := strings.Repeat("first\n", DefaultMaxLines) + "last\n"
	tests := []struct {
		name   string
		limits headtail.HeadTailParams
		want   string
	}{
		{"default limit", headtail.HeadTailParams{}, strings.TrimSpace(strings.Repeat("first\n", DefaultMaxLines))},
		{"head", headtail.HeadTailParams{Head: 1}, "first"},
		{"tail", headtail.HeadTailParams{Tail: 1}, "last"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MCPServer{runPodExecCommand: func(_ context.Context, _, _, _ string, _ []string) (string, string, error) {
				return stdout, "native explanation\n", cause
			}}
			_, _, err := s.Show(t.Context(), nil, ovntypes.ShowParams{Database: "nbdb", HeadTailParams: tt.limits})
			if !errors.Is(err, cause) {
				t.Fatalf("error = %v, want original cause %v", err, cause)
			}
			want := "\npartial stdout:\n" + tt.want + "\nstderr:\nnative explanation"
			if !strings.HasSuffix(err.Error(), want) {
				t.Errorf("error = %v, want limited stdout and stderr suffix %q", err, want)
			}
		})
	}
}

func TestTraceCommand(t *testing.T) {
	tests := []struct {
		name string
		mode ovntypes.TraceMode
		flag string
	}{
		{"default", "", "--detailed"},
		{"detailed", ovntypes.TraceModeDetailed, "--detailed"},
		{"summary", ovntypes.TraceModeSummary, "--summary"},
		{"minimal", ovntypes.TraceModeMinimal, "--minimal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var command []string
			s := &MCPServer{runPodExecCommand: func(_ context.Context, _, _, _ string, cmd []string) (string, string, error) {
				command = cmd
				return "trace output", "", nil
			}}
			in := ovntypes.OVNTraceParams{
				Datapath:  "ovn-worker",
				Microflow: `(inport=="web-port") && ip6.src==2001:db8::1`,
				Mode:      tt.mode,
			}
			if _, _, err := s.Trace(t.Context(), nil, in); err != nil {
				t.Fatal(err)
			}
			want := []string{"ovn-trace", tt.flag, "--", in.Datapath, in.Microflow}
			if !slices.Equal(command, want) {
				t.Errorf("command = %q, want %q", command, want)
			}
		})
	}
}

func TestTraceRejectsOptionOperands(t *testing.T) {
	tests := []struct {
		name      string
		datapath  string
		microflow string
	}{
		{"datapath database option", "--db=tcp:198.51.100.7:6642", `inport=="pod1"`},
		{"microflow database option", "join", "--db=tcp:198.51.100.7:6642"},
		{"datapath option after whitespace", " \t--db=tcp:198.51.100.7:6642", `inport=="pod1"`},
		{"microflow option after whitespace", "join", " \t--db=tcp:198.51.100.7:6642"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MCPServer{runPodExecCommand: func(_ context.Context, _, _, _ string, _ []string) (string, string, error) {
				t.Fatal("option-like operand reached pod execution")
				return "", "", nil
			}}
			_, _, err := s.Trace(t.Context(), nil, ovntypes.OVNTraceParams{
				Datapath:  tt.datapath,
				Microflow: tt.microflow,
			})
			if err == nil || !strings.Contains(err.Error(), "must not start with '-'") {
				t.Errorf("error = %v, want leading-hyphen validation error", err)
			}
		})
	}
}

func TestListLogicalFlowsDatapath(t *testing.T) {
	tests := []struct {
		name     string
		datapath string
		want     []string
	}{
		{"all datapaths", "", []string{"ovn-sbctl", "lflow-list"}},
		{"named datapath", "ovn-worker", []string{"ovn-sbctl", "lflow-list", "ovn-worker"}},
		{"database option", "--db=tcp:198.51.100.7:6642", nil},
		{"option separator", "--", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var command []string
			s := &MCPServer{runPodExecCommand: func(_ context.Context, _, _, _ string, cmd []string) (string, string, error) {
				command = cmd
				return "logical flows", "", nil
			}}
			_, _, err := s.ListLogicalFlows(t.Context(), nil, ovntypes.LogicalFlowListParams{Datapath: tt.datapath})
			if tt.want == nil {
				if err == nil || !strings.Contains(err.Error(), "must not start with '-'") {
					t.Errorf("error = %v, want leading-hyphen validation error", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(command, tt.want) {
				t.Errorf("command = %q, want %q", command, tt.want)
			}
		})
	}
}
