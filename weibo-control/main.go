package main

import (
	"fmt"

	"weibo-control/ui"
	"weibo-driver/pkg/config"
	"weibo-driver/pkg/device"
	"weibo-driver/pkg/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.NewWithID("com.weibo.control")
	myApp.Settings().SetTheme(&ui.CustomTheme{})

	window := myApp.NewWindow("Weibo Control")
	window.Resize(fyne.NewSize(700, 540))
	window.CenterOnScreen()

	// Carrega configuração salva
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Cria o dispositivo uinput para macros virtuais
	uinputDev, _ := system.NewUInputDevice()
	if uinputDev != nil {
		defer uinputDev.Close()
	}

	// Cria a view de DPI reativa
	dpiView := ui.NewDPIView(cfg)

	// Inicia interceptação dos botões laterais automaticamente em segundo plano
	_ = ui.StartButtonDaemon(cfg, dpiView)
	defer ui.StopButtonDaemon()

	// Aplica a sensibilidade salva do perfil ativo
	activeProf := cfg.Profiles[cfg.ActiveProfileID]
	if activeProf.ActiveDPI > 0 {
		_ = system.ApplySensitivity(activeProf.ActiveDPI)
	}

	// Barra superior de Diagnóstico de Hardware
	headerTitle := widget.NewLabelWithStyle("WEIBO CONTROL", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	headerSubtitle := widget.NewLabel("Central de Configuração do Mouse Weibo")
	headerSubtitle.TextStyle = fyne.TextStyle{Italic: true}

	var statusText string
	if path, err := device.FindHidrawPath(); err == nil {
		statusText = fmt.Sprintf("Hardware: Receptor YICHIP (3151:3020) Conectado em %s", path)
	} else {
		statusText = "Hardware: Receptor USB não detectado (operando em modo software)"
	}
	hwStatusLabel := widget.NewLabel(statusText)

	resetBtn := widget.NewButton("Resetar Dongle USB", func() {
		dev := device.NewDevice(false)
		if err := dev.ResetUSB(); err != nil {
			hwStatusLabel.SetText(fmt.Sprintf("Erro no reset: %v", err))
		} else {
			hwStatusLabel.SetText("Dongle USB resetado e reconectado com sucesso!")
		}
	})

	headerLeft := container.NewVBox(headerTitle, headerSubtitle, hwStatusLabel)
	header := container.NewBorder(nil, nil, nil, resetBtn, headerLeft)

	// Abas da aplicação
	tabDPI := container.NewTabItem("Sensibilidade & DPI", dpiView.Container)
	tabButtons := container.NewTabItem("Botões & Laterais", ui.MakeButtonsView(cfg, dpiView))
	tabMacros := container.NewTabItem("Macros & Atalhos", ui.MakeMacroView())
	tabProfiles := container.NewTabItem("Perfis", ui.MakeProfilesView(cfg, func() {
		active := cfg.Profiles[cfg.ActiveProfileID]
		dpiView.UpdateDPI(active.ActiveDPI)
	}))

	tabs := container.NewAppTabs(tabDPI, tabButtons, tabMacros, tabProfiles)
	tabs.SetTabLocation(container.TabLocationTop)

	mainContent := container.NewBorder(
		container.NewVBox(header, widget.NewSeparator()),
		nil,
		nil,
		nil,
		tabs,
	)

	window.SetContent(mainContent)
	window.ShowAndRun()
}
