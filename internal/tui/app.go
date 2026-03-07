package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/spaceclam/batctl/internal/backend"
	"github.com/spaceclam/batctl/internal/battery"
	"github.com/spaceclam/batctl/internal/persist"
	"github.com/spaceclam/batctl/internal/preset"
)

type view int

const (
	viewDashboard view = iota
	viewPresets
)

type field int

const (
	fieldStart field = iota
	fieldStop
	fieldBehaviour
)

type model struct {
	backend      backend.Backend
	batteries    []string
	activeView   view
	activeField  field
	editMode     bool
	startVal     int
	stopVal      int
	behaviourIdx int
	behaviourOpts []string
	behaviourCur string
	presetIdx    int
	message      string
	messageStyle lipgloss.Style
	width        int
	height       int
}

type refreshMsg struct{}

func initialModel() (model, error) {
	b, err := backend.Detect()
	if err != nil {
		return model{}, err
	}

	bats := battery.ListBatteries()
	if len(bats) == 0 {
		return model{}, fmt.Errorf("no batteries found")
	}

	m := model{
		backend:      b,
		batteries:    bats,
		activeView:   viewDashboard,
		activeField:  fieldStart,
		messageStyle: dimStyle,
	}

	m.refreshThresholds()
	return m, nil
}

func (m *model) refreshThresholds() {
	if len(m.batteries) == 0 {
		return
	}
	start, stop, err := m.backend.GetThresholds(m.batteries[0])
	if err == nil {
		m.startVal = start
		m.stopVal = stop
	}

	caps := m.backend.Capabilities()
	if caps.ChargeBehaviour {
		cur, avail, err := m.backend.GetChargeBehaviour(m.batteries[0])
		if err == nil {
			m.behaviourCur = cur
			m.behaviourOpts = avail
			for i, o := range avail {
				if o == cur {
					m.behaviourIdx = i
					break
				}
			}
		}
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch m.activeView {
		case viewDashboard:
			return m.updateDashboard(msg)
		case viewPresets:
			return m.updatePresets(msg)
		}
	}
	return m, nil
}

func (m model) updateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if !m.editMode {
			m.prevField()
		}

	case "down", "j":
		if !m.editMode {
			m.nextField()
		}

	case "enter":
		m.editMode = !m.editMode

	case "esc":
		if m.editMode {
			m.editMode = false
			m.refreshThresholds()
		}

	case "left", "h":
		if m.editMode {
			m.adjustValue(-1)
		}

	case "right", "l":
		if m.editMode {
			m.adjustValue(1)
		}

	case "H":
		if m.editMode {
			m.adjustValue(-5)
		}

	case "L":
		if m.editMode {
			m.adjustValue(5)
		}

	case "p":
		if !m.editMode {
			m.activeView = viewPresets
			m.presetIdx = 0
		}

	case "a":
		if !m.editMode {
			m.applyThresholds()
		}

	case "s":
		if !m.editMode {
			m.saveAndPersist()
		}

	case "r":
		if !m.editMode {
			m.refreshThresholds()
			m.setMessage("Refreshed", successStyle)
		}
	}

	return m, nil
}

func (m model) updatePresets(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.activeView = viewDashboard

	case "up", "k":
		if m.presetIdx > 0 {
			m.presetIdx--
		}

	case "down", "j":
		if m.presetIdx < len(preset.Presets)-1 {
			m.presetIdx++
		}

	case "enter":
		p := preset.Presets[m.presetIdx]
		start, stop, err := preset.AdaptToBackend(p, m.backend)
		if err != nil {
			m.setMessage(fmt.Sprintf("Preset error: %v", err), errorStyle)
		} else {
			m.startVal = start
			m.stopVal = stop
			m.setMessage(fmt.Sprintf("Preset %q loaded: %d%%–%d%% (press 'a' to apply)", p.Name, start, stop), successStyle)
		}
		m.activeView = viewDashboard
	}

	return m, nil
}

func (m *model) prevField() {
	caps := m.backend.Capabilities()
	switch m.activeField {
	case fieldStop:
		if caps.StartThreshold {
			m.activeField = fieldStart
		}
	case fieldBehaviour:
		m.activeField = fieldStop
	}
}

