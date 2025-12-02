package kernel

import (
	"testing"
)

func TestValidateConntrackCommand(t *testing.T) {
	tests := []struct {
		name         string
		command      string
		cliAvailable bool
		wantError    bool
	}{
		{
			name:         "empty string with CLI available",
			command:      "",
			cliAvailable: true,
			wantError:    false,
		},
		{
			name:         "empty string without CLI",
			command:      "",
			cliAvailable: false,
			wantError:    false,
		},
		{
			name:         "numeric string with CLI available",
			command:      "321",
			cliAvailable: true,
			wantError:    true,
		},
		{
			name:         "numeric string without CLI",
			command:      "321",
			cliAvailable: false,
			wantError:    true,
		},
		{
			name:         "valid command --dump with CLI available",
			command:      "--dump",
			cliAvailable: true,
			wantError:    false,
		},
		{
			name:         "valid command --dump without CLI",
			command:      "--dump",
			cliAvailable: false,
			wantError:    false,
		},
		{
			name:         "valid command --stats with CLI available",
			command:      "--stats",
			cliAvailable: true,
			wantError:    false,
		},
		{
			name:         "valid command --stats without CLI",
			command:      "--stats",
			cliAvailable: false,
			wantError:    true,
		},
		{
			name:         "valid command --count with CLI available",
			command:      "--count",
			cliAvailable: true,
			wantError:    false,
		},
		{
			name:         "valid command --count without CLI",
			command:      "--count",
			cliAvailable: false,
			wantError:    true,
		},
		{
			name:         "valid command -L with CLI available",
			command:      "-L",
			cliAvailable: true,
			wantError:    false,
		},
		{
			name:         "valid command -L without CLI",
			command:      "-L",
			cliAvailable: false,
			wantError:    false,
		},
		{
			name:         "valid command -S with CLI available",
			command:      "-S",
			cliAvailable: true,
			wantError:    false,
		},
		{
			name:         "valid command -S without CLI",
			command:      "-S",
			cliAvailable: false,
			wantError:    true,
		},
		{
			name:         "valid command -C with CLI available",
			command:      "-C",
			cliAvailable: true,
			wantError:    false,
		},
		{
			name:         "valid command -C without CLI",
			command:      "-C",
			cliAvailable: false,
			wantError:    true,
		},
		{
			name:         "invalid command with CLI available",
			command:      "-E",
			cliAvailable: true,
			wantError:    true,
		},
		{
			name:         "invalid command without CLI",
			command:      "-E",
			cliAvailable: false,
			wantError:    true,
		},
		{
			name:         "invalid command -D with CLI available",
			command:      "-D",
			cliAvailable: true,
			wantError:    true,
		},
		{
			name:         "lowercase -l with CLI available",
			command:      "-l",
			cliAvailable: true,
			wantError:    true,
		},
		{
			name:         "command with spaces",
			command:      "-L ",
			cliAvailable: true,
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConntrackCommand(tt.command, tt.cliAvailable)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateConntrackCommand(%q, %v) error = %v, wantError %v", tt.command, tt.cliAvailable, err, tt.wantError)
			}
		})
	}
}
