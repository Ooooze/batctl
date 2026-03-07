package backend

import (
	"fmt"
	"strings"

	"github.com/spaceclam/batctl/internal/battery"
)

type ThinkPadBackend struct{}

func (b *ThinkPadBackend) Name() string {
	return "ThinkPad"
}

func (b *ThinkPadBackend) Detect() bool {
	vendor := DetectVendor()
	if !strings.Contains(vendor, "LENOVO") {
		return false
	}
	if !battery.SysfsExists(battery.BatPath("BAT0", "charge_control_start_threshold")) {
		return false
	}
	for _, attr := range []string{"product_name", "product_family", "product_version"} {
		if strings.Contains(battery.DMIRead(attr), "ThinkPad") {
			return true
		}
	}
	return false
}

func (b *ThinkPadBackend) Capabilities() Capabilities {
	return Capabilities{
		StartThreshold:  true,
		StopThreshold:    true,
		ChargeBehaviour:  true,
		StartRange:       [2]int{0, 99},
		StopRange:        [2]int{1, 100},
		DiscreteStopVals: nil,
	}
}

func (b *ThinkPadBackend) GetThresholds(bat string) (start, stop int, err error) {
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

func (b *ThinkPadBackend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	if err := battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_start_threshold"), start); err != nil {
		return err
	}
	return battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_end_threshold"), stop)
}

func (b *ThinkPadBackend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	s, err := battery.SysfsReadString(battery.BatPath(bat, "charge_behaviour"))
	if err != nil {
		return "", nil, err
	}
	available = []string{"auto", "inhibit-charge", "force-discharge"}
	idx := strings.Index(s, "[")
	if idx >= 0 {
		end := strings.Index(s[idx:], "]")
		if end >= 0 {
			current = strings.TrimSpace(s[idx+1 : idx+end])
		}
	}
	return current, available, nil
}

func (b *ThinkPadBackend) SetChargeBehaviour(bat string, mode string) error {
	return battery.SysfsWriteString(battery.BatPath(bat, "charge_behaviour"), mode)
}

func (b *ThinkPadBackend) ValidateThresholds(start, stop int) error {
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
