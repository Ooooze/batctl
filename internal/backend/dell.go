package backend

import (
	"fmt"
	"strings"

	"github.com/Ooooze/batctl/internal/battery"
)

type DellBackend struct{}

func (b *DellBackend) Name() string {
	return "Dell"
}

func (b *DellBackend) Detect() bool {
	vendor := DetectVendor()
	if !strings.Contains(vendor, "Dell") {
		return false
	}
	for _, bat := range battery.ListBatteries() {
		if battery.SysfsExists(battery.BatPath(bat, "charge_control_end_threshold")) {
			return true
		}
	}
	return false
}

func (b *DellBackend) Capabilities() Capabilities {
	return Capabilities{
		StartThreshold:  true,
		StopThreshold:   true,
		ChargeBehaviour: false,
		StartRange:      [2]int{50, 95},
		StopRange:       [2]int{55, 100},
		DiscreteStopVals: nil,
	}
}

func (b *DellBackend) GetThresholds(bat string) (start, stop int, err error) {
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

func (b *DellBackend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	chargeTypesPath := battery.BatPath(bat, "charge_types")
	if battery.SysfsExists(chargeTypesPath) {
		if err := battery.SysfsWriteString(chargeTypesPath, "Custom"); err != nil {
			return err
		}
	}
	if err := battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_start_threshold"), start); err != nil {
		return err
	}
	return battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_end_threshold"), stop)
}

func (b *DellBackend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	return "", nil, fmt.Errorf("not supported")
}

func (b *DellBackend) SetChargeBehaviour(bat string, mode string) error {
	return fmt.Errorf("not supported")
}

func (b *DellBackend) ValidateThresholds(start, stop int) error {
	if start < 50 || start > 95 {
		return fmt.Errorf("start threshold must be 50-95, got %d", start)
	}
	if stop < 55 || stop > 100 {
		return fmt.Errorf("stop threshold must be 55-100, got %d", stop)
	}
	if start >= stop {
		return fmt.Errorf("start (%d) must be less than stop (%d)", start, stop)
	}
	if stop-start != 5 {
		return fmt.Errorf("hardware enforces start = stop - 5, got start=%d stop=%d", start, stop)
	}
	return nil
}
