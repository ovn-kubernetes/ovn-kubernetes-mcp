package kubernetes

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
)

// validatePath validates that a path is safe to use for mounting.
// It ensures the path:
// - Is absolute (starts with /)
// - Does not contain path traversal patterns (..)
// - Contains only safe characters
func validatePath(path, pathType string) error {
	if path == "" {
		return nil // Empty paths are allowed and will be set to defaults
	}

	// Ensure path is absolute
	if !filepath.IsAbs(path) {
		return fmt.Errorf("%s must be an absolute path (start with /), got: %s", pathType, path)
	}

	// Check for path traversal patterns: reject any path element that is exactly ".."
	for _, elem := range strings.Split(path, string(filepath.Separator)) {
		if elem == ".." {
			return fmt.Errorf("%s contains path traversal element '..': %s", pathType, path)
		}
	}

	// Check for dangerous characters (null bytes, control characters, shell special characters)
	for i, r := range path {
		// Allow alphanumeric, /, -, _, ., ~
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '/' || r == '-' || r == '_' || r == '.' || r == '~' {
			continue
		}
		return fmt.Errorf("%s contains unsafe character at position %d: %c (U+%04X)", pathType, i, r, r)
	}

	return nil
}

// DebugNode debugs a node by name, image and command.
func (s *MCPServer) DebugNode(ctx context.Context, req *mcp.CallToolRequest, in types.DebugNodeParams) (*mcp.CallToolResult, types.DebugNodeResult, error) {
	// Validate paths before creating the pod
	if err := validatePath(in.HostPath, "hostPath"); err != nil {
		return nil, types.DebugNodeResult{}, err
	}

	if err := validatePath(in.MountPath, "mountPath"); err != nil {
		return nil, types.DebugNodeResult{}, err
	}

	stdout, stderr, err := s.clientSet.DebugNode(ctx, in.Name, in.Image, in.Command, in.HostPath, in.MountPath)
	if err != nil {
		return nil, types.DebugNodeResult{}, err
	}

	return nil, types.DebugNodeResult{Stdout: stdout, Stderr: stderr}, nil
}
