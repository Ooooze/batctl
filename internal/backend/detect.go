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
	{"LENOVO", "Hint: try loading the kernel module (included in kernel):\n" +
		"      ThinkPad:       sudo modprobe thinkpad_acpi\n" +
		"      IdeaPad/Yoga:   sudo modprobe ideapad_laptop"},
	{"ASUSTeK", "Hint: try loading the kernel module (included in kernel):\n" +
		"      sudo modprobe asus-nb-wmi"},
	{"Dell", "Hint: try loading the kernel module (included in kernel):\n" +
		"      sudo modprobe dell-laptop"},
	{"HUAWEI", "Hint: try loading the kernel module (included in kernel):\n" +
		"      sudo modprobe huawei-wmi"},
	{"SAMSUNG", "Hint: try loading the kernel module (included in kernel):\n" +
		"      sudo modprobe samsung-laptop"},
	{"LG", "Hint: try loading the kernel module (included in kernel):\n" +
		"      sudo modprobe lg-laptop"},
	{"Micro-Star", "Hint: try loading the kernel module (included in kernel 6.4+):\n" +
		"      sudo modprobe msi-ec\n" +
		"      On older kernels: yay -S msi-ec-dkms"},
	{"Acer", "Hint: this backend requires a separate kernel module (not in mainline).\n" +
		"      On Arch: yay -S acer-wmi-battery-dkms && sudo modprobe acer-wmi-battery"},
	{"Framework", "Hint: try loading the kernel module (included in kernel):\n" +
		"      sudo modprobe cros_ec_lpcs"},
	{"System76", "Hint: try loading the kernel module (included in kernel):\n" +
		"      sudo modprobe system76_acpi\n" +
		"      Or install from AUR: yay -S system76-acpi-dkms"},
	{"Sony", "Hint: try loading the kernel module (included in kernel):\n" +
		"      sudo modprobe sony-laptop"},
	{"TOSHIBA", "Hint: try loading the kernel module (included in kernel):\n" +
		"      sudo modprobe toshiba_acpi"},
	{"Dynabook", "Hint: try loading the kernel module (included in kernel):\n" +
		"      sudo modprobe toshiba_acpi"},
	{"TUXEDO", "Hint: this backend requires a separate kernel module (not in mainline).\n" +
		"      On Arch: yay -S tuxedo-drivers-dkms && sudo modprobe tuxedo_keyboard"},
	{"Apple", "Hint: try loading the kernel module:\n" +
		"      Apple Silicon: sudo modprobe macsmc-battery (requires Asahi Linux kernel)\n" +
		"      Intel Mac:     apply applesmc-next patches (see github.com/c---/applesmc-next)"},
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
	Register(&SurfaceBackend{})
	// Generic must be last — it's the fallback
	Register(&GenericBackend{})
}
