package device

// Identificadores de hardware do Weibo / YICHIP
const (
	VendorID        = 0x3151
	ProductID       = 0x3020
	VendorUsagePage = 0xFF00
	VendorUsage     = 0x000E

	ReportIDVendor  = 0xBA
	ReportPayloadLen = 31
	ReportPacketLen  = 32 // 1 byte ID (0xBA) + 31 bytes payload
)

// Subcomandos conhecidos no protocolo YICHIP
const (
	CmdDPI = 0x01
	CmdRGB = 0x02
)

// RGBMode define os modos de iluminação LED
type RGBMode byte

const (
	RGBModeOff       RGBMode = 0x00
	RGBModeStatic    RGBMode = 0x01
	RGBModeBreathing RGBMode = 0x02
	RGBModeRainbow   RGBMode = 0x03
	RGBModeWave      RGBMode = 0x04
	RGBModeReactive  RGBMode = 0x05
)

func (m RGBMode) String() string {
	switch m {
	case RGBModeOff:
		return "Off"
	case RGBModeStatic:
		return "Static"
	case RGBModeBreathing:
		return "Breathing"
	case RGBModeRainbow:
		return "Rainbow"
	case RGBModeWave:
		return "Wave"
	case RGBModeReactive:
		return "Reactive"
	default:
		return "Unknown"
	}
}

// Níveis de DPI suportados por padrão pelo hardware YICHIP
var DefaultSupportedDPI = []int{800, 1600, 3200, 4800, 6000}
