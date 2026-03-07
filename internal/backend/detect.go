package backend

import (
	"strings"

	"github.com/spaceclam/batctl/internal/battery"
)

func DetectVendor() string {
	return strings.TrimSpace(battery.DMIRead("sys_vendor"))
}

func DetectProductName() string {
	return strings.TrimSpace(battery.DMIRead("product_name"))
}

func init() {
	Register(&ThinkPadBackend{})
	Register(&ASUSBackend{})
	Register(&DellBackend{})
	Register(&LenovoIdeaPadBackend{})
	Register(&HuaweiBackend{})
	Register(&SamsungBackend{})
	Register(&LGBackend{})
	Register(&MSIBackend{})
	Register(&FrameworkBackend{})
	Register(&System76Backend{})
	Register(&SonyBackend{})
	Register(&ToshibaBackend{})
	Register(&TuxedoBackend{})
	Register(&AppleBackend{})
	// Generic must be last — it's the fallback
	Register(&GenericBackend{})
}
