package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"weibo-driver/pkg/config"
	"weibo-driver/pkg/device"
	"weibo-driver/pkg/system"
)

func printHelp() {
	fmt.Println("Weibo Control Driver - Ferramenta de Linha de Comando")
	fmt.Println("\nUso:")
	fmt.Println("  weibo-cli [flags] <comando> [argumentos...]")
	fmt.Println("\nComandos de Sensibilidade e DPI:")
	fmt.Println("  dpi <valor>           Define o DPI real no Linux (ex: 800, 1200, 1600, 2400, 3200)")
	fmt.Println("  dpi-scale <fator>     Define o multiplicador exato da matriz (ex: 0.8, 1.0, 1.5, 2.0)")
	fmt.Println("  dpi-status            Exibe a sensibilidade e escala ativas no sistema")
	fmt.Println("\nComandos de Teclas, Macros e Botões:")
	fmt.Println("  macro <atalho>        Emite tecla/atalho via /dev/uinput (ex: ctrl+c, ctrl+v, volume_up)")
	fmt.Println("  listen-buttons        Monitora cliques dos botões físicos do mouse ao vivo")
	fmt.Println("  sniff                 Monitora simultaneamente TODOS os canais (HIDRAW + XInput) para achar o botão DPI")
	fmt.Println("  daemon                Executa serviço em segundo plano para atalhos e botões laterais")
	fmt.Println("\nDiagnóstico de Hardware:")
	fmt.Println("  info                  Exibe diagnóstico do dispositivo e canal hidraw")
	fmt.Println("  reset                 Executa reset de baixo nível no dongle USB")
	fmt.Println("  config                Exibe a configuração persistente em JSON")
	fmt.Println("\nFlags:")
	fmt.Println("  -mock                 Executa em modo simulação sem tocar no hardware real")
}

