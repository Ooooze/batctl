package backend

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Ooooze/batctl/internal/battery"
)

const huaweiThresholdsPath = "/sys/devices/platform/huawei-wmi/charge_control_thresholds"

type HuaweiBackend struct{}

func (b *HuaweiBackend) Name() string {
	return "Huawei"
}

func (b *HuaweiBackend) Detect() bool {
	vendor := DetectVendor()
	if !strings.Contains(vendor, "HUAWEI") {
		return false
	}
	return battery.SysfsExists(huaweiThresholdsPath)
}

func (b *HuaweiBackend) Capabilities() Capabilities {
	return Capabilities{
		StartThreshold:    true,
		StopThreshold:     true,
		ChargeBehaviour:   false,
		StartRange:        [2]int{0, 99},
		StopRange:         [2]int{1, 100},
		DiscreteStopVals:  nil,
		StartAutoComputed: false,
	}
}

func (b *HuaweiBackend) GetThresholds(bat string) (start, stop int, err error) {
	s, err := battery.SysfsReadString(huaweiThresholdsPath)
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Fields(s)
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid format: expected 'start stop'")
	}
	start, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	stop, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return start, stop, nil
}

func (b *HuaweiBackend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	return battery.SysfsWriteString(huaweiThresholdsPath, fmt.Sprintf("%d %d", start, stop))
}

func (b *HuaweiBackend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	return "", nil, fmt.Errorf("not supported")
}

func (b *HuaweiBackend) SetChargeBehaviour(bat string, mode string) error {
	return fmt.Errorf("not supported")
}

func (b *HuaweiBackend) ValidateThresholds(start, stop int) error {
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
