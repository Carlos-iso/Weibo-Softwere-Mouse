package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// SniffAllDevices monitora simultaneamente hidraw1, hidraw2 e todos os IDs XInput
func SniffAllDevices() {
	fmt.Println("=================================================================")
	fmt.Println("🔍 WEIBO UNIVERSAL SNIFFER — Diagnóstico de Botões e LED")
	fmt.Println("=================================================================")
	fmt.Println("Monitorando todos os canais de hardware e software do Weibo...")
	fmt.Println("👉 CLIQUE NO BOTÃO SUPERIOR (DPI) AGORA para testar!")
	fmt.Println("(Pressione Ctrl+C para encerrar)")
	fmt.Println("-----------------------------------------------------------------")

	// 1. Monitorar canais HIDRAW (/dev/hidraw1 e /dev/hidraw2)
	hidFiles := make(map[string]*os.File)
	for _, p := range []string{"/dev/hidraw1", "/dev/hidraw2"} {
		f, err := os.OpenFile(p, os.O_RDONLY|syscall.O_NONBLOCK, 0)
		if err == nil {
			hidFiles[p] = f
			defer f.Close()
		}
	}

	stopChan := make(chan struct{})

	// Goroutine para leitura contínua dos canais HIDRAW
	for path, file := range hidFiles {
		devPath := path
		devFile := file
		go func() {
			buf := make([]byte, 64)
			for {
				select {
				case <-stopChan:
					return
				default:
					n, err := devFile.Read(buf)
					if err == nil && n > 0 {
						// Filtra movimento comum de mouse (Report ID 0x01 com apenas dx/dy)
						if buf[0] == 0x01 && n >= 4 && buf[1] == 0 && (buf[2] != 0 || buf[3] != 0) {
							// Ignora movimento puro para não poluir o terminal
							continue
						}
						fmt.Printf("[%s] 📡 CANAL HID BRUTO (%s): % X\n", time.Now().Format("15:04:05.000"), devPath, buf[:n])
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()
	}

	// 2. Monitorar dispositivos XInput (Mouse, Teclado, Consumer Control)
	xinputIDs := []int{9, 10, 11, 12, 13}
	for _, id := range xinputIDs {
		devID := id
		cmd := exec.Command("xinput", "test", strconv.Itoa(devID))
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			continue
		}
		if err := cmd.Start(); err != nil {
			continue
		}
		defer cmd.Process.Kill()

		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				// Ignora movimento puro
				if strings.HasPrefix(line, "motion") {
					continue
				}
				fmt.Printf("[%s] ⚡ EVENTO XINPUT (Device ID %d): %s\n", time.Now().Format("15:04:05.000"), devID, line)
			}
		}()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	close(stopChan)
	fmt.Println("\nSniffer finalizado.")
}