func (m *model) nextField() {
	caps := m.backend.Capabilities()
	switch m.activeField {
	case fieldStart:
		m.activeField = fieldStop
	case fieldStop:
		if caps.ChargeBehaviour {
			m.activeField = fieldBehaviour
		}
	}
}

func (m *model) adjustValue(delta int) {
	caps := m.backend.Capabilities()

	switch m.activeField {
	case fieldStart:
		if !caps.StartThreshold {
			return
		}
		m.startVal += delta
		if m.startVal < caps.StartRange[0] {
			m.startVal = caps.StartRange[0]
		}
		if m.startVal > caps.StartRange[1] {
			m.startVal = caps.StartRange[1]
		}
		if m.startVal >= m.stopVal {
			m.startVal = m.stopVal - 1
		}

	case fieldStop:
		if len(caps.DiscreteStopVals) > 0 {
			m.stopVal = nextDiscrete(m.stopVal, delta, caps.DiscreteStopVals)
		} else {
			m.stopVal += delta
			if m.stopVal < caps.StopRange[0] {
				m.stopVal = caps.StopRange[0]
			}
			if m.stopVal > caps.StopRange[1] {
				m.stopVal = caps.StopRange[1]
			}
		}

	case fieldBehaviour:
		if len(m.behaviourOpts) > 0 {
			m.behaviourIdx += delta
			if m.behaviourIdx < 0 {
				m.behaviourIdx = 0
			}
			if m.behaviourIdx >= len(m.behaviourOpts) {
				m.behaviourIdx = len(m.behaviourOpts) - 1
			}
			m.behaviourCur = m.behaviourOpts[m.behaviourIdx]
		}
	}
}

func (m *model) applyThresholds() {
	if len(m.batteries) == 0 {
		return
	}
	bat := m.batteries[0]

	if err := m.backend.SetThresholds(bat, m.startVal, m.stopVal); err != nil {
		m.setMessage(fmt.Sprintf("Error: %v", err), errorStyle)
		return
	}

	caps := m.backend.Capabilities()
	if caps.ChargeBehaviour && m.behaviourCur != "" {
		if err := m.backend.SetChargeBehaviour(bat, m.behaviourCur); err != nil {
			m.setMessage(fmt.Sprintf("Thresholds set, but behaviour error: %v", err), errorStyle)
			return
		}
	}

	m.setMessage(fmt.Sprintf("Applied: start=%d%% stop=%d%%", m.startVal, m.stopVal), successStyle)
}

func (m *model) saveAndPersist() {
	cfg := persist.Config{
		Battery: m.batteries[0],
		Start:   m.startVal,
		Stop:    m.stopVal,
	}

	if err := persist.SaveConfig(cfg); err != nil {
		m.setMessage(fmt.Sprintf("Save error: %v (try with sudo)", err), errorStyle)
		return
	}

	if err := persist.InstallService(); err != nil {
		m.setMessage(fmt.Sprintf("Service install error: %v (try with sudo)", err), errorStyle)
		return
	}

	if err := persist.InstallUdevRule(); err != nil {
		m.setMessage(fmt.Sprintf("Udev rule error: %v", err), errorStyle)
		return
	}

	m.setMessage(fmt.Sprintf("Saved & persistence enabled: %d%%–%d%%", m.startVal, m.stopVal), successStyle)
}

func (m *model) setMessage(msg string, style lipgloss.Style) {
	m.message = msg
	m.messageStyle = style
}

func (m model) View() string {
	switch m.activeView {
	case viewPresets:
		return renderPresetPicker(m)
	default:
		return renderDashboard(m)
	}
}

func nextDiscrete(current, direction int, vals []int) int {
	idx := 0
	for i, v := range vals {
		if v == current {
			idx = i
			break
		}
		if v > current {
			idx = i
			break
		}
	}
	idx += direction
	if idx < 0 {
		idx = 0
	}
	if idx >= len(vals) {
		idx = len(vals) - 1
	}
	return vals[idx]
}

func Run() error {
	m, err := initialModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
