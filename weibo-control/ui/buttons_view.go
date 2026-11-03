package ui

import (
	"fmt"

	"weibo-driver/pkg/config"
	"weibo-driver/pkg/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	activeListener  *system.ButtonListener
	activeUInput    *system.UInputDevice
	listenerRunning bool
)


// StartButtonDaemon inicia a escuta e execução de macros/DPI em segundo plano
func StartButtonDaemon(cfg *config.Config, dpiView *DPIView) error {
	if listenerRunning {
		return nil
	}

	uinputDev, err := system.NewUInputDevice()
	if err != nil {
		return fmt.Errorf("erro ao iniciar uinput: %w", err)
	}
	activeUInput = uinputDev

	listener, err := system.NewButtonListener()
	if err != nil {
		_ = activeUInput.Close()
		activeUInput = nil
		return fmt.Errorf("erro ao iniciar listener: %w", err)
	}

	// Monitor de Botão Lateral 8 (Back)
	listener.RegisterHandler(8, func(btn int, state string) {
		if state == "press" {
			act := cfg.Profiles[cfg.ActiveProfileID].Buttons["button_5"].Value
			if act == "cycle" {
				prof := cfg.Profiles[cfg.ActiveProfileID]
				nextDPI := prof.NextDPI()
				dpiView.UpdateDPI(nextDPI)
			} else if act != "default" && act != "backward" && act != "" {
				_ = activeUInput.ExecuteAction(act)
			}
		}
	})

	// Monitor de Botão Lateral 9 (Forward)
	listener.RegisterHandler(9, func(btn int, state string) {
		if state == "press" {
			act := cfg.Profiles[cfg.ActiveProfileID].Buttons["button_4"].Value
			if act == "cycle" {
				prof := cfg.Profiles[cfg.ActiveProfileID]
				nextDPI := prof.NextDPI()
				dpiView.UpdateDPI(nextDPI)
			} else if act != "default" && act != "forward" && act != "" {
				_ = activeUInput.ExecuteAction(act)
			}
		}
	})

	if err := listener.Start(); err != nil {
		_ = activeUInput.Close()
		activeUInput = nil
		return fmt.Errorf("falha ao iniciar escuta: %w", err)
	}

	activeListener = listener
	listenerRunning = true
	return nil
}

// StopButtonDaemon encerra a escuta e o dispositivo uinput
func StopButtonDaemon() {
	if activeListener != nil {
		_ = activeListener.Stop()
		activeListener = nil
	}
	if activeUInput != nil {
		_ = activeUInput.Close()
		activeUInput = nil
	}
	listenerRunning = false
}

// IsButtonDaemonRunning retorna o status da escuta
func IsButtonDaemonRunning() bool {
	return listenerRunning
}

