package backend

import (
	"fmt"
	"strings"

	"github.com/Ooooze/batctl/internal/battery"
)

const sonyBatteryCareLimiterPath = "/sys/devices/platform/sony-laptop/battery_care_limiter"

type SonyBackend struct{}

func (b *SonyBackend) Name() string {
	return "Sony"
}

func (b *SonyBackend) Detect() bool {
	vendor := DetectVendor()
	if !strings.Contains(vendor, "Sony") {
		return false
	}
	return battery.SysfsExists(sonyBatteryCareLimiterPath)
}

func (b *SonyBackend) Capabilities() Capabilities {
	return Capabilities{
		StartThreshold:   false,
		StopThreshold:    true,
		ChargeBehaviour:  false,
		StartRange:       [2]int{0, 0},
		StopRange:        [2]int{0, 0},
		DiscreteStopVals: []int{50, 80, 100},
	}
}

func (b *SonyBackend) GetThresholds(bat string) (start, stop int, err error) {
	stop, err = battery.SysfsReadInt(sonyBatteryCareLimiterPath)
	if err != nil {
		return 0, 0, err
	}
	return 0, stop, nil
}

func (b *SonyBackend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	return battery.SysfsWriteInt(sonyBatteryCareLimiterPath, stop)
}

func (b *SonyBackend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	return "", nil, fmt.Errorf("not supported")
}

func (b *SonyBackend) SetChargeBehaviour(bat string, mode string) error {
	return fmt.Errorf("not supported")
}

func (b *SonyBackend) ValidateThresholds(start, stop int) error {
	valid := false
	for _, v := range []int{50, 80, 100} {
		if stop == v {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("stop threshold must be 50, 80, or 100, got %d", stop)
	}
	return nil
}
