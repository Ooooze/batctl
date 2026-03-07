# batctl — Battery Charge Threshold Manager

TUI and CLI tool for managing battery charge thresholds on Linux laptops.
Supports 14+ laptop vendors with automatic hardware detection.

## Supported Vendors

| Vendor | Start | Stop | Charge Behaviour | Driver |
|--------|-------|------|-------------------|--------|
| ThinkPad | 0–99 | 1–100 | Yes | thinkpad_acpi |
| ASUS | — | 0–100 | — | asus_wmi |
| Dell | 50–95 | 55–100 | — | dell_laptop |
| Lenovo IdeaPad | — | on/off | — | ideapad_laptop |
| Samsung | — | 80/100 | — | samsung_laptop |
| LG | — | 80/100 | — | lg_laptop |
| Huawei | 0–99 | 1–100 | — | huawei_wmi |
| MSI | auto | 10–100 | — | msi_ec |
| Framework | 0–99* | 1–100 | Yes | cros_charge-control |
| System76 | 0–99 | 1–100 | — | system76_acpi |
| Sony | — | 50/80/100 | — | sony_laptop |
| Toshiba | — | 80/100 | — | toshiba_acpi |
| Tuxedo | discrete | discrete | — | clevo_acpi |
| Apple Silicon | auto | 80/100 | — | macsmc_power |
| Generic | 0–99* | 1–100* | Yes* | any with sysfs |

## Installation

### From source

```bash
make
sudo make install
```

### Arch Linux (AUR)

```bash
# Manual
makepkg -si

# Or with an AUR helper
yay -S batctl-tui
```

## Usage

### TUI (interactive)

```bash
sudo batctl
```

### CLI

```bash
# Show battery status and current thresholds
batctl status

# Set thresholds
sudo batctl set --start 40 --stop 80

# Apply a preset
sudo batctl set --preset max-lifespan

# Detect hardware
batctl detect

# Enable persistence (survives reboot + suspend)
sudo batctl persist enable

# Check persistence status
batctl persist status

# Disable persistence
sudo batctl persist disable
```

## Presets

| Preset | Start | Stop | Use Case |
|--------|-------|------|----------|
| max-lifespan | 20% | 80% | Best for battery longevity |
| balanced | 40% | 80% | Good balance |
| full-charge | 0% | 100% | No restrictions |
| plugged-in | 70% | 80% | Mostly plugged in |

Presets are automatically adapted to your hardware's capabilities.

## Persistence

When enabled via `batctl persist enable`:
- A systemd service applies thresholds on boot
- A udev rule restores thresholds after resume from suspend
- Config stored in `/etc/batctl.conf`

## Requirements

- Linux kernel with appropriate vendor driver
- Root access for writing thresholds
