package system

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// Constantes do subsistema Linux UInput e Input Event
const (
	uinputPath = "/dev/uinput"

	uiSetEvBit   = 0x40045564 // _IOW('U', 100, int)
	uiSetKeyBit  = 0x40045565 // _IOW('U', 101, int)
	uiDevSetup   = 0x405c5503 // _IOW('U', 3, struct uinput_setup)
	uiDevCreate  = 0x5501     // _IO('U', 1)
	uiDevDestroy = 0x5502     // _IO('U', 2)

	evSyn = 0x00
	evKey = 0x01

	synReport = 0x00
)

// Códigos Linux Input Event Keycodes (linux/input-event-codes.h)
const (
	KeyEsc        = 1
	Key1          = 2
	Key2          = 3
	Key3          = 4
	Key4          = 5
	KeyTab        = 15
	KeyQ          = 16
	KeyW          = 17
	KeyE          = 18
	KeyR          = 19
	KeyT          = 20
	KeyEnter      = 28
	KeyLeftCtrl   = 29
	KeyA          = 30
	KeyS          = 31
	KeyD          = 32
	KeyF          = 33
	KeyLeftShift  = 42
	KeyZ          = 44
	KeyX          = 45
	KeyC          = 46
	KeyV          = 47
	KeyLeftAlt    = 56
	KeySpace      = 57
	KeyF1         = 59
	KeyF2         = 60
	KeyF3         = 61
	KeyF4         = 62
	KeyF5         = 63
	KeyF11        = 87
	KeyF12        = 88
	KeyMute       = 113
	KeyVolumeDown = 114
	KeyVolumeUp   = 115
	KeyPlayPause  = 164
	KeyPrevious   = 165
	KeyNext       = 163
	KeyPrint      = 210
)

// KeyMap mapeia nomes amigáveis para códigos Linux evdev
var KeyMap = map[string]uint16{
	"ctrl":        KeyLeftCtrl,
	"control":     KeyLeftCtrl,
	"shift":       KeyLeftShift,
	"alt":         KeyLeftAlt,
	"enter":       KeyEnter,
	"tab":         KeyTab,
	"space":       KeySpace,
	"esc":         KeyEsc,
	"a":           KeyA,
	"c":           KeyC,
	"v":           KeyV,
	"x":           KeyX,
	"z":           KeyZ,
	"f5":          KeyF5,
	"mute":        KeyMute,
	"volume_up":   KeyVolumeUp,
	"volume_down": KeyVolumeDown,
	"play_pause":  KeyPlayPause,
	"next":        KeyNext,
	"prev":        KeyPrevious,
	"print":       KeyPrint,
}

// UInputDevice gerencia o teclado virtual do Weibo Control
type UInputDevice struct {
	file *os.File
}

// NewUInputDevice inicializa e registra o dispositivo virtual no kernel
func NewUInputDevice() (*UInputDevice, error) {
	file, err := os.OpenFile(uinputPath, os.O_WRONLY|syscall.O_NONBLOCK, 0660)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir %s: %w", uinputPath, err)
	}

	fd := file.Fd()

	// Habilita eventos de teclado
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(uiSetEvBit), uintptr(evKey)); errno != 0 {
		file.Close()
		return nil, fmt.Errorf("ioctl UI_SET_EVBIT falhou: %w", errno)
	}

	// Registra todas as teclas mapeadas
	for _, code := range KeyMap {
		_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(uiSetKeyBit), uintptr(code))
	}

	// Setup do dispositivo virtual (struct uinput_setup: 92 bytes)
	// struct input_id (8 bytes: bustype=3 USB, vendor=0x3151, product=0x3020, version=1)
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, uint16(0x03))   // bustype (BUS_USB)
	_ = binary.Write(buf, binary.LittleEndian, uint16(0x3151)) // vendor
	_ = binary.Write(buf, binary.LittleEndian, uint16(0x3020)) // product
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))      // version

	nameBytes := make([]byte, 80)
	copy(nameBytes, "Weibo Virtual Controller")
	buf.Write(nameBytes)

	_ = binary.Write(buf, binary.LittleEndian, uint32(0)) // ff_effects_max

	setupBytes := buf.Bytes()
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(uiDevSetup), uintptr(setupBytesPointer(setupBytes))); errno != 0 {
		file.Close()
		return nil, fmt.Errorf("ioctl UI_DEV_SETUP falhou: %w", errno)
	}

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(uiDevCreate), 0); errno != 0 {
		file.Close()
		return nil, fmt.Errorf("ioctl UI_DEV_CREATE falhou: %w", errno)
	}

	// Pequeno delay para o X11 e udev reconhecerem o novo dispositivo
	time.Sleep(100 * time.Millisecond)

	return &UInputDevice{file: file}, nil
}

