package backend

import (
	"fmt"
	"strings"

	"github.com/spaceclam/batctl/internal/battery"
)

const tuxedoKeyboardPath = "/sys/devices/platform/tuxedo_keyboard"

type TuxedoBackend struct {
	discreteStartVals []int
}

func (b *TuxedoBackend) Name() string {
	return "Tuxedo"
}

func (b *TuxedoBackend) Detect() bool {
	vendor := DetectVendor()
	hasVendor := strings.Contains(vendor, "TUXEDO")
	hasKeyboard := battery.SysfsExists(tuxedoKeyboardPath)
	if !hasVendor && !hasKeyboard {
		return false
	}
	for _, bat := range battery.ListBatteries() {
		if battery.SysfsExists(battery.BatPath(bat, "charge_control_end_threshold")) {
			return true
		}
	}
	return false
}

func (b *TuxedoBackend) Capabilities() Capabilities {
	return Capabilities{
		StartThreshold:   true,
		StopThreshold:    true,
		ChargeBehaviour:  false,
		StartRange:       [2]int{40, 95},
		StopRange:        [2]int{60, 100},
		DiscreteStopVals: []int{60, 70, 80, 90, 100},
	}
}

func (b *TuxedoBackend) discreteStartValues() []int {
	if b.discreteStartVals == nil {
		return []int{40, 50, 60, 70, 80, 95}
	}
	return b.discreteStartVals
}

func (b *TuxedoBackend) GetThresholds(bat string) (start, stop int, err error) {
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

func (b *TuxedoBackend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	chargeTypePath := battery.BatPath(bat, "charge_type")
	if battery.SysfsExists(chargeTypePath) {
		if err := battery.SysfsWriteString(chargeTypePath, "Custom"); err != nil {
			return err
		}
	}
	if err := battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_start_threshold"), start); err != nil {
		return err
	}
	return battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_end_threshold"), stop)
}

func (b *TuxedoBackend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	return "", nil, fmt.Errorf("not supported")
}

func (b *TuxedoBackend) SetChargeBehaviour(bat string, mode string) error {
	return fmt.Errorf("not supported")
}

func (b *TuxedoBackend) ValidateThresholds(start, stop int) error {
	startValid := false
	for _, v := range b.discreteStartValues() {
		if start == v {
			startValid = true
			break
		}
	}
	if !startValid {
		return fmt.Errorf("start threshold must be one of [40, 50, 60, 70, 80, 95], got %d", start)
	}
	stopValid := false
	for _, v := range []int{60, 70, 80, 90, 100} {
		if stop == v {
			stopValid = true
			break
		}
	}
	if !stopValid {
		return fmt.Errorf("stop threshold must be one of [60, 70, 80, 90, 100], got %d", stop)
	}
	if start >= stop {
		return fmt.Errorf("start (%d) must be less than stop (%d)", start, stop)
	}
	return nil
}