// MakeButtonsView cria a aba de mapeamento dos botões físicos e laterais
func MakeButtonsView(cfg *config.Config, dpiView *DPIView) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Mapeamento dos Botões do Mouse", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	desc := widget.NewLabel("Configure ações, atalhos de teclado e ciclar DPI para os botões do Weibo.")

	activeProfile := cfg.Profiles[cfg.ActiveProfileID]

	// Botões 1, 2, 3 (Fixos padrão)
	b1Label := widget.NewLabel("Botão 1 (Esquerdo):")
	b1Action := widget.NewLabel("Clique Primário (Padrão do Sistema)")
	b1Action.TextStyle = fyne.TextStyle{Italic: true}

	b2Label := widget.NewLabel("Botão 2 (Direito):")
	b2Action := widget.NewLabel("Clique Secundário / Menu Contexto")
	b2Action.TextStyle = fyne.TextStyle{Italic: true}

	b3Label := widget.NewLabel("Botão 3 (Scroll):")
	b3Action := widget.NewLabel("Clique do Meio / Rolagem")
	b3Action.TextStyle = fyne.TextStyle{Italic: true}

	// Informação sobre o botão físico superior de DPI
	bDpiInfoLabel := widget.NewLabel("Botão Superior (DPI Físico):")
	bDpiInfoDesc := widget.NewLabel("Chaveador Analógico do Sensor (Hardware Interno)")
	bDpiInfoDesc.TextStyle = fyne.TextStyle{Italic: true}

	// Opções para botões laterais
	actionOptions := []string{
		"Padrão do Navegador",
		"Ciclar DPI (Sincronizar Software)",
		"Copiar (Ctrl+C)",
		"Colar (Ctrl+V)",
		"Desfazer (Ctrl+Z)",
		"Mudo (Mute)",
		"Aumentar Volume",
		"Diminuir Volume",
		"Play / Pause",
		"Atualizar Página (F5)",
		"PrintScreen",
	}

	mapDisplayToAction := func(display string) string {
		switch display {
		case "Ciclar DPI (Sincronizar Software)":
			return "cycle"
		case "Copiar (Ctrl+C)":
			return "ctrl+c"
		case "Colar (Ctrl+V)":
			return "ctrl+v"
		case "Desfazer (Ctrl+Z)":
			return "ctrl+z"
		case "Mudo (Mute)":
			return "mute"
		case "Aumentar Volume":
			return "volume_up"
		case "Diminuir Volume":
			return "volume_down"
		case "Play / Pause":
			return "play_pause"
		case "Atualizar Página (F5)":
			return "f5"
		case "PrintScreen":
			return "print"
		default:
			return "default"
		}
	}

	mapActionToDisplay := func(action string) string {
		switch action {
		case "cycle", "dpi_cycle":
			return "Ciclar DPI (Sincronizar Software)"
		case "ctrl+c":
			return "Copiar (Ctrl+C)"
		case "ctrl+v":
			return "Colar (Ctrl+V)"
		case "ctrl+z":
			return "Desfazer (Ctrl+Z)"
		case "mute":
			return "Mudo (Mute)"
		case "volume_up":
			return "Aumentar Volume"
		case "volume_down":
			return "Diminuir Volume"
		case "play_pause":
			return "Play / Pause"
		case "f5":
			return "Atualizar Página (F5)"
		case "print":
			return "PrintScreen"
		default:
			return "Padrão do Navegador"
		}
	}

	// Botão Lateral 8 (Back / Traseiro)
	b8Label := widget.NewLabel("Botão 8 (Lateral Traseiro / Voltar):")
	currentB8Action := activeProfile.Buttons["button_5"].Value
	b8Select := widget.NewSelect(actionOptions, func(s string) {
		if s == "" {
			return
		}
		act := mapDisplayToAction(s)
		prof := cfg.Profiles[cfg.ActiveProfileID]
		prof.Buttons["button_5"] = config.ButtonAction{Type: "action", Value: act}
		cfg.Profiles[cfg.ActiveProfileID] = prof
		_ = config.Save(cfg)
	})
	b8Select.PlaceHolder = "Selecione a ação"
	b8Select.Selected = mapActionToDisplay(currentB8Action)

	// Botão Lateral 9 (Forward / Dianteiro)
	b9Label := widget.NewLabel("Botão 9 (Lateral Dianteiro / Avançar):")
	currentB9Action := activeProfile.Buttons["button_4"].Value
	b9Select := widget.NewSelect(actionOptions, func(s string) {
		if s == "" {
			return
		}
		act := mapDisplayToAction(s)
		prof := cfg.Profiles[cfg.ActiveProfileID]
		prof.Buttons["button_4"] = config.ButtonAction{Type: "action", Value: act}
		cfg.Profiles[cfg.ActiveProfileID] = prof
		_ = config.Save(cfg)
	})
	b9Select.PlaceHolder = "Selecione a ação"
	b9Select.Selected = mapActionToDisplay(currentB9Action)

	// Seção do Daemon / Interceptador
	daemonTitle := widget.NewLabelWithStyle("Serviço de Interceptação dos Botões Laterais:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	daemonStatus := widget.NewLabel("Status: Desativado")
	toggleDaemonBtnText := "Ativar Interceptação em Segundo Plano"
	if listenerRunning {
		daemonStatus.SetText("Status: Ativo (respondendo aos botões laterais automaticamente)")
		toggleDaemonBtnText = "Desativar Interceptação"
	}

	var toggleDaemonBtn *widget.Button
	toggleDaemonBtn = widget.NewButton(toggleDaemonBtnText, func() {
		if !listenerRunning {
			if err := StartButtonDaemon(cfg, dpiView); err != nil {
				daemonStatus.SetText(fmt.Sprintf("Erro ao iniciar: %v", err))
				return
			}
			daemonStatus.SetText("Status: Ativo (respondendo aos botões laterais)")
			toggleDaemonBtn.SetText("Desativar Interceptação")
		} else {
			StopButtonDaemon()
			daemonStatus.SetText("Status: Desativado")
			toggleDaemonBtn.SetText("Ativar Interceptação em Segundo Plano")
		}
	})

	grid := container.NewGridWithColumns(2,
		b1Label, b1Action,
		b2Label, b2Action,
		b3Label, b3Action,
		bDpiInfoLabel, bDpiInfoDesc,
		b8Label, b8Select,
		b9Label, b9Select,
	)

	return container.NewVBox(
		title,
		desc,
		widget.NewSeparator(),
		grid,
		widget.NewSeparator(),
		daemonTitle,
		daemonStatus,
		toggleDaemonBtn,
	)
}
