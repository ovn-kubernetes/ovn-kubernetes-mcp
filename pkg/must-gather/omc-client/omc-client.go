package omcclient

import (
	"fmt"
	"os/exec"

	mustgatherUtils "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/must-gather/utils"
)

// OmcClient is a client for the omc command.
type OmcClient struct {
	CommandPath string
}

// NewOmcClient creates a new OmcClient. It will return an error if the omc command path is not found.
func NewOmcClient() (*OmcClient, error) {
	commandPath, err := getOmcCommandPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get omc command path: %w", err)
	}
	return &OmcClient{
		CommandPath: commandPath,
	}, nil
}

// useMustGather uses the omc command to use the must gather path. It will return an error if the
// must gather path is not valid.
func (c *OmcClient) useMustGather(mustGatherPath string) error {
	err := mustgatherUtils.ValidateMustGatherPath(mustGatherPath)
	if err != nil {
		return fmt.Errorf("failed to validate must gather path: %w", err)
	}
	args := []string{"use", mustGatherPath}
	output, err := exec.Command(c.CommandPath, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute omc command. output: %s, error: %w", string(output), err)
	}
	return nil
}
