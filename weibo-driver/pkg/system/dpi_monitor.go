package system

import (
	"os"
	"sync"
	"syscall"
	"time"
)

// DPIClickHandler é chamado quando o botão físico de DPI é pressionado
type DPIClickHandler func()

// DPIMonitor escuta /dev/hidraw2 para capturar o evento do botão de DPI
type DPIMonitor struct {
	stopChan chan struct{}
	running  bool
	mu       sync.Mutex
	handler  DPIClickHandler
}

// NewDPIMonitor cria um monitor de cliques para o botão superior
func NewDPIMonitor(handler DPIClickHandler) *DPIMonitor {
	return &DPIMonitor{
		handler: handler,
	}
}

// Start inicia a leitura em background de /dev/hidraw2
func (m *DPIMonitor) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true
	m.stopChan = make(chan struct{})
	m.mu.Unlock()

	go func() {
		file, err := os.OpenFile("/dev/hidraw2", os.O_RDONLY|syscall.O_NONBLOCK, 0)
		if err != nil {
			return
		}
		defer file.Close()

		buf := make([]byte, 64)
		lastTrigger := time.Now()

		for {
			select {
			case <-m.stopChan:
				return
			default:
				n, err := file.Read(buf)
				if err == nil && n >= 8 {
					// Assinatura do botão de DPI:
					// Report ID = 0x01
					// Buttons mask = 0x00 (sem clique nos botões 1 a 5)
					// Motion bytes (2 a 7) == 0 (sem movimento)
					if buf[0] == 0x01 && buf[1] == 0x00 {
						allZero := true
						for i := 2; i < 8 && i < n; i++ {
							if buf[i] != 0 {
								allZero = false
								break
							}
						}

						if allZero {
							// Debounce de 300ms para evitar disparos duplicados
							if time.Since(lastTrigger) > 300*time.Millisecond {
								lastTrigger = time.Now()
								if m.handler != nil {
									m.handler()
								}
							}
						}
					}
				}
				time.Sleep(15 * time.Millisecond)
			}
		}
	}()

	return nil
}

// Stop finaliza a escuta
func (m *DPIMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.running {
		close(m.stopChan)
		m.running = false
	}
}
