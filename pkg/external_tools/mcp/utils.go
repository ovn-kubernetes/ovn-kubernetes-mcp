package external_tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/client"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/types"
)

// executeAndWrapResult is a helper that executes a command and wraps the result
// in the standard format, handling errors consistently
func (s *MCPServer) executeAndWrapResult(ctx context.Context, target types.TargetParams, command []string) (*mcp.CallToolResult, types.CommandResult, error) {
	output, err := s.executeCommand(ctx, target, command)
	if err != nil {
		return nil, types.CommandResult{Output: output, Error: err.Error()}, err
	}
	return nil, types.CommandResult{Output: output}, nil
}

// commandBuilder helps build commands with a fluent interface
type commandBuilder struct {
	args []string
}

// newCommand creates a new command builder with the base command
func newCommand(baseCmd ...string) *commandBuilder {
	return &commandBuilder{args: baseCmd}
}

// add adds arguments to the command
func (cb *commandBuilder) add(args ...string) *commandBuilder {
	cb.args = append(cb.args, args...)
	return cb
}

// addIf adds arguments to the command only if the condition is true
func (cb *commandBuilder) addIf(condition bool, args ...string) *commandBuilder {
	if condition {
		cb.args = append(cb.args, args...)
	}
	return cb
}

// addIfNotEmpty adds arguments to the command only if the value is not empty
func (cb *commandBuilder) addIfNotEmpty(value string, args ...string) *commandBuilder {
	if value != "" {
		cb.args = append(cb.args, args...)
	}
	return cb
}

// build returns the final command slice
func (cb *commandBuilder) build() []string {
	return cb.args
}

// validateIntWithDefault applies a default value if input is 0, then validates the range
func validateIntWithDefault(value, defaultValue, min, max int, fieldName string) (int, error) {
	if value == 0 {
		value = defaultValue
	}
	if err := client.ValidateIntRange(value, min, max, fieldName); err != nil {
		return 0, err
	}
	return value, nil
}

// validateIntMax checks if a value exceeds the maximum and returns an error if it does
func validateIntMax(value, max int, fieldName, unit string) error {
	if value > max {
		if unit != "" {
			return fmt.Errorf("%s cannot exceed %d %s", fieldName, max, unit)
		}
		return fmt.Errorf("%s cannot exceed %d", fieldName, max)
	}
	return nil
}

// stringWithDefault returns the value if non-empty, otherwise returns the default
func stringWithDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// requireAtLeastNParams validates that at least N of the provided parameters are set
func requireAtLeastNParams(required int, params map[string]bool) error {
	count := 0
	var paramNames []string
	for name, isSet := range params {
		paramNames = append(paramNames, name)
		if isSet {
			count++
		}
	}
	if count < required {
		return fmt.Errorf("requires at least %d of: %v for safety", required, paramNames)
	}
	return nil
}