func main() {
	mockFlag := flag.Bool("mock", false, "Executa em modo simulação")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printHelp()
		return
	}

	cmd := strings.ToLower(args[0])

	switch cmd {
	case "config":
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Erro ao carregar configurações: %v\n", err)
			os.Exit(1)
		}
		path, _ := config.GetConfigPath()
		fmt.Printf("Arquivo de configuração: %s\n", path)
		fmt.Printf("Perfil ativo: %s\n", cfg.ActiveProfileID)
		activeProfile := cfg.Profiles[cfg.ActiveProfileID]
		fmt.Printf("DPI atual: %d | Níveis: %v\n", activeProfile.ActiveDPI, activeProfile.DPIStages)
		fmt.Printf("Mapeamento de Botões: %+v\n", activeProfile.Buttons)
		return

	case "info":
		dev := device.NewDevice(*mockFlag)
		_ = dev.Connect()
		defer dev.Close()

		info := dev.Info()
		fmt.Println("=== Diagnóstico Weibo Control ===")
		fmt.Printf("Conectado Dongle USB: %v\n", info.Connected)
		fmt.Printf("VID:PID:              0x%04X:0x%04X\n", info.VID, info.PID)
		fmt.Printf("Canal Hidraw:         %s\n", info.HidrawPath)

		if xid, err := system.FindMouseDeviceID(); err == nil {
			scale, dpi, _ := system.GetCurrentSensitivity()
			fmt.Printf("Dispositivo XInput:   Encontrado (ID=%d)\n", xid)
			fmt.Printf("Sensibilidade Atual:  %.2fx (equivalente a ~%d DPI)\n", scale, dpi)
		} else {
			fmt.Printf("Dispositivo XInput:   %v\n", err)
		}

	case "reset":
		dev := device.NewDevice(*mockFlag)
		fmt.Println("Enviando sinal USBDEVFS_RESET para o receptor USB...")
		if err := dev.ResetUSB(); err != nil {
			fmt.Printf("Erro ao resetar USB: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[SUCESSO] Dongle USB resetado e reconectado!")

	case "dpi":
		if len(args) < 2 {
			fmt.Println("Erro: informe o valor de DPI. Ex: weibo-cli dpi 1600")
			os.Exit(1)
		}
		dpiVal, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Erro: valor de DPI inválido.")
			os.Exit(1)
		}

		scale := system.DPIToScale(dpiVal)
		fmt.Printf("Aplicando DPI %d no sistema (escala %.2fx na matriz)...\n", dpiVal, scale)

		if err := system.ApplySensitivity(dpiVal); err != nil {
			fmt.Printf("Erro ao aplicar sensibilidade no sistema: %v\n", err)
			os.Exit(1)
		}

		// Persiste a escolha no perfil
		if cfg, err := config.Load(); err == nil {
			prof := cfg.Profiles[cfg.ActiveProfileID]
			prof.ActiveDPI = dpiVal
			cfg.Profiles[cfg.ActiveProfileID] = prof
			_ = config.Save(cfg)
		}

		fmt.Printf("[SUCESSO] Sensibilidade atualizada para %d DPI (escala %.2fx)!\n", dpiVal, scale)

	case "dpi-scale":
		if len(args) < 2 {
			fmt.Println("Erro: informe o multiplicador. Ex: weibo-cli dpi-scale 1.5")
			os.Exit(1)
		}
		scaleVal, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			fmt.Println("Erro: valor inválido.")
			os.Exit(1)
		}

		xid, err := system.FindMouseDeviceID()
		if err != nil {
			fmt.Printf("Erro: %v\n", err)
			os.Exit(1)
		}

		if err := system.SetCoordinateScale(xid, scaleVal); err != nil {
			fmt.Printf("Erro ao aplicar: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("[SUCESSO] Escala de coordenadas definida para %.4fx (~%d DPI)!\n", scaleVal, system.ScaleToDPI(scaleVal))

	case "dpi-status":
		scale, dpi, err := system.GetCurrentSensitivity()
		if err != nil {
			fmt.Printf("Erro: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Sensibilidade ativa: %.4fx (~%d DPI nominal)\n", scale, dpi)

	case "macro":
		if len(args) < 2 {
			fmt.Println("Erro: informe o atalho ou tecla. Ex: weibo-cli macro ctrl+c")
			os.Exit(1)
		}
		shortcut := args[1]
		fmt.Printf("Inicializando /dev/uinput para injetar '%s'...\n", shortcut)

		uinputDev, err := system.NewUInputDevice()
		if err != nil {
			fmt.Printf("Erro ao abrir uinput: %v\n", err)
			os.Exit(1)
		}
		defer uinputDev.Close()

		if err := uinputDev.ExecuteAction(shortcut); err != nil {
			fmt.Printf("Erro ao emitir macro: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[SUCESSO] Tecla/Atalho '%s' emitido no sistema com sucesso!\n", shortcut)

	case "listen-buttons":
		fmt.Println("Iniciando monitoramento de cliques do mouse Weibo...")
		fmt.Println("Pressione qualquer botão (Esquerdo, Direito, Meio, Laterais) no mouse...")
		fmt.Println("(Pressione Ctrl+C para sair)")

		listener, err := system.NewButtonListener()
		if err != nil {
			fmt.Printf("Erro ao iniciar ouvinte: %v\n", err)
			os.Exit(1)
		}

		// Registra todos os botões comuns
		for i := 1; i <= 9; i++ {
			btnNum := i
			listener.RegisterHandler(btnNum, func(b int, state string) {
				name := "Desconhecido"
				switch b {
				case 1:
					name = "Botão Esquerdo"
				case 2:
					name = "Botão Meio (Scroll Click)"
				case 3:
					name = "Botão Direito"
				case 4:
					name = "Scroll Up"
				case 5:
					name = "Scroll Down"
				case 8:
					name = "Botão Lateral (Voltar / Back)"
				case 9:
					name = "Botão Lateral (Avançar / Forward)"
				}
				fmt.Printf("[%s] Botão %d (%s) -> %s\n", time.Now().Format("15:04:05.000"), b, name, state)
			})
		}

		if err := listener.Start(); err != nil {
			fmt.Printf("Falha ao iniciar escuta: %v\n", err)
			os.Exit(1)
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		_ = listener.Stop()
		fmt.Println("\nMonitoramento encerrado.")

	case "sniff":
		system.SniffAllDevices()

	case "daemon":
		fmt.Println("Iniciando Weibo Control Daemon (Gerenciador de Atalhos e Botões)...")
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Erro ao carregar configurações: %v\n", err)
			os.Exit(1)
		}

		activeProfile := cfg.Profiles[cfg.ActiveProfileID]
		fmt.Printf("Perfil ativo: %s | DPI: %d\n", activeProfile.Name, activeProfile.ActiveDPI)
		_ = system.ApplySensitivity(activeProfile.ActiveDPI)

		uinputDev, err := system.NewUInputDevice()
		if err != nil {
			fmt.Printf("Erro no uinput: %v\n", err)
			os.Exit(1)
		}
		defer uinputDev.Close()

		listener, err := system.NewButtonListener()
		if err != nil {
			fmt.Printf("Erro no listener: %v\n", err)
			os.Exit(1)
		}

		// Handler para botão lateral 8 (Back)
		listener.RegisterHandler(8, func(btn int, state string) {
			if state == "press" {
				action := activeProfile.Buttons["button_5"].Value
				if action == "backward" || action == "" {
					return
				}
				fmt.Printf("[DAEMON] Botão 8 pressionado -> Executando '%s'\n", action)
				_ = uinputDev.ExecuteAction(action)
			}
		})

		// Handler para botão lateral 9 (Forward)
		listener.RegisterHandler(9, func(btn int, state string) {
			if state == "press" {
				action := activeProfile.Buttons["button_4"].Value
				if action == "forward" || action == "" {
					return
				}
				fmt.Printf("[DAEMON] Botão 9 pressionado -> Executando '%s'\n", action)
				_ = uinputDev.ExecuteAction(action)
			}
		})

		if err := listener.Start(); err != nil {
			fmt.Printf("Erro ao iniciar listener: %v\n", err)
			os.Exit(1)
		}

		// Monitor do botão superior de DPI
		dpiMon := system.NewDPIMonitor(func() {
			action := activeProfile.Buttons["button_6"].Value
			if action == "" || action == "cycle" || action == "dpi_cycle" {
				nextDPI := activeProfile.NextDPI()
				cfg.Profiles[cfg.ActiveProfileID] = activeProfile
				_ = config.Save(cfg)
				_ = system.ApplySensitivity(nextDPI)
				fmt.Printf("[DAEMON] 🔘 Botão Superior DPI Clicado -> Sincronizado para %d DPI!\n", nextDPI)
			} else {
				fmt.Printf("[DAEMON] 🔘 Botão Superior DPI Clicado -> Executando macro '%s'\n", action)
				_ = uinputDev.ExecuteAction(action)
			}
		})
		_ = dpiMon.Start()

		fmt.Println("Daemon ativo e escutando eventos de DPI e botões. Pressione Ctrl+C para finalizar.")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		_ = listener.Stop()
		dpiMon.Stop()
		fmt.Println("\nDaemon finalizado.")

	default:
		fmt.Printf("Comando desconhecido: %s\n", cmd)
		printHelp()
	}
}
