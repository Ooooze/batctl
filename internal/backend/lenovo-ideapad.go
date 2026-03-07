package backend

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spaceclam/batctl/internal/battery"
)

type LenovoIdeaPadBackend struct {
	conservationPath string
}

func (b *LenovoIdeaPadBackend) Name() string {
	return "Lenovo IdeaPad"
}

func (b *LenovoIdeaPadBackend) Detect() bool {
	vendor := DetectVendor()
	if !strings.Contains(vendor, "LENOVO") {
		return false
	}
	product := DetectProductName()
	if strings.Contains(product, "ThinkPad") {
		return false
	}
	matches, err := filepath.Glob("/sys/bus/platform/drivers/ideapad_acpi/VPC*/conservation_mode")
	if err != nil || len(matches) == 0 {
		return false
	}
	b.conservationPath = matches[0]
	return true
}

func (b *LenovoIdeaPadBackend) Capabilities() Capabilities {
	return Capabilities{
		StartThreshold:    false,
		StopThreshold:     true,
		ChargeBehaviour:   false,
		StartRange:        [2]int{0, 0},
		StopRange:         [2]int{0, 1},
		DiscreteStopVals:  []int{0, 1},
		StartAutoComputed: false,
	}
}

func (b *LenovoIdeaPadBackend) GetThresholds(bat string) (start, stop int, err error) {
	val, err := battery.SysfsReadInt(b.conservationPath)
	if err != nil {
		return 0, 0, err
	}
	if val == 1 {
		return 0, 60, nil
	}
	return 0, 100, nil
}

func (b *LenovoIdeaPadBackend) SetThresholds(bat string, start, stop int) error {
	if err := b.ValidateThresholds(start, stop); err != nil {
		return err
	}
	if stop < 100 {
		return battery.SysfsWriteInt(b.conservationPath, 1)
	}
	return battery.SysfsWriteInt(b.conservationPath, 0)
}

func (b *LenovoIdeaPadBackend) GetChargeBehaviour(bat string) (current string, available []string, err error) {
	return "", nil, fmt.Errorf("not supported")
}

func (b *LenovoIdeaPadBackend) SetChargeBehaviour(bat string, mode string) error {
	return fmt.Errorf("not supported")
}

func (b *LenovoIdeaPadBackend) ValidateThresholds(start, stop int) error {
	if stop != 0 && stop != 1 && stop != 60 && stop != 100 {
		return fmt.Errorf("stop threshold must be 0, 1, 60, or 100, got %d", stop)
	}
	return nil
}
