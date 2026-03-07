package persist

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	serviceName = "batctl.service"
	servicePath = "/etc/systemd/system/" + serviceName
)

const serviceTemplate = `[Unit]
Description=Apply battery charge thresholds (batctl)
After=multi-user.target

[Service]
Type=oneshot
ExecStart=/usr/bin/batctl apply
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
`

func InstallService() error {
	if err := os.WriteFile(servicePath, []byte(serviceTemplate), 0644); err != nil {
		return fmt.Errorf("writing service file: %w", err)
	}
	if err := systemctl("daemon-reload"); err != nil {
		return err
	}
	return systemctl("enable", serviceName)
}

func RemoveService() error {
	_ = systemctl("disable", serviceName)
	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing service file: %w", err)
	}
	return systemctl("daemon-reload")
}

func ServiceEnabled() bool {
	out, err := exec.Command("systemctl", "is-enabled", serviceName).Output()
	if err != nil {
		return false
	}
	return len(out) > 0 && out[0] == 'e' // "enabled"
}

func systemctl(args ...string) error {
	cmd := exec.Command("systemctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl %v: %w", args, err)
	}
	return nil
}
