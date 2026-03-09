<p align="center">
  <h1 align="center">⚡ batctl</h1>
  <p align="center">
    <strong>Battery charge threshold manager for Linux laptops</strong>
  </p>
  <p align="center">
    <a href="#installation">Installation</a> •
    <a href="#usage">Usage</a> •
    <a href="#supported-hardware">Hardware</a> •
    <a href="#presets">Presets</a> •
    <a href="#persistence">Persistence</a>
  </p>
</p>

<br>

**batctl** is a terminal UI and CLI tool that lets you control battery charge thresholds on Linux.
Set start/stop charge levels to extend battery lifespan, choose from built-in presets,
and persist your settings across reboots — all from a single, zero-dependency binary.

<p align="center">
  <img src="demo.gif" alt="batctl demo" width="800">
</p>

## Why batctl?

Most laptops support charge thresholds in hardware, but the Linux interface is fragmented:
each vendor exposes different sysfs paths, value ranges, and quirks.
Tools like TLP are powerful but heavy and config-file-driven.

**batctl** gives you:

- **One binary, zero config** — auto-detects your hardware and shows what's possible
- **Interactive TUI** — see battery health, adjust thresholds with arrow keys, pick presets
- **Scriptable CLI** — `batctl set --stop 80` for automation and dotfiles
- **16+ vendor backends** — from ThinkPad to Apple Silicon, with a generic fallback
- **Persistence** — survives reboots and suspend/resume via systemd

## Installation

### Quick install (any distro)

```bash
curl -fsSL https://raw.githubusercontent.com/Ooooze/batctl/master/install.sh | sudo bash
```

To uninstall:

```bash
curl -fsSL https://raw.githubusercontent.com/Ooooze/batctl/master/install.sh | sudo bash -s -- --uninstall
```

### From source

```bash
git clone https://github.com/Ooooze/batctl.git
cd batctl
make
sudo make install
```

### Arch Linux / Omarchy (AUR)

```bash
yay -S batctl-tui
```

Or manually:

```bash
makepkg -si
```

### Pre-built binary

