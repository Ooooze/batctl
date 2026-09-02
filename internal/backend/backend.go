package backend

import (
	"errors"
	"fmt"
)

type Capabilities struct {
	StartThreshold    bool
	StopThreshold     bool
	ChargeBehaviour   bool
	StartRange        [2]int // min, max
	StopRange         [2]int // min, max
	DiscreteStopVals  []int  // for vendors with fixed stop values (Sony: 50/80/100)
	StartAutoComputed bool   // MSI: start = stop - 10 automatically
	StartStopGap      int    // if non-zero, hardware enforces start = stop - gap (Dell: 5)
}

// ErrNotChargeable is returned by SetThresholds when the target battery does
// not expose any charge control attributes (e.g. HID++ peripheral batteries,
// which appear under /sys/class/power_supply/ with type=Battery but have no
// charge_control_* files).
var ErrNotChargeable = errors.New("battery does not expose charge control attributes")

type Backend interface {
	Name() string
	Detect() bool
	Capabilities() Capabilities
	GetThresholds(bat string) (start, stop int, err error)
	SetThresholds(bat string, start, stop int) error
	GetChargeBehaviour(bat string) (current string, available []string, err error)
	SetChargeBehaviour(bat string, mode string) error
	ValidateThresholds(start, stop int) error
}

var registry []Backend

func Register(b Backend) {
	registry = append(registry, b)
}

func Detect() (Backend, error) {
	for _, b := range registry {
		if b.Detect() {
			return b, nil
		}
	}
	return nil, fmt.Errorf("no supported battery backend detected")
}

func All() []Backend {
	return registry
}