func setupBytesPointer(b []byte) uintptr {
	if len(b) == 0 {
		return 0
	}
	return uintptr(unsafe.Pointer(&b[0]))
}

// Close desconecta e destrói o dispositivo virtual
func (u *UInputDevice) Close() error {
	if u.file != nil {
		_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, u.file.Fd(), uintptr(uiDevDestroy), 0)
		err := u.file.Close()
		u.file = nil
		return err
	}
	return nil
}

// emitEvent envia um evento bruto de input_event para o Linux
func (u *UInputDevice) emitEvent(evType uint16, code uint16, value int32) error {
	// struct input_event:
	// struct timeval time (16 bytes em 64-bit: 8 bytes sec + 8 bytes usec)
	// __u16 type
	// __u16 code
	// __s32 value
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, int64(0)) // time.tv_sec
	_ = binary.Write(buf, binary.LittleEndian, int64(0)) // time.tv_usec
	_ = binary.Write(buf, binary.LittleEndian, evType)
	_ = binary.Write(buf, binary.LittleEndian, code)
	_ = binary.Write(buf, binary.LittleEndian, value)

	_, err := u.file.Write(buf.Bytes())
	return err
}

// Sync sincroniza os eventos pendentes com o kernel
func (u *UInputDevice) Sync() error {
	return u.emitEvent(evSyn, synReport, 0)
}

// KeyPress pressiona (1) ou solta (0) uma tecla
func (u *UInputDevice) KeyPress(code uint16, pressed bool) error {
	val := int32(0)
	if pressed {
		val = 1
	}
	if err := u.emitEvent(evKey, code, val); err != nil {
		return err
	}
	return u.Sync()
}

// TapKey pressiona e solta uma tecla com um intervalo
func (u *UInputDevice) TapKey(code uint16, holdDuration time.Duration) error {
	if err := u.KeyPress(code, true); err != nil {
		return err
	}
	time.Sleep(holdDuration)
	return u.KeyPress(code, false)
}

// SendCombo envia uma combinação de modificadores + tecla (ex: Ctrl+C)
func (u *UInputDevice) SendCombo(modifiers []uint16, key uint16) error {
	for _, mod := range modifiers {
		_ = u.KeyPress(mod, true)
	}
	time.Sleep(20 * time.Millisecond)

	_ = u.TapKey(key, 30*time.Millisecond)

	for i := len(modifiers) - 1; i >= 0; i-- {
		_ = u.KeyPress(modifiers[i], false)
	}
	return nil
}

// ExecuteAction interpreta uma string de ação (ex: "ctrl+c", "volume_up", "media_play")
func (u *UInputDevice) ExecuteAction(action string) error {
	action = strings.ToLower(strings.TrimSpace(action))
	parts := strings.Split(action, "+")

	if len(parts) == 1 {
		code, ok := KeyMap[parts[0]]
		if !ok {
			return fmt.Errorf("tecla desconhecida: %s", parts[0])
		}
		return u.TapKey(code, 40*time.Millisecond)
	}

	var mods []uint16
	for i := 0; i < len(parts)-1; i++ {
		code, ok := KeyMap[strings.TrimSpace(parts[i])]
		if !ok {
			return fmt.Errorf("modificador desconhecido: %s", parts[i])
		}
		mods = append(mods, code)
	}

	mainKeyStr := strings.TrimSpace(parts[len(parts)-1])
	mainCode, ok := KeyMap[mainKeyStr]
	if !ok {
		return fmt.Errorf("tecla principal desconhecida: %s", mainKeyStr)
	}

	return u.SendCombo(mods, mainCode)
}
