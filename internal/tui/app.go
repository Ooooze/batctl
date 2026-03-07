package tui

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Ooooze/batctl/internal/backend"
	"github.com/Ooooze/batctl/internal/battery"
	"github.com/Ooooze/batctl/internal/persist"
	"github.com/Ooooze/batctl/internal/preset"
)

type clearMessageMsg struct{}

func clearMessageAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}

type field int

const (
	fieldStart field = iota
	fieldStop
	fieldBehaviour
	fieldPreset
	fieldPersist
)

type model struct {
	backend       backend.Backend
	batteries     []string
	activeField   field
	startVal      int
	stopVal       int
	behaviourIdx  int
	behaviourOpts []string
	behaviourCur  string
	presetIdx     int
	message       string
	messageStyle  lipgloss.Style
	width         int
	height        int

	vendorName  string
	productName string
	batInfo     battery.Info
	persistSvc  bool
	persistUdev bool
	persistCfg  *persist.Config
}

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
		activeField:  fieldStart,
		messageStyle: dimStyle,
		vendorName:   backend.DetectVendor(),
		productName:  backend.DetectProductName(),
	}

	m.refreshAll()
	return m, nil
}

func (m *model) refreshAll() {
	m.refreshThresholds()
	m.refreshBatInfo()
	m.refreshPersistStatus()
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

func (m *model) refreshBatInfo() {
	if len(m.batteries) == 0 {
		return
	}
	info, err := battery.ReadInfo(m.batteries[0])
	if err == nil {
		m.batInfo = info
	}
}

func (m *model) refreshPersistStatus() {
	m.persistSvc = persist.ServiceEnabled()
	m.persistUdev = persist.UdevRuleInstalled()
	if cfg, err := persist.LoadConfig(); err == nil {
		m.persistCfg = &cfg
	} else {
		m.persistCfg = nil
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
		return m, nil
	case clearMessageMsg:
		m.message = ""
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	var cmd tea.Cmd

	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		m.prevField()
	case "down", "j":
		m.nextField()

	case "left", "h":
		m.adjustCurrent(-1)
	case "right", "l":
		m.adjustCurrent(1)
	case "H":
		m.adjustCurrent(-5)
	case "L":
		m.adjustCurrent(5)

	case "enter":
		cmd = m.handleEnter()

	case "a":
		cmd = m.applyAndSave()
	case "r":
		m.refreshAll()
		cmd = m.setMessage("Refreshed", successStyle)
	}

	return m, cmd
}

func (m *model) handleEnter() tea.Cmd {
	switch m.activeField {
	case fieldPreset:
		p := preset.Presets[m.presetIdx]
		start, stop, err := preset.AdaptToBackend(p, m.backend)
		if err != nil {
			return m.setMessage(fmt.Sprintf("Preset error: %v", err), errorStyle)
		}
		m.startVal = start
		m.stopVal = stop
		return m.setMessage(fmt.Sprintf("Preset %q: %d%%–%d%% (press 'a' to apply)", p.Name, start, stop), successStyle)
	case fieldPersist:
		return m.togglePersist()
	}
	return nil
}

func (m *model) adjustCurrent(delta int) {
	switch m.activeField {
	case fieldStart, fieldStop, fieldBehaviour:
		m.adjustValue(delta)
	case fieldPreset:
		m.presetIdx += delta
		if m.presetIdx < 0 {
			m.presetIdx = len(preset.Presets) - 1
		}
		if m.presetIdx >= len(preset.Presets) {
			m.presetIdx = 0
		}
	}
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
	case fieldPreset:
		if caps.ChargeBehaviour && len(m.behaviourOpts) > 0 {
			m.activeField = fieldBehaviour
		} else {
			m.activeField = fieldStop
		}
	case fieldPersist:
		m.activeField = fieldPreset
	}
}

func (m *model) nextField() {
	caps := m.backend.Capabilities()
	switch m.activeField {
	case fieldStart:
		m.activeField = fieldStop
	case fieldStop:
		if caps.ChargeBehaviour && len(m.behaviourOpts) > 0 {
			m.activeField = fieldBehaviour
		} else {
			m.activeField = fieldPreset
		}
	case fieldBehaviour:
		m.activeField = fieldPreset
	case fieldPreset:
		m.activeField = fieldPersist
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

func (m *model) applyAndSave() tea.Cmd {
	if len(m.batteries) == 0 {
		return nil
	}
	bat := m.batteries[0]

	if err := m.backend.SetThresholds(bat, m.startVal, m.stopVal); err != nil {
		return m.setMessage(fmt.Sprintf("Error: %v", err), errorStyle)
	}

	caps := m.backend.Capabilities()
	if caps.ChargeBehaviour && m.behaviourCur != "" {
		if err := m.backend.SetChargeBehaviour(bat, m.behaviourCur); err != nil {
			return m.setMessage(fmt.Sprintf("Thresholds set, but behaviour error: %v", err), errorStyle)
		}
	}

	cfg := persist.Config{Battery: bat, Start: m.startVal, Stop: m.stopVal}
	if err := persist.SaveConfig(cfg); err != nil {
		m.refreshBatInfo()
		return m.setMessage(fmt.Sprintf("Applied %d/%d (config save failed: %v)", m.startVal, m.stopVal, errorStyle), errorStyle)
	}

	m.refreshBatInfo()
	m.refreshPersistStatus()
	return m.setMessage(fmt.Sprintf("Applied & saved: %d%%–%d%%", m.startVal, m.stopVal), successStyle)
}

func (m *model) togglePersist() tea.Cmd {
	if m.persistSvc || m.persistUdev {
		if err := persist.RemoveService(); err != nil {
			return m.setMessage(fmt.Sprintf("Error: %v", err), errorStyle)
		}
		if err := persist.RemoveUdevRule(); err != nil {
			return m.setMessage(fmt.Sprintf("Error: %v", err), errorStyle)
		}
		m.refreshPersistStatus()
		return m.setMessage("Persistence disabled", dimStyle)
	}

	if err := persist.InstallService(); err != nil {
		return m.setMessage(fmt.Sprintf("Error: %v (try with sudo)", err), errorStyle)
	}
	if err := persist.InstallUdevRule(); err != nil {
		return m.setMessage(fmt.Sprintf("Error: %v", err), errorStyle)
	}
	m.refreshPersistStatus()
	return m.setMessage("Persistence enabled (systemd + udev)", successStyle)
}

func (m *model) setMessage(msg string, style lipgloss.Style) tea.Cmd {
	m.message = msg
	m.messageStyle = style
	return clearMessageAfter(3 * time.Second)
}

func (m model) View() string {
	return renderDashboard(m)
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
