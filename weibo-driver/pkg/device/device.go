package device

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	// IOCTL USBDEVFS_RESET = _IO('U', 20) = 0x5514
	usbDevfsReset = 0x5514
)

// Device gerencia a comunicação direta com o receptor do mouse Weibo via hidraw
type Device struct {
	hidrawPath string
	file       *os.File
	isMock     bool
}

// DeviceInfo contém informações de diagnóstico sobre a conexão
type DeviceInfo struct {
	Connected  bool
	HidrawPath string
	IsMock     bool
	VID        uint16
	PID        uint16
}

// NewDevice inicializa o controlador do mouse.
func NewDevice(mock bool) *Device {
	return &Device{
		isMock: mock,
	}
}

// FindHidrawPath escaneia /sys/class/hidraw para localizar o canal do receptor YICHIP (interface 1)
func FindHidrawPath() (string, error) {
	entries, err := os.ReadDir("/sys/class/hidraw")
	if err != nil {
		return "", fmt.Errorf("erro ao ler /sys/class/hidraw: %w", err)
	}

	for _, entry := range entries {
		ueventPath := filepath.Join("/sys/class/hidraw", entry.Name(), "device", "uevent")
		content, err := os.ReadFile(ueventPath)
		if err != nil {
			continue
		}

		str := string(content)
		if strings.Contains(str, "3151") && strings.Contains(str, "3020") {
			if strings.Contains(str, "input1") {
				devPath := filepath.Join("/dev", entry.Name())
				return devPath, nil
			}
		}
	}

	return "", fmt.Errorf("receptor Weibo (3151:3020 / input1) não encontrado em /dev/hidraw*")
}

// findUSBDevPath descobre o caminho /dev/bus/usb/BBB/DDD para o dongle
func findUSBDevPath(hidrawPath string) (string, error) {
	if hidrawPath == "" {
		var err error
		hidrawPath, err = FindHidrawPath()
		if err != nil {
			return "", err
		}
	}

	devName := filepath.Base(hidrawPath)
	realPath, err := filepath.EvalSymlinks(filepath.Join("/sys/class/hidraw", devName, "device"))
	if err != nil {
		return "", fmt.Errorf("não foi possível resolver symlink sysfs: %w", err)
	}

	// 2 níveis acima: HID -> USB Interface -> USB Device
	usbDir := filepath.Dir(filepath.Dir(realPath))

	busData, err := os.ReadFile(filepath.Join(usbDir, "busnum"))
	if err != nil {
		return "", fmt.Errorf("não foi possível ler busnum: %w", err)
	}
	devData, err := os.ReadFile(filepath.Join(usbDir, "devnum"))
	if err != nil {
		return "", fmt.Errorf("não foi possível ler devnum: %w", err)
	}

	busNum, err := strconv.Atoi(strings.TrimSpace(string(busData)))
	if err != nil {
		return "", err
	}
	devNum, err := strconv.Atoi(strings.TrimSpace(string(devData)))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("/dev/bus/usb/%03d/%03d", busNum, devNum), nil
}

// Connect estabelece a conexão com o hardware
func (d *Device) Connect() error {
	if d.isMock {
		return nil
	}

	path, err := FindHidrawPath()
	if err != nil {
		return err
	}

	d.hidrawPath = path
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("falha ao abrir %s: %w (verifique permissões)", path, err)
	}

	d.file = file
	return nil
}

// Close encerra a conexão com o dispositivo
func (d *Device) Close() error {
	if d.file != nil {
		err := d.file.Close()
		d.file = nil
		return err
	}
	return nil
}

// ResetUSB executa USBDEVFS_RESET para destravar o endpoint USB após timeout de RF
func (d *Device) ResetUSB() error {
	if d.isMock {
		return nil
	}

	path := d.hidrawPath
	if path == "" {
		var err error
		path, err = FindHidrawPath()
		if err != nil {
			return err
		}
	}

	usbPath, err := findUSBDevPath(path)
	if err != nil {
		return fmt.Errorf("erro ao localizar nó USB: %w", err)
	}

	_ = d.Close()

	usbFile, err := os.OpenFile(usbPath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("falha ao abrir %s para reset: %w", usbPath, err)
	}
	defer usbFile.Close()

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, usbFile.Fd(), uintptr(usbDevfsReset), 0)
	if errno != 0 {
		return fmt.Errorf("falha no ioctl USBDEVFS_RESET: %w", errno)
	}

	// Aguarda o kernel re-enumerar o dongle
	time.Sleep(800 * time.Millisecond)

	return d.Connect()
}

// Info retorna o status de diagnóstico da conexão
func (d *Device) Info() DeviceInfo {
	return DeviceInfo{
		Connected:  d.isMock || d.file != nil,
		HidrawPath: d.hidrawPath,
		IsMock:     d.isMock,
		VID:        VendorID,
		PID:        ProductID,
	}
}

// SendReport envia um pacote Output Report com Report ID 0xBA
func (d *Device) SendReport(payload []byte) error {
	if d.isMock {
		return nil
	}

	if d.file == nil {
		if err := d.Connect(); err != nil {
			return fmt.Errorf("dispositivo não conectado: %w", err)
		}
	}

	if len(payload) > ReportPayloadLen {
		return fmt.Errorf("payload excede %d bytes (tamanho fornecido: %d)", ReportPayloadLen, len(payload))
	}

	packet := make([]byte, ReportPacketLen)
	packet[0] = ReportIDVendor
	copy(packet[1:], payload)

	n, err := d.file.Write(packet)
	if err != nil {
		// Se houve timeout ou erro de protocolo, tenta resetar automaticamente para não deixar o USB travado
		isTimeoutOrProto := strings.Contains(err.Error(), "timed out") || strings.Contains(err.Error(), "protocol error")
		if isTimeoutOrProto {
			_ = d.ResetUSB()
		}
		return fmt.Errorf("erro ao gravar relatório no hidraw: %w", err)
	}

	if n != len(packet) {
		return fmt.Errorf("gravação incompleta: enviou %d de %d bytes", n, len(packet))
	}

	return nil
}

// ReadReport tenta ler um Input Report com timeout estipulado
func (d *Device) ReadReport(timeout time.Duration) ([]byte, error) {
	if d.isMock {
		return make([]byte, ReportPacketLen), nil
	}

	if d.file == nil {
		return nil, fmt.Errorf("dispositivo não conectado")
	}

	fd := int(d.file.Fd())

	tv := syscall.NsecToTimeval(timeout.Nanoseconds())
	var readFds syscall.FdSet
	readFds.Bits[fd/64] |= 1 << (uint(fd) % 64)

	n, err := syscall.Select(fd+1, &readFds, nil, nil, &tv)
	if err != nil {
		return nil, fmt.Errorf("erro no select hidraw: %w", err)
	}

	if n == 0 {
		return nil, nil
	}

	buf := make([]byte, 64)
	bytesRead, err := d.file.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler hidraw: %w", err)
	}

	return buf[:bytesRead], nil
}
