package backend

import (
	"fmt"
	"strings"

	"github.com/Ooooze/batctl/internal/battery"
)

type SurfaceBackend struct{}

func (b *SurfaceBackend) Name() string {
	return "Surface"
}

func (b *SurfaceBackend) Detect() bool {
	vendor := DetectVendor()
	if !strings.Contains(vendor, "Microsoft") {
		return false
	}
	for _, bat := range battery.ListBatteries() {
		if battery.SysfsExists(battery.BatPath(bat, "charge_control_end_threshold")) {
			return true
		}
	}
	return false
}

func (b *SurfaceBackend) Capabilities() Capabilities {
	caps := Capabilities{
		StopThreshold:   true,
		ChargeBehaviour: false,
		StopRange:       [2]int{1, 100},
	}
	for _, bat := range battery.ListBatteries() {
		if battery.SysfsExists(battery.BatPath(bat, "charge_control_start_threshold")) {
			caps.StartThreshold = true
			caps.StartRange = [2]int{0, 99}
			break
		}
	}
	return caps
}

func (b *SurfaceBackend) GetThresholds(bat string) (start, stop int, err error) {
	stop, err = battery.SysfsReadInt(battery.BatPath(bat, "charge_control_end_threshold"))
	if err != nil {
		return 0, 0, err
	}
	startPath := battery.BatPath(bat, "charge_control_start_threshold")
	if battery.SysfsExists(startPath) {
		start, err = battery.SysfsReadInt(startPath)
		if err != nil {
			return 0, 0, err
		}
	}
	return start, stop, nil
}

func (b *SurfaceBackend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	caps := b.Capabilities()
	if caps.StartThreshold {
		if err := battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_start_threshold"), start); err != nil {
			return err
		}
	}
	return battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_end_threshold"), stop)
}

func (b *SurfaceBackend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	return "", nil, fmt.Errorf("not supported")
}

func (b *SurfaceBackend) SetChargeBehaviour(bat string, mode string) error {
	return fmt.Errorf("not supported")
}

func (b *SurfaceBackend) ValidateThresholds(start, stop int) error {
	caps := b.Capabilities()
	if caps.StartThreshold {
		if start < caps.StartRange[0] || start > caps.StartRange[1] {
			return fmt.Errorf("start threshold must be %d-%d, got %d", caps.StartRange[0], caps.StartRange[1], start)
		}
	}
	if stop < caps.StopRange[0] || stop > caps.StopRange[1] {
		return fmt.Errorf("stop threshold must be %d-%d, got %d", caps.StopRange[0], caps.StopRange[1], stop)
	}
	if caps.StartThreshold && start >= stop {
		return fmt.Errorf("start (%d) must be less than stop (%d)", start, stop)
	}
	return nil
}
