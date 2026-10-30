// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"testing"

	remoteexec "k8s.io/client-go/util/exec"
)

func TestFormatCommandExecutionError(t *testing.T) {
	exitErr := remoteexec.CodeExitError{Err: errors.New("command terminated with exit code 1"), Code: 1}
	transportErr := errors.New("connection reset")
	tests := []struct {
		name      string
		stdout    string
		stderr    string
		cause     error
		wantError string
	}{
		{
			name: "successful command with output", stdout: "result\n", stderr: "warning\n",
		},
		{
			name: "both streams", stdout: "\npartial result\n", stderr: "\nexplanation\nsecond diagnostic\n", cause: exitErr,
			wantError: "command terminated with exit code 1\npartial stdout:\npartial result\nstderr:\nexplanation\nsecond diagnostic",
		},
		{
			name: "stderr only", stderr: "explanation\n", cause: exitErr,
			wantError: "command terminated with exit code 1\nstderr:\nexplanation",
		},
		{
			name: "stdout only", stdout: "partial result\n", cause: exitErr,
			wantError: "command terminated with exit code 1\npartial stdout:\npartial result",
		},
		{
			name: "empty output", cause: exitErr, wantError: exitErr.Error(),
		},
		{
			name: "whitespace output", stdout: " \n\t", stderr: " \n\t", cause: exitErr, wantError: exitErr.Error(),
		},
		{
			name: "transport failure", cause: transportErr, wantError: transportErr.Error(),
		},
		{
			name: "transport failure with output", stdout: "partial result\n", stderr: "connection interrupted\n", cause: transportErr,
			wantError: "connection reset\npartial stdout:\npartial result\nstderr:\nconnection interrupted",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FormatCommandExecutionError(tt.stdout, tt.stderr, tt.cause)
			if tt.cause == nil {
				if err != nil {
					t.Fatalf("successful command returned an error: %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantError {
				t.Fatalf("error = %v, want %q", err, tt.wantError)
			}
			if !errors.Is(err, tt.cause) {
				t.Errorf("error does not wrap original cause: %v", err)
			}
			var originalExit remoteexec.ExitError
			if errors.As(tt.cause, &originalExit) {
				var wrappedExit remoteexec.ExitError
				if !errors.As(err, &wrappedExit) || wrappedExit.ExitStatus() != originalExit.ExitStatus() {
					t.Errorf("original exit status is not accessible through errors.As: %v", err)
				}
			}
			if tt.wantError == tt.cause.Error() && err != tt.cause {
				t.Errorf("empty output changed the original error: %v", err)
			}
		})
	}
}
