package backend

import (
	"fmt"
	"strings"

	"github.com/Ooooze/batctl/internal/battery"
)

const macsmcBattery = "macsmc-battery"

type AppleBackend struct{}

func (b *AppleBackend) Name() string {
	return "Apple"
}

func (b *AppleBackend) Detect() bool {
	vendor := DetectVendor()
	if !strings.Contains(vendor, "Apple") {
		return false
	}
	return battery.SysfsExists(battery.BatPath(macsmcBattery, "charge_control_end_threshold"))
}

func (b *AppleBackend) Capabilities() Capabilities {
	return Capabilities{
		StartThreshold:    false,
		StopThreshold:     true,
		ChargeBehaviour:   false,
		DiscreteStopVals:  []int{80, 100},
		StartAutoComputed: true,
	}
}

func (b *AppleBackend) GetThresholds(bat string) (start, stop int, err error) {
	start, err = battery.SysfsReadInt(battery.BatPath(macsmcBattery, "charge_control_start_threshold"))
	if err != nil {
		return 0, 0, err
	}
	stop, err = battery.SysfsReadInt(battery.BatPath(macsmcBattery, "charge_control_end_threshold"))
	if err != nil {
		return 0, 0, err
	}
	return start, stop, nil
}

func (b *AppleBackend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	return battery.SysfsWriteInt(battery.BatPath(macsmcBattery, "charge_control_end_threshold"), stop)
}

func (b *AppleBackend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	return "", nil, fmt.Errorf("not supported")
}

func (b *AppleBackend) SetChargeBehaviour(bat string, mode string) error {
	return fmt.Errorf("not supported")
}

func (b *AppleBackend) ValidateThresholds(start, stop int) error {
	if stop != 80 && stop != 100 {
		return fmt.Errorf("stop threshold must be 80 or 100, got %d", stop)
	}
	return nil
}
