package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Ooooze/batctl/internal/backend"
	"github.com/Ooooze/batctl/internal/battery"
	"github.com/Ooooze/batctl/internal/persist"
	"github.com/Ooooze/batctl/internal/preset"
	"github.com/Ooooze/batctl/internal/tui"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "batctl",
		Short: "Battery charge threshold manager",
		Long:  "TUI and CLI tool for managing battery charge thresholds on Linux laptops.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Run()
		},
		SilenceUsage: true,
	}

	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(setCmd())
	rootCmd.AddCommand(applyCmd())
	rootCmd.AddCommand(persistCmd())
	rootCmd.AddCommand(detectCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show battery info and current thresholds",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := backend.Detect()
			if err != nil {
				return err
			}

			bats := battery.ListBatteries()
			if len(bats) == 0 {
				return fmt.Errorf("no batteries found")
			}

			fmt.Printf("Backend: %s\n", b.Name())
			caps := b.Capabilities()

			for _, bat := range bats {
				info, err := battery.ReadInfo(bat)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: %s: %v\n", bat, err)
					continue
				}

				fmt.Printf("\n%s (%s %s)\n", info.Name, info.Manufacturer, info.Model)
				fmt.Printf("  Status:     %s\n", info.Status)
				fmt.Printf("  Capacity:   %d%%\n", info.Capacity)
				fmt.Printf("  Health:     %.1f%%\n", info.HealthPercent)
				fmt.Printf("  Cycles:     %d\n", info.CycleCount)
				fmt.Printf("  Energy:     %.1f / %.1f Wh (design: %.1f Wh)\n",
					info.EnergyNow, info.EnergyFull, info.EnergyDesign)
				if info.PowerNow > 0 {
					fmt.Printf("  Power:      %.1f W\n", info.PowerNow)
				}

				start, stop, err := b.GetThresholds(bat)
				if err != nil {
					fmt.Printf("  Thresholds: error reading (%v)\n", err)
				} else {
					fmt.Printf("  Thresholds: start=%d%% stop=%d%%\n", start, stop)
				}

				if caps.ChargeBehaviour {
					cur, avail, err := b.GetChargeBehaviour(bat)
					if err == nil {
						fmt.Printf("  Behaviour:  %s (available: %s)\n", cur, strings.Join(avail, ", "))
					}
				}
			}

			fmt.Printf("\nPersistence:  systemd=%v  udev=%v\n",
				persist.ServiceEnabled(), persist.UdevRuleInstalled())

			return nil
		},
	}
}

func setCmd() *cobra.Command {
	var startVal, stopVal int
	var presetName string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set battery charge thresholds",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := backend.Detect()
			if err != nil {
				return err
			}

			bats := battery.ListBatteries()
			if len(bats) == 0 {
				return fmt.Errorf("no batteries found")
			}
			bat := bats[0]

			if presetName != "" {
				p, ok := preset.FindByID(presetName)
				if !ok {
					return fmt.Errorf("unknown preset %q (available: max-lifespan, balanced, full-charge, plugged-in)", presetName)
				}
				startVal, stopVal, err = preset.AdaptToBackend(p, b)
				if err != nil {
					return err
				}
				fmt.Printf("Applying preset %q (adapted: start=%d, stop=%d)\n", p.Name, startVal, stopVal)
			}

			if err := b.ValidateThresholds(startVal, stopVal); err != nil {
				return fmt.Errorf("invalid thresholds: %w", err)
			}

			if err := b.SetThresholds(bat, startVal, stopVal); err != nil {
				return fmt.Errorf("setting thresholds: %w", err)
			}

			fmt.Printf("Thresholds set: start=%d%% stop=%d%% on %s\n", startVal, stopVal, bat)
			return nil
		},
	}

	cmd.Flags().IntVar(&startVal, "start", 0, "Start charge threshold (%%)")
	cmd.Flags().IntVar(&stopVal, "stop", 100, "Stop charge threshold (%%)")
	cmd.Flags().StringVar(&presetName, "preset", "", "Apply a named preset (max-lifespan, balanced, full-charge, plugged-in)")

	return cmd
}

func applyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "Apply thresholds from /etc/batctl.conf (used by systemd)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := persist.LoadConfig()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			b, err := backend.Detect()
			if err != nil {
				return err
			}

			if err := b.SetThresholds(cfg.Battery, cfg.Start, cfg.Stop); err != nil {
				return fmt.Errorf("applying thresholds: %w", err)
			}

			fmt.Printf("Applied: start=%d%% stop=%d%% on %s\n", cfg.Start, cfg.Stop, cfg.Battery)
			return nil
		},
	}
}

func persistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "persist [enable|disable|status]",
		Short: "Manage systemd persistence for charge thresholds",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "enable":
				b, err := backend.Detect()
				if err != nil {
					return err
				}
				bats := battery.ListBatteries()
				if len(bats) == 0 {
					return fmt.Errorf("no batteries found")
				}
				start, stop, err := b.GetThresholds(bats[0])
				if err != nil {
					return fmt.Errorf("reading current thresholds: %w", err)
				}
				if err := persist.SaveConfig(persist.Config{
					Battery: bats[0], Start: start, Stop: stop,
				}); err != nil {
					return fmt.Errorf("saving config: %w", err)
				}
				if err := persist.InstallService(); err != nil {
					return fmt.Errorf("installing systemd service: %w", err)
				}
				if err := persist.InstallUdevRule(); err != nil {
					return fmt.Errorf("installing udev rule: %w", err)
				}
				fmt.Printf("Persistence enabled: start=%d%% stop=%d%% on %s\n", start, stop, bats[0])
				fmt.Println("Systemd service and udev rule installed.")
				return nil

			case "disable":
				if err := persist.RemoveService(); err != nil {
					return fmt.Errorf("removing systemd service: %w", err)
				}
				if err := persist.RemoveUdevRule(); err != nil {
					return fmt.Errorf("removing udev rule: %w", err)
				}
				fmt.Println("Persistence disabled. Systemd service and udev rule removed.")
				return nil

			case "status":
				svc := persist.ServiceEnabled()
				udev := persist.UdevRuleInstalled()
				fmt.Printf("Systemd service: %s\n", enabledStr(svc))
				fmt.Printf("Udev rule:       %s\n", enabledStr(udev))
				if cfg, err := persist.LoadConfig(); err == nil {
					fmt.Printf("Config:          battery=%s start=%d%% stop=%d%%\n",
						cfg.Battery, cfg.Start, cfg.Stop)
				} else {
					fmt.Printf("Config:          not found (%v)\n", err)
				}
				return nil

			default:
				return fmt.Errorf("unknown action %q (use enable, disable, or status)", args[0])
			}
		},
	}
	return cmd
}

func detectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "detect",
		Short: "Show detected vendor and capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			vendor := backend.DetectVendor()
			product := backend.DetectProductName()
			fmt.Printf("Vendor:  %s\n", vendor)
			fmt.Printf("Product: %s\n", product)

			b, err := backend.Detect()
			if err != nil {
				fmt.Printf("Backend: none detected (%v)\n", err)
				return nil
			}

			fmt.Printf("Backend: %s\n", b.Name())
			caps := b.Capabilities()
			fmt.Printf("Capabilities:\n")
			fmt.Printf("  Start threshold:    %v", caps.StartThreshold)
			if caps.StartThreshold {
				fmt.Printf(" (range: %d..%d)", caps.StartRange[0], caps.StartRange[1])
			}
			fmt.Println()
			fmt.Printf("  Stop threshold:     %v", caps.StopThreshold)
			if caps.StopThreshold {
				if len(caps.DiscreteStopVals) > 0 {
					fmt.Printf(" (values: %v)", caps.DiscreteStopVals)
				} else {
					fmt.Printf(" (range: %d..%d)", caps.StopRange[0], caps.StopRange[1])
				}
			}
			fmt.Println()
			fmt.Printf("  Charge behaviour:   %v\n", caps.ChargeBehaviour)
			if caps.StartAutoComputed {
				fmt.Printf("  Start auto-computed: yes\n")
			}

			bats := battery.ListBatteries()
			fmt.Printf("Batteries: %v\n", bats)

			return nil
		},
	}
}

func enabledStr(v bool) string {
	if v {
		return "enabled"
	}
	return "disabled"
}
