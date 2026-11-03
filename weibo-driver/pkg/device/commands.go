package device

import (
	"fmt"
)

// SetDPI localiza o índice do DPI e envia o subcomando 0x01
func (d *Device) SetDPI(dpi int) error {
	idx := -1
	for i, v := range DefaultSupportedDPI {
		if v == dpi {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("DPI %d não é suportado pelo hardware. Opções: %v", dpi, DefaultSupportedDPI)
	}

	return d.SetDPIIndex(byte(idx))
}

// SetDPIIndex envia diretamente o índice do nível de DPI (0 = 800, 1 = 1600, etc.)
func (d *Device) SetDPIIndex(index byte) error {
	payload := []byte{CmdDPI, index}
	return d.SendReport(payload)
}

// SetRGB configura o modo de iluminação e as cores
func (d *Device) SetRGB(mode RGBMode, r, g, b byte, brightness, speed byte) error {
	payload := []byte{
		CmdRGB,
		byte(mode),
		r,
		g,
		b,
		brightness,
		speed,
	}
	return d.SendReport(payload)
}

// SendRawCommand permite o envio de comandos arbitrários para engenharia reversa
func (d *Device) SendRawCommand(subcmd byte, params ...byte) error {
	payload := append([]byte{subcmd}, params...)
	return d.SendReport(payload)
}
