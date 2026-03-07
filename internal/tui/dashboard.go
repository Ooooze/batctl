package tui

import (
	"fmt"
	"strings"

	"github.com/Ooooze/batctl/internal/preset"
)

func renderDashboard(m model) string {
	var b strings.Builder

	header := titleStyle.Render("batctl") + "  " +
		subtitleStyle.Render("Battery Charge Manager")
	b.WriteString(header + "\n\n")

	if m.productName != "" {
		b.WriteString("  " + valueStyle.Render(m.productName) + "  " +
			dimStyle.Render("·") + "  " +
			subtitleStyle.Render(m.backend.Name()) + "\n\n")
	} else {
		b.WriteString("  " + valueStyle.Render(m.vendorName) + "  " +
			dimStyle.Render("·") + "  " +
			subtitleStyle.Render(m.backend.Name()) + "\n\n")
	}

	// Battery info
	info := m.batInfo
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

	// Thresholds
	b.WriteString("  " + labelStyle.Render("Charge Thresholds") + "\n")
	b.WriteString(renderThresholdLine("Start", m.startVal,
		m.activeField == fieldStart && m.editMode, m.activeField == fieldStart) + "\n")
	b.WriteString(renderThresholdLine("Stop", m.stopVal,
		m.activeField == fieldStop && m.editMode, m.activeField == fieldStop) + "\n")

	caps := m.backend.Capabilities()
	if caps.ChargeBehaviour && len(m.behaviourOpts) > 0 {
		b.WriteString(renderBehaviour(m.behaviourCur, m.behaviourOpts,
			m.activeField == fieldBehaviour, m.editMode) + "\n")
	}

	b.WriteString("\n")

	// Presets
	b.WriteString("  " + labelStyle.Render("Presets") + "\n")
	b.WriteString(renderPresetRow(m.presetIdx, m.activeField == fieldPreset) + "\n")

	b.WriteString("\n")

	// Persistence
	b.WriteString("  " + labelStyle.Render("Persistence") + "\n")
	b.WriteString(renderPersistRow(m) + "\n")

	b.WriteString("\n")

	// Help
	b.WriteString(renderHelpBar(m))

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

func renderPresetRow(activeIdx int, focused bool) string {
	var parts []string
	for i, p := range preset.Presets {
		label := fmt.Sprintf("%s (%d–%d)", p.Name, p.Start, p.Stop)
		if i == activeIdx {
			if focused {
				parts = append(parts, selectedStyle.Render("["+label+"]"))
			} else {
				parts = append(parts, accentStyle.Render("["+label+"]"))
			}
		} else {
			parts = append(parts, dimStyle.Render(label))
		}
	}

	prefix := "  "
	if focused {
		prefix = accentStyle.Render("▸ ")
	}

	return prefix + strings.Join(parts, "  ")
}

func renderPersistRow(m model) string {
	prefix := "  "
	if m.activeField == fieldPersist {
		prefix = accentStyle.Render("▸ ")
	}

	var status string
	if m.persistSvc && m.persistUdev {
		status = successStyle.Render("● enabled") + dimStyle.Render(" (systemd + udev)")
	} else if m.persistSvc {
		status = successStyle.Render("● enabled") + dimStyle.Render(" (systemd)")
	} else {
		status = dimStyle.Render("○ disabled")
	}

	cfgStr := ""
	if m.persistCfg != nil {
		cfgStr = dimStyle.Render(fmt.Sprintf("  ·  %d/%d on %s",
			m.persistCfg.Start, m.persistCfg.Stop, m.persistCfg.Battery))
	}

	hint := ""
	if m.activeField == fieldPersist {
		hint = dimStyle.Render("  [enter to toggle]")
	}

	return prefix + status + cfgStr + hint
}

func renderHelpBar(m model) string {
	type helpItem struct {
		key  string
		desc string
	}

	items := []helpItem{
		{"↑↓", "navigate"},
		{"←→", "adjust"},
		{"enter", "select/edit"},
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
