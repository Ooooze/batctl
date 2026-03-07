package persist

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	resumeServiceName = "batctl-resume.service"
	resumeServicePath = "/etc/systemd/system/" + resumeServiceName
)

const resumeServiceTemplate = `[Unit]
Description=Restore battery charge thresholds after resume (batctl)
After=suspend.target hibernate.target hybrid-sleep.target suspend-then-hibernate.target

[Service]
Type=oneshot
ExecStart=/usr/bin/batctl apply

[Install]
WantedBy=suspend.target hibernate.target hybrid-sleep.target suspend-then-hibernate.target
`

func InstallResumeService() error {
	removeLegacyUdevRule()
	if err := os.WriteFile(resumeServicePath, []byte(resumeServiceTemplate), 0644); err != nil {
		return fmt.Errorf("writing resume service: %w", err)
	}
	if err := systemctl("daemon-reload"); err != nil {
		return err
	}
	return systemctl("enable", resumeServiceName)
}

func RemoveResumeService() error {
	removeLegacyUdevRule()
	_ = systemctl("disable", resumeServiceName)
	if err := os.Remove(resumeServicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing resume service: %w", err)
	}
	return systemctl("daemon-reload")
}

func removeLegacyUdevRule() {
	legacyPath := "/etc/udev/rules.d/99-batctl-resume.rules"
	if _, err := os.Stat(legacyPath); err == nil {
		os.Remove(legacyPath)
		exec.Command("udevadm", "control", "--reload-rules").Run()
	}
}

func ResumeServiceEnabled() bool {
	out, err := exec.Command("systemctl", "is-enabled", resumeServiceName).Output()
	if err != nil {
		return false
	}
	return len(out) > 0 && out[0] == 'e'
}
