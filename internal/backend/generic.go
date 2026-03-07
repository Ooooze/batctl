package backend

import (
	"fmt"
	"strings"

	"github.com/Ooooze/batctl/internal/battery"
)

type GenericBackend struct {
	caps *Capabilities
}

func (b *GenericBackend) Name() string {
	return "Generic"
}

func (b *GenericBackend) Detect() bool {
	for _, bat := range battery.ListBatteries() {
		if battery.SysfsExists(battery.BatPath(bat, "charge_control_end_threshold")) {
			return true
		}
	}
	return false
}

func (b *GenericBackend) Capabilities() Capabilities {
	if b.caps != nil {
		return *b.caps
	}
	caps := Capabilities{
		StartRange: [2]int{0, 99},
		StopRange:  [2]int{1, 100},
	}
	for _, bat := range battery.ListBatteries() {
		if battery.SysfsExists(battery.BatPath(bat, "charge_control_start_threshold")) {
			caps.StartThreshold = true
		}
		if battery.SysfsExists(battery.BatPath(bat, "charge_control_end_threshold")) {
			caps.StopThreshold = true
		}
		if battery.SysfsExists(battery.BatPath(bat, "charge_behaviour")) {
			caps.ChargeBehaviour = true
		}
	}
	b.caps = &caps
	return caps
}

func (b *GenericBackend) GetThresholds(bat string) (start, stop int, err error) {
	startPath := battery.BatPath(bat, "charge_control_start_threshold")
	stopPath := battery.BatPath(bat, "charge_control_end_threshold")
	if battery.SysfsExists(stopPath) {
		stop, err = battery.SysfsReadInt(stopPath)
		if err != nil {
			return 0, 0, err
		}
	}
	if battery.SysfsExists(startPath) {
		start, err = battery.SysfsReadInt(startPath)
		if err != nil {
			return 0, 0, err
		}
	}
	return start, stop, nil
}

func (b *GenericBackend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	caps := b.Capabilities()
	if caps.StartThreshold {
		if err := battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_start_threshold"), start); err != nil {
			return err
		}
	}
	if caps.StopThreshold {
		if err := battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_end_threshold"), stop); err != nil {
			return err
		}
	}
	return nil
}

func (b *GenericBackend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	path := battery.BatPath(bat, "charge_behaviour")
	if !battery.SysfsExists(path) {
		return "", nil, fmt.Errorf("not supported")
	}
	s, err := battery.SysfsReadString(path)
	if err != nil {
		return "", nil, err
	}
	available = []string{"auto", "inhibit-charge", "force-discharge"}
	parts := strings.Fields(s)
	for _, p := range parts {
		if strings.HasPrefix(p, "[") && strings.HasSuffix(p, "]") {
			current = strings.Trim(p, "[]")
			break
		}
	}
	return current, available, nil
}

func (b *GenericBackend) SetChargeBehaviour(bat string, mode string) error {
	path := battery.BatPath(bat, "charge_behaviour")
	if !battery.SysfsExists(path) {
		return fmt.Errorf("not supported")
	}
	return battery.SysfsWriteString(path, mode)
}

func (b *GenericBackend) ValidateThresholds(start, stop int) error {
	caps := b.Capabilities()
	if caps.StartThreshold {
		if start < caps.StartRange[0] || start > caps.StartRange[1] {
			return fmt.Errorf("start threshold must be %d-%d, got %d", caps.StartRange[0], caps.StartRange[1], start)
		}
	}
	if caps.StopThreshold {
		if stop < caps.StopRange[0] || stop > caps.StopRange[1] {
			return fmt.Errorf("stop threshold must be %d-%d, got %d", caps.StopRange[0], caps.StopRange[1], stop)
		}
	}
	if caps.StartThreshold && caps.StopThreshold && start >= stop {
		return fmt.Errorf("start (%d) must be less than stop (%d)", start, stop)
	}
	return nil
}
