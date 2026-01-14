package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateMustGatherPath validates the must gather path. It will check if the path is not empty,
// exists, is an absolute path and contains the must-gather.log file. If the path is not valid,
// it will return an error.
func ValidateMustGatherPath(mustGatherPath string) error {
	if mustGatherPath == "" {
		return fmt.Errorf("must gather path is required")
	}
	if !filepath.IsAbs(mustGatherPath) {
		return fmt.Errorf("must gather path %s is not an absolute path", mustGatherPath)
	}
	mustGatherPath = filepath.Clean(mustGatherPath)
	if _, err := os.Stat(mustGatherPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("must gather path %s does not exist: %w", mustGatherPath, err)
		}
		return fmt.Errorf("failed to stat must gather path %s: %w", mustGatherPath, err)
	}

	if _, err := os.Stat(filepath.Join(mustGatherPath, "must-gather.log")); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("must gather log file %s not found: %w", filepath.Join(mustGatherPath, "must-gather.log"), err)
		}
		return fmt.Errorf("failed to stat must gather log file %s: %w", filepath.Join(mustGatherPath, "must-gather.log"), err)
	}
	return nil
}
