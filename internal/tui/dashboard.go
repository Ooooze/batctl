package tui

import (
	"fmt"
	"strings"

	"github.com/spaceclam/batctl/internal/backend"
	"github.com/spaceclam/batctl/internal/battery"
	"github.com/spaceclam/batctl/internal/persist"
)

func renderDashboard(m model) string {
	var b strings.Builder

	header := titleStyle.Render("batctl") + "  " +
		subtitleStyle.Render("Battery Charge Manager")
	b.WriteString(header + "\n\n")

	product := backend.DetectProductName()
	vendor := backend.DetectVendor()
	if product != "" {
		b.WriteString("  " + valueStyle.Render(product) + "  " +
			dimStyle.Render("·") + "  " +
			subtitleStyle.Render(m.backend.Name()) + "\n\n")
	} else {
		b.WriteString("  " + valueStyle.Render(vendor) + "  " +
			dimStyle.Render("·") + "  " +
			subtitleStyle.Render(m.backend.Name()) + "\n\n")
	}

	for _, bat := range m.batteries {
		info, err := battery.ReadInfo(bat)
		if err != nil {
			b.WriteString(errorStyle.Render(fmt.Sprintf("  %s: %v", bat, err)) + "\n")
			continue
		}

		b.WriteString("  " + accentStyle.Render(info.Name) + "\n")
		b.WriteString("  " + renderGauge(info.Capacity, 30) + "  " +
			valueStyle.Render(fmt.Sprintf("%d%%", info.Capacity)) + "  " +
			renderStatus(info.Status) + "\n")

		healthStr := fmt.Sprintf("%.1f%%", info.HealthPercent)
		energyStr := fmt.Sprintf("%.1f / %.1f Wh", info.EnergyNow, info.EnergyFull)
		b.WriteString("  " + dimStyle.Render(fmt.Sprintf("Health: %s  ·  Cycles: %d  ·  %s",
			healthStr, info.CycleCount, energyStr)) + "\n")
		if info.PowerNow > 0 {
			b.WriteString("  " + dimStyle.Render(fmt.Sprintf("Power: %.1f W", info.PowerNow)) + "\n")
		}

		b.WriteString("\n")

		start, stop, err := m.backend.GetThresholds(bat)
		if err != nil {
			b.WriteString("  " + errorStyle.Render(fmt.Sprintf("Thresholds: %v", err)) + "\n")
		} else {
			b.WriteString("  " + labelStyle.Render("Charge Thresholds") + "\n")

			startLine := renderThresholdLine("Start", start, m.activeField == fieldStart && m.editMode, m.activeField == fieldStart)
			stopLine := renderThresholdLine("Stop", stop, m.activeField == fieldStop && m.editMode, m.activeField == fieldStop)
			b.WriteString(startLine + "\n")
			b.WriteString(stopLine + "\n")
		}

		caps := m.backend.Capabilities()
		if caps.ChargeBehaviour {
			cur, avail, err := m.backend.GetChargeBehaviour(bat)
			if err == nil {
				b.WriteString(renderBehaviour(cur, avail, m.activeField == fieldBehaviour, m.editMode) + "\n")
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(renderHelpBar(m))
	b.WriteString("\n")

	svcStatus := persist.ServiceEnabled()
	udevStatus := persist.UdevRuleInstalled()
	persistStr := "disabled"
	if svcStatus && udevStatus {
		persistStr = successStyle.Render("enabled (systemd + udev)")
	} else if svcStatus {
		persistStr = successStyle.Render("enabled (systemd)")
	}

	cfgStr := ""
	if cfg, err := persist.LoadConfig(); err == nil {
		cfgStr = fmt.Sprintf("  ·  Config: %d/%d", cfg.Start, cfg.Stop)
	}

	b.WriteString(statusBarStyle.Render(
		fmt.Sprintf("  Persistence: %s%s", persistStr, cfgStr)))

	if m.message != "" {
		b.WriteString("\n" + m.messageStyle.Render("  "+m.message))
	}

	return b.String()
}

func renderGauge(percent, width int) string {
	filled := percent * width / 100
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return gaugeFullStyle.Render(strings.Repeat("█", filled)) +
		gaugeEmptyStyle.Render(strings.Repeat("░", width-filled))
}

func renderStatus(status string) string {
	switch status {
	case "Charging":
		return successStyle.Render("⚡ Charging")
	case "Discharging":
		return accentStyle.Render("🔋 Discharging")
	case "Full":
		return successStyle.Render("✓ Full")
	case "Not charging":
		return dimStyle.Render("⏸ Not charging")
	default:
		return dimStyle.Render(status)
	}
}

func renderThresholdLine(label string, value int, editing bool, focused bool) string {
	const barWidth = 30
	pos := value * barWidth / 100
	if pos >= barWidth {
		pos = barWidth - 1
	}

	var bar strings.Builder
	for i := 0; i < barWidth; i++ {
		if i == pos {
			if editing {
				bar.WriteString(selectedStyle.Render("●"))
			} else {
				bar.WriteString(accentStyle.Render("●"))
			}
		} else {
			bar.WriteString(dimStyle.Render("─"))
		}
	}

	prefix := "  "
	if focused {
		prefix = accentStyle.Render("▸ ")
	}

	valStr := fmt.Sprintf("%d%%", value)
	if editing {
		valStr = selectedStyle.Render(valStr)
	}

	return fmt.Sprintf("%s%-6s ◄%s► %s", prefix, label+":", bar.String(), valStr)
}

func renderBehaviour(current string, available []string, focused bool, editing bool) string {
	var parts []string
	for _, mode := range available {
		if mode == current {
			if editing && focused {
				parts = append(parts, selectedStyle.Render("["+mode+"]"))
			} else {
				parts = append(parts, accentStyle.Render("["+mode+"]"))
			}
		} else {
			parts = append(parts, dimStyle.Render(mode))
		}
	}

	prefix := "  "
	if focused {
		prefix = accentStyle.Render("▸ ")
	}

	return prefix + "Behaviour: " + strings.Join(parts, " ")
}

func renderHelpBar(m model) string {
	type helpItem struct {
		key  string
		desc string
	}

	items := []helpItem{
		{"↑↓", "navigate"},
		{"←→", "adjust"},
		{"enter", "edit"},
		{"p", "presets"},
		{"a", "apply"},
		{"s", "save"},
		{"r", "refresh"},
		{"q", "quit"},
	}

	var parts []string
	for _, item := range items {
		parts = append(parts, helpKeyStyle.Render(item.key)+" "+helpDescStyle.Render(item.desc))
	}

	return "  " + strings.Join(parts, "  ")
}
