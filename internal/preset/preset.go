package preset

import (
	"fmt"

	"github.com/Ooooze/batctl/internal/backend"
)

type Preset struct {
	ID          string
	Name        string
	Description string
	Start       int
	Stop        int
}

var Presets = []Preset{
	{
		ID:          "max-lifespan",
		Name:        "Max Lifespan",
		Description: "Best for battery longevity, keeps charge between 20-80%",
		Start:       20,
		Stop:        80,
	},
	{
		ID:          "balanced",
		Name:        "Balanced",
		Description: "Good balance between capacity and battery health, 40-80%",
		Start:       40,
		Stop:        80,
	},
	{
		ID:          "full-charge",
		Name:        "Full Charge",
		Description: "Charge to 100%, no restrictions",
		Start:       0,
		Stop:        100,
	},
	{
		ID:          "plugged-in",
		Name:        "Plugged In",
		Description: "For mostly-plugged-in use, narrow band at 70-80%",
		Start:       70,
		Stop:        80,
	},
}

func FindByID(id string) (Preset, bool) {
	for _, p := range Presets {
		if p.ID == id {
			return p, true
		}
	}
	return Preset{}, false
}

func AdaptToBackend(p Preset, b backend.Backend) (int, int, error) {
	caps := b.Capabilities()
	start, stop := p.Start, p.Stop

	if len(caps.DiscreteStopVals) > 0 {
		stop = nearestDiscrete(stop, caps.DiscreteStopVals)
	} else {
		stop = clamp(stop, caps.StopRange[0], caps.StopRange[1])
	}

	if caps.StartThreshold {
		start = clamp(start, caps.StartRange[0], caps.StartRange[1])
		if start >= stop {
			start = stop - 1
			if start < caps.StartRange[0] {
				start = caps.StartRange[0]
			}
		}
	} else {
		start = 0
	}

	if err := b.ValidateThresholds(start, stop); err != nil {
		return 0, 0, fmt.Errorf("preset %q not compatible with %s: %w", p.ID, b.Name(), err)
	}

	return start, stop, nil
}

func nearestDiscrete(target int, vals []int) int {
	best := vals[0]
	for _, v := range vals {
		if abs(v-target) < abs(best-target) {
			best = v
		}
	}
	return best
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
