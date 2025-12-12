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
	if _, err := os.Stat(mustGatherPath); os.IsNotExist(err) {
		return fmt.Errorf("must gather path %s does not exist: %w", mustGatherPath, err)
	}
	if _, err := os.Stat(filepath.Join(mustGatherPath, "must-gather.log")); os.IsNotExist(err) {
		return fmt.Errorf("must gather path %s is not a valid must gather path: %w", mustGatherPath, err)
	}
	return nil
}
