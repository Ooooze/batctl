package backend

import (
	"fmt"
	"strings"

	"github.com/Ooooze/batctl/internal/battery"
)

const samsungBatteryLifeExtenderPath = "/sys/devices/platform/samsung/battery_life_extender"

type SamsungBackend struct{}

func (b *SamsungBackend) Name() string {
	return "Samsung"
}

func (b *SamsungBackend) Detect() bool {
	vendor := DetectVendor()
	if !strings.Contains(vendor, "SAMSUNG") {
		return false
	}
	return battery.SysfsExists(samsungBatteryLifeExtenderPath)
}

func (b *SamsungBackend) Capabilities() Capabilities {
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

func (b *SamsungBackend) GetThresholds(bat string) (start, stop int, err error) {
	val, err := battery.SysfsReadInt(samsungBatteryLifeExtenderPath)
	if err != nil {
		return 0, 0, err
	}
	if val == 1 {
		return 0, 80, nil
	}
	return 0, 100, nil
}

func (b *SamsungBackend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	if stop <= 80 {
		return battery.SysfsWriteInt(samsungBatteryLifeExtenderPath, 1)
	}
	return battery.SysfsWriteInt(samsungBatteryLifeExtenderPath, 0)
}

func (b *SamsungBackend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	return "", nil, fmt.Errorf("not supported")
}

func (b *SamsungBackend) SetChargeBehaviour(bat string, mode string) error {
	return fmt.Errorf("not supported")
}

func (b *SamsungBackend) ValidateThresholds(start, stop int) error {
	if stop != 80 && stop != 100 {
		return fmt.Errorf("stop threshold must be 80 or 100, got %d", stop)
	}
	return nil
}
