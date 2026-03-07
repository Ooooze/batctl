package battery

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const PowerSupplyBase = "/sys/class/power_supply"

func SysfsReadString(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}

func SysfsReadInt(path string) (int, error) {
	s, err := SysfsReadString(path)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parsing int from %s (%q): %w", path, s, err)
	}
	return v, nil
}

func SysfsWriteInt(path string, value int) error {
	return SysfsWriteString(path, strconv.Itoa(value))
}

func SysfsWriteString(path string, value string) error {
	if err := os.WriteFile(path, []byte(value), 0644); err != nil {
		return fmt.Errorf("writing %q to %s: %w", value, path, err)
	}
	return nil
}

func SysfsExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func BatPath(bat, attr string) string {
	return filepath.Join(PowerSupplyBase, bat, attr)
}

func DMIRead(attr string) string {
	s, _ := SysfsReadString(filepath.Join("/sys/class/dmi/id", attr))
	return s
}

func ListBatteries() []string {
	entries, err := os.ReadDir(PowerSupplyBase)
	if err != nil {
		return nil
	}
	var bats []string
	for _, e := range entries {
		typePath := filepath.Join(PowerSupplyBase, e.Name(), "type")
		t, err := SysfsReadString(typePath)
		if err != nil {
			continue
		}
		if strings.EqualFold(t, "battery") {
			bats = append(bats, e.Name())
		}
	}
	return bats
}
