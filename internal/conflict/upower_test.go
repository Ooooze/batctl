package conflict

import (
	"fmt"
	"testing"
)

func TestParseBool(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{
			name: "should parse true",
			input: "method return time=1773048546.054101 sender=:1.25 -> destination=:1.1519 serial=9129 reply_serial=2\n" +
				"   variant       boolean true\n",
			want: true,
		},
		{
			name: "should parse false",
			input: "method return time=1773048546.051324 sender=:1.25 -> destination=:1.1518 serial=9128 reply_serial=2\n" +
				"   variant       boolean false\n",
			want: false,
		},
		{
			name:    "should error on missing variant line",
			input:   "some other output\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBool(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseUint(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name: "should parse 80",
			input: "method return time=1773048546.056970 sender=:1.25 -> destination=:1.1520 serial=9130 reply_serial=2\n" +
				"   variant       uint32 80\n",
			want: 80,
		},
		{
			name: "should parse zero",
			input: "method return time=1773048546.056970 sender=:1.25 -> destination=:1.1520 serial=9130 reply_serial=2\n" +
				"   variant       uint32 0\n",
			want: 0,
		},
		{
			name:    "should error on missing variant line",
			input:   "unexpected output\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUint(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseUint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseUint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckUPower(t *testing.T) {
	t.Run("should detect upower managing thresholds", func(t *testing.T) {
		original := runDBusGet
		defer func() { runDBusGet = original }()

		runDBusGet = func(devPath, prop string) (string, error) {
			responses := map[string]string{
				"ChargeThresholdSupported": "method return\n   variant       boolean true\n",
				"ChargeThresholdEnabled":   "method return\n   variant       boolean true\n",
				"ChargeEndThreshold":       "method return\n   variant       uint32 80\n",
				"ChargeStartThreshold":     "method return\n   variant       uint32 75\n",
			}
			if resp, ok := responses[prop]; ok {
				return resp, nil
			}
			return "", fmt.Errorf("unknown property")
		}

		info := CheckUPower("BAT0")
		if !info.Available {
			t.Error("expected Available=true")
		}
		if !info.Supported {
			t.Error("expected Supported=true")
		}
		if !info.Enabled {
			t.Error("expected Enabled=true")
		}
		if info.EndThreshold != 80 {
			t.Errorf("expected EndThreshold=80, got %d", info.EndThreshold)
		}
		if info.StartThreshold != 75 {
			t.Errorf("expected StartThreshold=75, got %d", info.StartThreshold)
		}
	})

	t.Run("should handle upower not available", func(t *testing.T) {
		original := runDBusGet
		defer func() { runDBusGet = original }()

		runDBusGet = func(devPath, prop string) (string, error) {
			return "", fmt.Errorf("connection refused")
		}

		info := CheckUPower("BAT0")
		if info.Available {
			t.Error("expected Available=false")
		}
		if info.Enabled {
			t.Error("expected Enabled=false")
		}
	})

	t.Run("should handle thresholds not supported", func(t *testing.T) {
		original := runDBusGet
		defer func() { runDBusGet = original }()

		runDBusGet = func(devPath, prop string) (string, error) {
			if prop == "ChargeThresholdSupported" {
				return "method return\n   variant       boolean false\n", nil
			}
			return "", fmt.Errorf("should not be called")
		}

		info := CheckUPower("BAT0")
		if !info.Available {
			t.Error("expected Available=true")
		}
		if info.Supported {
			t.Error("expected Supported=false")
		}
		if info.Enabled {
			t.Error("expected Enabled=false")
		}
	})

	t.Run("should handle supported but not enabled", func(t *testing.T) {
		original := runDBusGet
		defer func() { runDBusGet = original }()

		runDBusGet = func(devPath, prop string) (string, error) {
			responses := map[string]string{
				"ChargeThresholdSupported": "method return\n   variant       boolean true\n",
				"ChargeThresholdEnabled":   "method return\n   variant       boolean false\n",
				"ChargeEndThreshold":       "method return\n   variant       uint32 80\n",
				"ChargeStartThreshold":     "method return\n   variant       uint32 75\n",
			}
			if resp, ok := responses[prop]; ok {
				return resp, nil
			}
			return "", fmt.Errorf("unknown property")
		}

		info := CheckUPower("BAT0")
		if !info.Available {
			t.Error("expected Available=true")
		}
		if !info.Supported {
			t.Error("expected Supported=true")
		}
		if info.Enabled {
			t.Error("expected Enabled=false")
		}
	})

	t.Run("should use correct dbus path", func(t *testing.T) {
		original := runDBusGet
		defer func() { runDBusGet = original }()

		var capturedPath string
		runDBusGet = func(devPath, prop string) (string, error) {
			capturedPath = devPath
			return "method return\n   variant       boolean false\n", nil
		}

		CheckUPower("BAT1")
		expected := "/org/freedesktop/UPower/devices/battery_BAT1"
		if capturedPath != expected {
			t.Errorf("expected devPath=%q, got %q", expected, capturedPath)
		}
	})
}
