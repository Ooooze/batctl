package backend

import "fmt"

type Capabilities struct {
	StartThreshold    bool
	StopThreshold     bool
	ChargeBehaviour   bool
	StartRange        [2]int // min, max
	StopRange         [2]int // min, max
	DiscreteStopVals  []int  // for vendors with fixed stop values (Sony: 50/80/100)
	StartAutoComputed bool   // MSI: start = stop - 10 automatically
}

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
