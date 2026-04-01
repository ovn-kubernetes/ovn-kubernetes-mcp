package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
)

// ValidateMustGatherPath validates the must gather path. It will check if the path is not empty,
// exists, is an absolute path and contains the must-gather.logs file. If the path is not valid,
// it will return an error.
func ValidateMustGatherPath(mustGatherPath string) error {
	if err := utils.ValidatePath(mustGatherPath, "must gather path", false); err != nil {
		return err
	}
	mustGatherPath = filepath.Clean(mustGatherPath)
	if _, err := os.Stat(mustGatherPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("must gather path %s does not exist: %w", mustGatherPath, err)
		}
		return fmt.Errorf("failed to stat must gather path %s: %w", mustGatherPath, err)
	}

	if _, err := os.Stat(filepath.Join(mustGatherPath, "must-gather.logs")); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("must gather log file %s not found: %w", filepath.Join(mustGatherPath, "must-gather.logs"), err)
		}
		return fmt.Errorf("failed to stat must gather log file %s: %w", filepath.Join(mustGatherPath, "must-gather.logs"), err)
	}
	return nil
}
