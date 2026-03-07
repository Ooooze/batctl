package tui

import (
	"fmt"
	"strings"

	"github.com/spaceclam/batctl/internal/preset"
)

func renderPresetPicker(m model) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Select Preset") + "\n\n")

	for i, p := range preset.Presets {
		cursor := "  "
		nameStyle := dimStyle
		if i == m.presetIdx {
			cursor = accentStyle.Render("▸ ")
			nameStyle = selectedStyle
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, nameStyle.Render(p.Name)))
		b.WriteString(fmt.Sprintf("    %s  (%d%%–%d%%)\n",
			dimStyle.Render(p.Description), p.Start, p.Stop))
		b.WriteString("\n")
	}

	b.WriteString("\n" + helpKeyStyle.Render("enter") + " " + helpDescStyle.Render("select") +
		"  " + helpKeyStyle.Render("esc") + " " + helpDescStyle.Render("back"))

	return b.String()
}
