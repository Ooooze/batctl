package battery

import "fmt"

type Info struct {
	Name           string
	Status         string // Charging, Discharging, Not charging, Full, Unknown
	Capacity       int    // percent 0-100
	CycleCount     int
	EnergyNow      float64 // Wh
	EnergyFull     float64 // Wh
	EnergyDesign   float64 // Wh
	PowerNow       float64 // W
	Manufacturer   string
	Model          string
	Technology     string
	HealthPercent  float64 // energy_full / energy_full_design * 100
}

func ReadInfo(bat string) (Info, error) {
	info := Info{Name: bat}

	status, err := SysfsReadString(BatPath(bat, "status"))
	if err != nil {
		return info, fmt.Errorf("battery %s not found: %w", bat, err)
	}
	info.Status = status

	info.Capacity, _ = SysfsReadInt(BatPath(bat, "capacity"))
	info.CycleCount, _ = SysfsReadInt(BatPath(bat, "cycle_count"))
	info.Manufacturer, _ = SysfsReadString(BatPath(bat, "manufacturer"))
	info.Model, _ = SysfsReadString(BatPath(bat, "model_name"))
	info.Technology, _ = SysfsReadString(BatPath(bat, "technology"))

	info.EnergyNow = readMicro(BatPath(bat, "energy_now"))
	info.EnergyFull = readMicro(BatPath(bat, "energy_full"))
	info.EnergyDesign = readMicro(BatPath(bat, "energy_full_design"))
	info.PowerNow = readMicro(BatPath(bat, "power_now"))

	if info.EnergyFull == 0 {
		info.EnergyNow = readMicro(BatPath(bat, "charge_now"))
		info.EnergyFull = readMicro(BatPath(bat, "charge_full"))
		info.EnergyDesign = readMicro(BatPath(bat, "charge_full_design"))
		info.PowerNow = readMicro(BatPath(bat, "current_now"))
	}

	if info.EnergyDesign > 0 {
		info.HealthPercent = info.EnergyFull / info.EnergyDesign * 100
	}

	return info, nil
}

func readMicro(path string) float64 {
	v, err := SysfsReadInt(path)
	if err != nil {
		return 0
	}
	return float64(v) / 1_000_000
}
