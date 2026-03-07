package persist

import (
	"fmt"
	"os"
)

const (
	udevRuleName = "99-batctl-resume.rules"
	udevRulePath = "/etc/udev/rules.d/" + udevRuleName
)

const udevRuleContent = `# Restore battery charge thresholds after resume from suspend
ACTION=="change", SUBSYSTEM=="power_supply", ATTR{type}=="Battery", RUN+="/usr/bin/batctl apply"
`

func InstallUdevRule() error {
	if err := os.WriteFile(udevRulePath, []byte(udevRuleContent), 0644); err != nil {
		return fmt.Errorf("writing udev rule: %w", err)
	}
	return nil
}

func RemoveUdevRule() error {
	if err := os.Remove(udevRulePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing udev rule: %w", err)
	}
	return nil
}

func UdevRuleInstalled() bool {
	_, err := os.Stat(udevRulePath)
	return err == nil
}
