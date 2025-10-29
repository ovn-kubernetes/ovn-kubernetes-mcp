package sosreport

import (
	"testing"
)

func TestLoadManifest(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid manifest",
			path:      sosreportTestData,
			wantError: false,
		},
		{
			name:      "non-existent path",
			path:      "testdata/non-existent",
			wantError: true,
			errorMsg:  "failed to read manifest.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := loadManifest(tt.path)
			if tt.wantError {
				if err == nil {
					t.Errorf("loadManifest() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("loadManifest() unexpected error = %v", err)
					return
				}
				if manifest == nil {
					t.Errorf("loadManifest() returned nil manifest")
					return
				}
				// Verify the manifest has expected structure
				if manifest.Components.Report.Plugins == nil {
					t.Errorf("loadManifest() manifest.Components.Report.Plugins is nil")
				}
			}
		})
	}
}
