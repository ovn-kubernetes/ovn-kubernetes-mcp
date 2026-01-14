package utils

import (
	"path/filepath"
	"testing"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
)

func TestValidateMustGatherPath(t *testing.T) {
	gitRoot, err := utils.GetGitRepositoryRoot()
	if err != nil {
		t.Fatalf("failed to get git repository root: %v", err)
	}
	tests := []struct {
		testName       string
		mustGatherPath string
		wantError      bool
	}{
		{
			testName:       "empty must gather path",
			mustGatherPath: "",
			wantError:      true,
		},
		{
			testName:       "valid must gather path",
			mustGatherPath: filepath.Join(gitRoot, "pkg", "must-gather", "testdata", "must-gather"),
			wantError:      false,
		},
		{
			testName:       "invalid must gather path",
			mustGatherPath: filepath.Join(gitRoot, "invalid", "path"),
			wantError:      true,
		},
		{
			testName:       "relative must gather path",
			mustGatherPath: "testdata/must-gather",
			wantError:      true,
		},
		{
			testName:       "absolute path without must-gather.log",
			mustGatherPath: filepath.Join(gitRoot, "pkg", "must-gather", "testdata", "invalid-must-gather"),
			wantError:      true,
		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			err := ValidateMustGatherPath(test.mustGatherPath)
			if (err != nil) != test.wantError {
				t.Fatalf("ValidateMustGatherPath() error = %v, wantErr %v", err, test.wantError)
			}
		})
	}
}
