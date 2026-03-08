package backend

import (
	"strings"

	"github.com/Ooooze/batctl/internal/battery"
)

func DetectVendor() string {
	return strings.TrimSpace(battery.DMIRead("sys_vendor"))
}

func DetectProductName() string {
	return strings.TrimSpace(battery.DMIRead("product_name"))
}

type vendorHint struct {
	vendorSubstring string
	message         string
}

var vendorHints = []vendorHint{
	{"LENOVO", "Hint: try loading the kernel module:\n" +
		"      ThinkPad:       sudo modprobe thinkpad_acpi\n" +
		"      IdeaPad/Yoga:   sudo modprobe ideapad_laptop"},
	{"ASUSTeK", "Hint: try loading the kernel module: sudo modprobe asus-nb-wmi"},
	{"Dell", "Hint: try loading the kernel module: sudo modprobe dell-laptop"},
	{"HUAWEI", "Hint: try loading the kernel module: sudo modprobe huawei-wmi"},
	{"SAMSUNG", "Hint: try loading the kernel module: sudo modprobe samsung-laptop"},
	{"LG", "Hint: try loading the kernel module: sudo modprobe lg-laptop"},
	{"Micro-Star", "Hint: try loading the kernel module: sudo modprobe msi-ec\n" +
		"      Requires kernel 6.4+. On older kernels install msi-ec-dkms."},
	{"Acer", "Hint: this backend requires the acer-wmi-battery module (not in mainline kernel).\n" +
		"      On Arch: yay -S acer-wmi-battery-dkms && sudo modprobe acer-wmi-battery"},
	{"Framework", "Hint: try loading the kernel module: sudo modprobe cros_ec_lpcs"},
	{"System76", "Hint: try loading the kernel module: sudo modprobe system76_acpi"},
	{"Sony", "Hint: try loading the kernel module: sudo modprobe sony-laptop"},
	{"TOSHIBA", "Hint: try loading the kernel module: sudo modprobe toshiba_acpi"},
	{"Dynabook", "Hint: try loading the kernel module: sudo modprobe toshiba_acpi"},
	{"TUXEDO", "Hint: this backend requires tuxedo-drivers (not in mainline kernel).\n" +
		"      See: https://github.com/tuxedocomputers/tuxedo-drivers"},
	{"Apple", "Hint: try loading the kernel module: sudo modprobe macsmc-battery\n" +
		"      Requires Asahi Linux kernel."},
}

func DetectHint() string {
	vendor := DetectVendor()
	for _, h := range vendorHints {
		if strings.Contains(vendor, h.vendorSubstring) {
			return h.message
		}
	}
	return ""
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
	Register(&AcerBackend{})
	Register(&FrameworkBackend{})
	Register(&System76Backend{})
	Register(&SonyBackend{})
	Register(&ToshibaBackend{})
	Register(&TuxedoBackend{})
	Register(&AppleBackend{})
	// Generic must be last — it's the fallback
	Register(&GenericBackend{})
}
