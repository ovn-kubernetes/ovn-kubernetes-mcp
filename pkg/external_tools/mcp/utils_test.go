package external_tools

import (
	"testing"
)

func TestValidateIntMax(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		max       int
		fieldName string
		unit      string
		wantErr   bool
	}{
		{
			name:      "zero value",
			value:     0,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   false,
		},
		{
			name:      "minimum valid value",
			value:     1,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   false,
		},
		{
			name:      "value below max",
			value:     10,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   false,
		},
		{
			name:      "value equals max",
			value:     30,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   false,
		},
		{
			name:      "value just over max",
			value:     31,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   true,
		},
		{
			name:      "value exceeds max with unit",
			value:     100,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   true,
		},
		{
			name:      "value exceeds max without unit",
			value:     1001,
			max:       1000,
			fieldName: "packet_count",
			unit:      "",
			wantErr:   true,
		},
		{
			name:      "negative value -1",
			value:     -1,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   true,
		},
		{
			name:      "large negative value",
			value:     -100,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIntMax(tt.value, tt.max, tt.fieldName, tt.unit)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateIntMax() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		defaultValue string
		want         string
	}{
		{
			name:         "empty value returns default",
			value:        "",
			defaultValue: "text",
			want:         "text",
		},
		{
			name:         "non-empty value returns value",
			value:        "pcap",
			defaultValue: "text",
			want:         "pcap",
		},
		{
			name:         "both empty returns empty",
			value:        "",
			defaultValue: "",
			want:         "",
		},
		{
			name:         "whitespace value returns value",
			value:        " ",
			defaultValue: "text",
			want:         " ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringWithDefault(tt.value, tt.defaultValue)
			if got != tt.want {
				t.Errorf("stringWithDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommandBuilder(t *testing.T) {
	tests := []struct {
		name    string
		builder func() *commandBuilder
		want    []string
	}{
		{
			name: "simple command",
			builder: func() *commandBuilder {
				return newCommand("tcpdump")
			},
			want: []string{"tcpdump"},
		},
		{
			name: "command with args",
			builder: func() *commandBuilder {
				return newCommand("tcpdump").add("-i", "eth0")
			},
			want: []string{"tcpdump", "-i", "eth0"},
		},
		{
			name: "command with conditional args true",
			builder: func() *commandBuilder {
				return newCommand("tcpdump").addIf(true, "-v")
			},
			want: []string{"tcpdump", "-v"},
		},
		{
			name: "command with conditional args false",
			builder: func() *commandBuilder {
				return newCommand("tcpdump").addIf(false, "-v")
			},
			want: []string{"tcpdump"},
		},
		{
			name: "command with addIfNotEmpty with value",
			builder: func() *commandBuilder {
				return newCommand("tcpdump").addIfNotEmpty("tcp port 80", "tcp port 80")
			},
			want: []string{"tcpdump", "tcp port 80"},
		},
		{
			name: "command with addIfNotEmpty without value",
			builder: func() *commandBuilder {
				return newCommand("tcpdump").addIfNotEmpty("", "-f")
			},
			want: []string{"tcpdump"},
		},
		{
			name: "command with chained operations",
			builder: func() *commandBuilder {
				return newCommand("tcpdump").
					add("-i", "eth0").
					addIf(true, "-n").
					addIf(false, "-v").
					addIfNotEmpty("tcp", "tcp").
					addIfNotEmpty("", "-x")
			},
			want: []string{"tcpdump", "-i", "eth0", "-n", "tcp"},
		},
		{
			name: "command with multiple base args",
			builder: func() *commandBuilder {
				return newCommand("timeout", "30s", "retis")
			},
			want: []string{"timeout", "30s", "retis"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.builder().build()
			if len(got) != len(tt.want) {
				t.Errorf("commandBuilder.build() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("commandBuilder.build()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
