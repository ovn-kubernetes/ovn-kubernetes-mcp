// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"errors"
	"testing"
	"time"

	remoteexec "k8s.io/client-go/util/exec"
)

func TestExecuteCommandPreservesFailureOutput(t *testing.T) {
	exitErr := remoteexec.CodeExitError{Err: errors.New("command terminated with exit code 1"), Code: 1}
	tests := []struct {
		name   string
		stdout string
		stderr string
		cause  error
	}{
		{"command failure with both streams", "partial output\n", "Warning: incomplete result\ncommand failed\n", exitErr},
		{"command failure with stdout only", "partial output\n", "", exitErr},
		{"command failure without output", "", "", exitErr},
		{"transport failure with partial output", "partial output\n", "connection interrupted\n", errors.New("unable to upgrade connection")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MCPServer{runDebugNodeCommand: func(_ context.Context, _, _, _ string, _ []string, _, _ string, _ time.Duration) (string, string, error) {
				return tt.stdout, tt.stderr, tt.cause
			}}
			stdout, stderr, err := s.executeCommand(t.Context(), "default", "node", []string{"ip", "link", "show"})
			if stdout != tt.stdout || stderr != tt.stderr {
				t.Errorf("output = (%q, %q), want (%q, %q)", stdout, stderr, tt.stdout, tt.stderr)
			}
			if !errors.Is(err, tt.cause) {
				t.Errorf("error = %v, want original cause %v", err, tt.cause)
			}
			var originalExit remoteexec.ExitError
			if errors.As(tt.cause, &originalExit) {
				var wrappedExit remoteexec.ExitError
				if !errors.As(err, &wrappedExit) || wrappedExit.ExitStatus() != originalExit.ExitStatus() {
					t.Errorf("error = %v, want original exit status accessible through errors.As", err)
				}
			}
		})
	}
}
