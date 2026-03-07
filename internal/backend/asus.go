package backend

import (
	"fmt"
	"strings"

	"github.com/spaceclam/batctl/internal/battery"
)

type ASUSBackend struct{}

func (b *ASUSBackend) Name() string {
	return "ASUS"
}

func (b *ASUSBackend) Detect() bool {
	vendor := DetectVendor()
	if !strings.Contains(vendor, "ASUSTeK") {
		return false
	}
	for _, bat := range battery.ListBatteries() {
		if battery.SysfsExists(battery.BatPath(bat, "charge_control_end_threshold")) {
			return true
		}
	}
	return false
}

func (b *ASUSBackend) Capabilities() Capabilities {
	return Capabilities{
		StartThreshold:  false,
		StopThreshold:   true,
		ChargeBehaviour: false,
		StartRange:      [2]int{0, 0},
		StopRange:       [2]int{0, 100},
		DiscreteStopVals: nil,
	}
}

func (b *ASUSBackend) GetThresholds(bat string) (start, stop int, err error) {
	stop, err = battery.SysfsReadInt(battery.BatPath(bat, "charge_control_end_threshold"))
	if err != nil {
		return 0, 0, err
	}
	return 0, stop, nil
}

func (b *ASUSBackend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	return battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_end_threshold"), stop)
}

func (b *ASUSBackend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	return "", nil, fmt.Errorf("not supported")
}

func (b *ASUSBackend) SetChargeBehaviour(bat string, mode string) error {
	return fmt.Errorf("not supported")
}

func (b *ASUSBackend) ValidateThresholds(start, stop int) error {
	if stop < 0 || stop > 100 {
		return fmt.Errorf("stop threshold must be 0-100, got %d", stop)
	}
	return nil
}
