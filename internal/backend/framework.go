package backend

import (
	"fmt"
	"strings"

	"github.com/Ooooze/batctl/internal/battery"
)

type FrameworkBackend struct{}

func (b *FrameworkBackend) Name() string {
	return "Framework"
}

func (b *FrameworkBackend) Detect() bool {
	vendor := DetectVendor()
	if !strings.Contains(vendor, "Framework") {
		return false
	}
	for _, bat := range battery.ListBatteries() {
		if battery.SysfsExists(battery.BatPath(bat, "charge_control_end_threshold")) {
			return true
		}
	}
	return false
}

func (b *FrameworkBackend) Capabilities() Capabilities {
	startSupported := false
	chargeBehaviourSupported := false
	for _, bat := range battery.ListBatteries() {
		if battery.SysfsExists(battery.BatPath(bat, "charge_control_start_threshold")) {
			startSupported = true
		}
		if battery.SysfsExists(battery.BatPath(bat, "charge_behaviour")) {
			chargeBehaviourSupported = true
		}
		if startSupported && chargeBehaviourSupported {
			break
		}
	}
	caps := Capabilities{
		StartThreshold:   startSupported,
		StopThreshold:    true,
		ChargeBehaviour:  chargeBehaviourSupported,
		StopRange:        [2]int{1, 100},
		DiscreteStopVals: nil,
	}
	if startSupported {
		caps.StartRange = [2]int{0, 99}
	}
	return caps
}

func (b *FrameworkBackend) GetThresholds(bat string) (start, stop int, err error) {
	stop, err = battery.SysfsReadInt(battery.BatPath(bat, "charge_control_end_threshold"))
	if err != nil {
		return 0, 0, err
	}
	if battery.SysfsExists(battery.BatPath(bat, "charge_control_start_threshold")) {
		start, err = battery.SysfsReadInt(battery.BatPath(bat, "charge_control_start_threshold"))
		if err != nil {
			return 0, 0, err
		}
	}
	return start, stop, nil
}

func (b *FrameworkBackend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	if err := battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_end_threshold"), stop); err != nil {
		return err
	}
	if battery.SysfsExists(battery.BatPath(bat, "charge_control_start_threshold")) {
		if err := battery.SysfsWriteInt(battery.BatPath(bat, "charge_control_start_threshold"), start); err != nil {
			return err
		}
	}
	return nil
}

func (b *FrameworkBackend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	if !battery.SysfsExists(battery.BatPath(bat, "charge_behaviour")) {
		return "", nil, fmt.Errorf("not supported")
	}
	s, err := battery.SysfsReadString(battery.BatPath(bat, "charge_behaviour"))
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

func (b *FrameworkBackend) SetChargeBehaviour(bat string, mode string) error {
	if !battery.SysfsExists(battery.BatPath(bat, "charge_behaviour")) {
		return fmt.Errorf("not supported")
	}
	return battery.SysfsWriteString(battery.BatPath(bat, "charge_behaviour"), mode)
}

func (b *FrameworkBackend) ValidateThresholds(start, stop int) error {
	caps := b.Capabilities()
	if caps.StartThreshold && (start < 0 || start > 99) {
		return fmt.Errorf("start threshold must be 0-99, got %d", start)
	}
	if stop < 1 || stop > 100 {
		return fmt.Errorf("stop threshold must be 1-100, got %d", stop)
	}
	return nil
}
