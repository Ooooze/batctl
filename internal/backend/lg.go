package backend

import (
	"fmt"
	"strings"

	"github.com/Ooooze/batctl/internal/battery"
)

const lgBatteryCareLimitPath = "/sys/devices/platform/lg-laptop/battery_care_limit"

type LGBackend struct {
	legacy bool
}

func (b *LGBackend) Name() string {
	return "LG"
}

func (b *LGBackend) Detect() bool {
	vendor := DetectVendor()
	if !strings.HasPrefix(vendor, "LG") {
		return false
	}
	if battery.SysfsExists(lgBatteryCareLimitPath) {
		b.legacy = true
		return true
	}
	for _, bat := range battery.ListBatteries() {
		if battery.SysfsExists(battery.BatPath(bat, "charge_control_end_threshold")) {
			b.legacy = false
			return true
		}
	}
	return false
}

func (b *LGBackend) Capabilities() Capabilities {
	return Capabilities{
		StartThreshold:    false,
		StopThreshold:     true,
		ChargeBehaviour:   false,
		StartRange:        [2]int{0, 0},
		StopRange:         [2]int{80, 100},
		DiscreteStopVals:  []int{80, 100},
		StartAutoComputed: false,
	}
}

func (b *LGBackend) GetThresholds(bat string) (start, stop int, err error) {
	if b.legacy {
		stop, err = battery.SysfsReadInt(lgBatteryCareLimitPath)
		if err != nil {
			return 0, 0, err
		}
		return 0, stop, nil
	}
	stop, err = battery.SysfsReadInt(battery.BatPath(bat, "charge_control_end_threshold"))
	if err != nil {
		return 0, 0, err
	}
	return 0, stop, nil
}

func (b *LGBackend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	if b.legacy {
		return battery.SysfsWriteInt(lgBatteryCareLimitPath, stop)
	}
	return battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_end_threshold"), stop)
}

func (b *LGBackend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	return "", nil, fmt.Errorf("not supported")
}

func (b *LGBackend) SetChargeBehaviour(bat string, mode string) error {
	return fmt.Errorf("not supported")
}

func (b *LGBackend) ValidateThresholds(start, stop int) error {
	if stop != 80 && stop != 100 {
		return fmt.Errorf("stop threshold must be 80 or 100, got %d", stop)
	}
	return nil
}