Download from [Releases](https://github.com/Ooooze/batctl/releases), then:

```bash
chmod +x batctl
sudo cp batctl /usr/bin/
```

## Usage

### Interactive TUI

Launch the full terminal interface (requires root for write operations):

```bash
sudo batctl
```

**Controls:**

| Key | Action |
|-----|--------|
| `↑` `↓` / `j` `k` | Navigate between fields |
| `←` `→` / `h` `l` | Adjust value (±1) |
| `H` `L` | Adjust value (±5) |
| `Enter` | Toggle edit mode |
| `Esc` | Cancel edit |
| `p` | Open preset picker |
| `a` | Apply current thresholds |
| `s` | Save config + enable persistence |
| `r` | Refresh battery info |
| `q` | Quit |

### CLI Commands

```bash
# Show battery info, thresholds, and persistence status
batctl status

# Detect hardware and show capabilities
batctl detect

# Set thresholds directly
sudo batctl set --start 40 --stop 80

# Apply a built-in preset
sudo batctl set --preset max-lifespan

# Apply thresholds from config (used by systemd on boot)
sudo batctl apply

# Enable persistence (systemd services: boot + resume)
sudo batctl persist enable

# Check persistence status
batctl persist status

# Disable persistence and clean up
sudo batctl persist disable
```

### Example: `batctl status`

```
Backend: ThinkPad

BAT0 (Sunwoda 5B10W51867)
  Status:     Charging
  Capacity:   85%
  Health:     103.6%
  Cycles:     54
  Energy:     48.3 / 52.8 Wh (design: 51.0 Wh)
  Power:      20.1 W
  Thresholds: start=40% stop=80%
  Behaviour:  auto (available: auto, inhibit-charge, force-discharge)

Persistence:  boot=true  resume=true
```

### Example: `batctl detect`

```
Vendor:  LENOVO
Product: 21AH00FGRT
Backend: ThinkPad
Capabilities:
  Start threshold:    true (range: 0..99)
  Stop threshold:     true (range: 1..100)
  Charge behaviour:   true
Batteries: [BAT0]
```

## Presets

Built-in presets adapt automatically to your hardware's supported value ranges:

| Preset | Start | Stop | Description |
|--------|------:|-----:|-------------|
| `max-lifespan` | 20% | 80% | Best for battery longevity. Ideal if you're mostly plugged in. |
| `balanced` | 40% | 80% | Good mix of available capacity and long-term health. |
| `plugged-in` | 70% | 80% | Narrow band for always-connected workstations. |
| `full-charge` | 0% | 100% | No restrictions. Use when you need maximum runtime. |

```bash
sudo batctl set --preset balanced
```

> On hardware with limited options (e.g. Samsung with only 80/100),
> presets snap to the nearest supported value.

## Supported Hardware

batctl auto-detects your laptop vendor via DMI and probes sysfs for the right driver.
If no specific backend matches, the **generic fallback** is used for any laptop
exposing standard `charge_control_{start,end}_threshold` files.

| Vendor | Start | Stop | Behaviour | Kernel Driver |
|--------|:-----:|:----:|:---------:|---------------|
| **Acer** | — | 80 or 100 | — | `acer-wmi-battery` |
| **Lenovo ThinkPad** | 0–99 | 1–100 | ✓ | `thinkpad_acpi` |
| **ASUS** | — | 0–100¹ | — | `asus_wmi` |
| **Dell** | 50–95 | 55–100 | — | `dell_laptop` |
| **Lenovo IdeaPad** | — | on/off² | — | `ideapad_laptop` |
| **Huawei MateBook** | 0–99 | 1–100 | — | `huawei_wmi` |
| **Samsung** | — | 80 or 100 | — | `samsung_laptop` |
| **LG Gram** | — | 80 or 100 | — | `lg_laptop` |
| **MSI** | auto³ | 10–100 | — | `msi_ec` |
| **Framework** | 0–99⁴ | 1–100 | ✓ | `cros_charge-control` |
| **System76** | 0–99 | 1–100 | — | `system76_acpi` |
| **Sony VAIO** | — | 50/80/100 | — | `sony_laptop` |
| **Toshiba/Dynabook** | — | 80 or 100 | — | `toshiba_acpi` |
| **Tuxedo (Clevo)** | discrete⁵ | discrete⁵ | — | `clevo_acpi` |
| **Apple Silicon** | auto³ | 80 or 100 | — | `macsmc_power` |
| **Microsoft Surface** | 0–99⁶ | 1–100 | — | `surface_battery` |
| **Generic fallback** | 0–99 | 1–100 | ✓ | any sysfs |

<sup>¹ Some ASUS models only accept 40, 60, or 80</sup><br>
<sup>² Conservation mode: fixed threshold (usually 60%)</sup><br>
<sup>³ Start threshold is computed by hardware from stop value</sup><br>
<sup>⁴ Start threshold requires EC firmware v3</sup><br>
<sup>⁵ Tuxedo start: 40/50/60/70/80/95 — stop: 60/70/80/90/100</sup>
<sup>⁶ Requires linux-surface kernel; start threshold availability varies by model</sup>

## Persistence

By default, charge thresholds reset on reboot or resume from suspend.
batctl solves this with a one-command setup:

```bash
sudo batctl persist enable
```

This installs:

| Component | Path | Purpose |
|-----------|------|---------|
| Config file | `/etc/batctl.conf` | Stores your threshold values |
| Boot service | `/etc/systemd/system/batctl.service` | Applies thresholds on boot |
| Resume service | `/etc/systemd/system/batctl-resume.service` | Restores thresholds after suspend/resume |

To disable and remove everything:

```bash
sudo batctl persist disable
```

## How It Works

```
batctl
├── Reads /sys/class/dmi/id/sys_vendor → identifies laptop vendor
├── Probes sysfs paths → confirms driver availability
├── Selects matching backend (or generic fallback)
├── Reads/writes /sys/class/power_supply/BAT*/charge_control_*
└── Manages systemd services for persistence
```

All operations go through the kernel's standard sysfs interface.
No direct hardware access, no custom kernel modules required.

## Architecture

```
batctl/
├── cmd/batctl/          → CLI entry point (cobra)
├── internal/
│   ├── backend/         → 16 vendor backends + generic + auto-detection
│   ├── battery/         → sysfs read/write helpers, battery info
│   ├── persist/         → systemd services, config file
│   ├── preset/          → built-in presets with hardware adaptation
│   └── tui/             → bubbletea TUI (dashboard, presets, styles)
├── configs/             → systemd service templates
├── Makefile
└── PKGBUILD             → Arch Linux package
```

## Requirements

- **Linux** with a kernel that includes your laptop's battery driver
- **Root access** (`sudo`) for writing thresholds and managing persistence
- No runtime dependencies — single static binary

## Contributing

Contributions are welcome! Areas where help is especially appreciated:

- **New vendor backends** — if your laptop isn't detected, check `batctl detect` output and open a PR
- **Testing** — try batctl on your hardware and report what works
- **Packaging** — help with Fedora, Debian, NixOS packages

## Donate

If batctl saved your battery some cycles, consider buying me a coffee in crypto:

| Currency | Network | Address |
|----------|---------|---------|
| **BTC** | Bitcoin | `bc1qflyxz75wkcyet89cttanyv7ws98lf8wjezdydq` |
| **ETH** | Ethereum | `0xAfAA1CEdb10ECfC696C9984e857c813CB1871b4C` |
| **USDT** | TRC-20 | `TSgwHUf6tiuJgFmaerb3TyTjggoP5cPecb` |
| **TON** | TON | `UQA1SYTIdmH7iPNcUtbOtQXtCLPzTYeR8YxCdxksU8HMhkSe` |

## License

MIT
