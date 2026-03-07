package backend

import (
	"fmt"
	"strings"

	"github.com/spaceclam/batctl/internal/battery"
)

type MSIBackend struct{}

func (b *MSIBackend) Name() string {
	return "MSI"
}

func (b *MSIBackend) Detect() bool {
	vendor := DetectVendor()
	if !strings.Contains(vendor, "Micro-Star") {
		return false
	}
	for _, bat := range []string{"BAT0", "BAT1"} {
		if battery.SysfsExists(battery.BatPath(bat, "charge_control_end_threshold")) {
			return true
		}
	}
	return false
}

func (b *MSIBackend) Capabilities() Capabilities {
	return Capabilities{
		StartThreshold:    false,
		StopThreshold:     true,
		ChargeBehaviour:   false,
		StartRange:        [2]int{0, 0},
		StopRange:         [2]int{10, 100},
		DiscreteStopVals:  nil,
		StartAutoComputed: true,
	}
}

func (b *MSIBackend) GetThresholds(bat string) (start, stop int, err error) {
	start, err = battery.SysfsReadInt(battery.BatPath(bat, "charge_control_start_threshold"))
	if err != nil {
		return 0, 0, err
	}
	stop, err = battery.SysfsReadInt(battery.BatPath(bat, "charge_control_end_threshold"))
	if err != nil {
		return 0, 0, err
	}
	return start, stop, nil
}

func (b *MSIBackend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	return battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_end_threshold"), stop)
}

func (b *MSIBackend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	return "", nil, fmt.Errorf("not supported")
}

func (b *MSIBackend) SetChargeBehaviour(bat string, mode string) error {
	return fmt.Errorf("not supported")
}

func (b *MSIBackend) ValidateThresholds(start, stop int) error {
	if stop < 10 || stop > 100 {
		return fmt.Errorf("stop threshold must be 10-100, got %d", stop)
	}
	return nil
}
