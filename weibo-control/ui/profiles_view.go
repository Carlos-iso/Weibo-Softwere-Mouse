package ui

import (
	"fmt"

	"weibo-driver/pkg/config"
	"weibo-driver/pkg/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MakeProfilesView cria a aba de gerenciamento de perfis de usuário
func MakeProfilesView(cfg *config.Config, onConfigChanged func()) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Gerenciador de Perfis", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	desc := widget.NewLabel("Alterne entre perfis de sensibilidade e funções de botões salvas localmente em JSON.")

	path, _ := config.GetConfigPath()
	pathLabel := widget.NewLabel(fmt.Sprintf("Arquivo: %s", path))
	pathLabel.TextStyle = fyne.TextStyle{Italic: true}

	statusLabel := widget.NewLabel(fmt.Sprintf("Perfil Ativo Atual: %s", cfg.ActiveProfileID))
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	applyProfile := func(profileID string, name string, dpi int, b8Act, b9Act string) {
		prof := config.Profile{
			ID:          profileID,
			Name:        name,
			ActiveDPI:   dpi,
			DPIStages:   []int{800, 1200, 1600, 2400, 3200},
			PollingRate: 500,
			Buttons: map[string]config.ButtonAction{
				"button_5": {Type: "action", Value: b8Act},
				"button_4": {Type: "action", Value: b9Act},
			},
		}

		cfg.ActiveProfileID = profileID
		cfg.Profiles[profileID] = prof
		_ = config.Save(cfg)
		_ = system.ApplySensitivity(dpi)

		statusLabel.SetText(fmt.Sprintf("Perfil Ativo: %s (%d DPI)", name, dpi))
		onConfigChanged()
	}

	btnDefault := widget.NewButton("Ativar Perfil: Escritório / Geral (1200 DPI)", func() {
		applyProfile("office", "Escritório", 1200, "backward", "forward")
	})

	btnNav := widget.NewButton("Ativar Perfil: Produtividade (1600 DPI, Ctrl+C / Ctrl+V)", func() {
		applyProfile("productivity", "Produtividade", 1600, "ctrl+c", "ctrl+v")
	})

	btnGaming := widget.NewButton("Ativar Perfil: Gaming (2400 DPI, Alta Velocidade)", func() {
		applyProfile("gaming", "Gaming", 2400, "f5", "print")
	})

	return container.NewVBox(
		title,
		desc,
		pathLabel,
		widget.NewSeparator(),
		statusLabel,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Predefinições Rápidas de Perfil:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		btnDefault,
		btnNav,
		btnGaming,
	)
}
