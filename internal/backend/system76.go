package backend

import (
	"fmt"
	"strings"

	"github.com/Ooooze/batctl/internal/battery"
)

type System76Backend struct{}

func (b *System76Backend) Name() string {
	return "System76"
}

func (b *System76Backend) Detect() bool {
	vendor := DetectVendor()
	if !strings.Contains(vendor, "System76") {
		return false
	}
	for _, bat := range battery.ListBatteries() {
		if battery.SysfsExists(battery.BatPath(bat, "charge_control_end_threshold")) {
			return true
		}
	}
	return false
}

func (b *System76Backend) Capabilities() Capabilities {
	return Capabilities{
		StartThreshold:   true,
		StopThreshold:    true,
		ChargeBehaviour:  false,
		StartRange:       [2]int{0, 99},
		StopRange:        [2]int{1, 100},
		DiscreteStopVals: nil,
	}
}

func (b *System76Backend) GetThresholds(bat string) (start, stop int, err error) {
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

func (b *System76Backend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	if err := battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_start_threshold"), start); err != nil {
		return err
	}
	return battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_end_threshold"), stop)
}

func (b *System76Backend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	return "", nil, fmt.Errorf("not supported")
}

func (b *System76Backend) SetChargeBehaviour(bat string, mode string) error {
	return fmt.Errorf("not supported")
}

func (b *System76Backend) ValidateThresholds(start, stop int) error {
	if start < 0 || start > 99 {
		return fmt.Errorf("start threshold must be 0-99, got %d", start)
	}
	if stop < 1 || stop > 100 {
		return fmt.Errorf("stop threshold must be 1-100, got %d", stop)
	}
	if start >= stop {
		return fmt.Errorf("start (%d) must be less than stop (%d)", start, stop)
	}
	return nil
}
