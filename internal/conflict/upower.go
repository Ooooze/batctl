package conflict

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type UPowerInfo struct {
	Available      bool
	Supported      bool
	Enabled        bool
	StartThreshold int
	EndThreshold   int
}

var runDBusGet = defaultDBusGet

func CheckUPower(bat string) UPowerInfo {
	devPath := "/org/freedesktop/UPower/devices/battery_" + bat
	info := UPowerInfo{}

	supported, err := getDBusBool(devPath, "ChargeThresholdSupported")
	if err != nil {
		return info
	}
	info.Available = true
	info.Supported = supported

	if !supported {
		return info
	}

	if enabled, err := getDBusBool(devPath, "ChargeThresholdEnabled"); err == nil {
		info.Enabled = enabled
	}
	if v, err := getDBusUint(devPath, "ChargeEndThreshold"); err == nil {
		info.EndThreshold = v
	}
	if v, err := getDBusUint(devPath, "ChargeStartThreshold"); err == nil {
		info.StartThreshold = v
	}

	return info
}

func getDBusBool(devPath, prop string) (bool, error) {
	out, err := runDBusGet(devPath, prop)
	if err != nil {
		return false, err
	}
	return parseBool(out)
}

func getDBusUint(devPath, prop string) (int, error) {
	out, err := runDBusGet(devPath, prop)
	if err != nil {
		return 0, err
	}
	return parseUint(out)
}

func parseBool(output string) (bool, error) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "variant") {
			if strings.Contains(line, "boolean true") {
				return true, nil
			}
			if strings.Contains(line, "boolean false") {
				return false, nil
			}
		}
	}
	return false, fmt.Errorf("no boolean value found in output")
}

func parseUint(output string) (int, error) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "variant") {
			fields := strings.Fields(line)
			if len(fields) >= 3 && fields[1] == "uint32" {
				return strconv.Atoi(fields[2])
			}
		}
	}
	return 0, fmt.Errorf("no uint32 value found in output")
}

func defaultDBusGet(devPath, prop string) (string, error) {
	cmd := exec.Command("dbus-send",
		"--system", "--print-reply",
		"--dest=org.freedesktop.UPower",
		devPath,
		"org.freedesktop.DBus.Properties.Get",
		"string:org.freedesktop.UPower.Device",
		"string:"+prop,
	)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
