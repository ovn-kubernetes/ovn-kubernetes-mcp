// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const maxCommandStderrBytes = 16 * 1024

// FormatCommandExecutionError adds nonblank command output to an execution error,
// preserving the original cause. Callers choose which streams to include and
// apply stdout limits before calling. Stderr is capped at 16 KiB with a truncation
// marker. It returns nil when err is nil.
func FormatCommandExecutionError(stdout, stderr string, err error) error {
	if err == nil {
		return nil
	}
	if output := strings.TrimSpace(stdout); output != "" {
		err = fmt.Errorf("%w\npartial stdout:\n%s", err, output)
	}
	if message := strings.TrimSpace(stderr); message != "" {
		if len(message) > maxCommandStderrBytes {
			end := maxCommandStderrBytes
			for end > 0 && !utf8.RuneStart(message[end]) {
				end--
			}
			message = message[:end] + "\n[stderr truncated]"
		}
		err = fmt.Errorf("%w\nstderr:\n%s", err, message)
	}
	return err
}
