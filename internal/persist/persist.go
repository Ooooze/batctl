package persist

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const ConfigPath = "/etc/batctl.conf"

type Config struct {
	Battery string
	Start   int
	Stop    int
}

func LoadConfig() (Config, error) {
	f, err := os.Open(ConfigPath)
	if err != nil {
		return Config{}, fmt.Errorf("opening config: %w", err)
	}
	defer f.Close()

	cfg := Config{Battery: "BAT0"}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "BATTERY":
			cfg.Battery = val
		case "START_THRESHOLD":
			cfg.Start, _ = strconv.Atoi(val)
		case "STOP_THRESHOLD":
			cfg.Stop, _ = strconv.Atoi(val)
		}
	}
	return cfg, scanner.Err()
}

func SaveConfig(cfg Config) error {
	content := fmt.Sprintf("# batctl configuration\n# Managed by batctl — do not edit manually\nBATTERY=%s\nSTART_THRESHOLD=%d\nSTOP_THRESHOLD=%d\n",
		cfg.Battery, cfg.Start, cfg.Stop)
	return os.WriteFile(ConfigPath, []byte(content), 0644)
}
