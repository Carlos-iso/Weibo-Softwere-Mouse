package ui

import (
	"fmt"

	"weibo-driver/pkg/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MakeMacroView cria a aba de atalhos e testes de macro via uinput
func MakeMacroView() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Editor e Testador de Macros Virtuais", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	desc := widget.NewLabel("Emita comandos e sequências de teclas no Linux através do dispositivo virtual /dev/uinput.")

	resultLabel := widget.NewLabel("Status: Pronto")

	testAction := func(action string) {
		uinputDev, err := system.NewUInputDevice()
		if err != nil {
			resultLabel.SetText(fmt.Sprintf("Erro ao abrir uinput: %v", err))
			return
		}
		defer uinputDev.Close()

		if err := uinputDev.ExecuteAction(action); err != nil {
			resultLabel.SetText(fmt.Sprintf("Erro: %v", err))
			return
		}

		resultLabel.SetText(fmt.Sprintf("Sucesso: Comando '%s' emitido no sistema!", action))
	}

	btnCopy := widget.NewButton("Testar Copiar (Ctrl+C)", func() { testAction("ctrl+c") })
	btnPaste := widget.NewButton("Testar Colar (Ctrl+V)", func() { testAction("ctrl+v") })
	btnUndo := widget.NewButton("Testar Desfazer (Ctrl+Z)", func() { testAction("ctrl+z") })
	btnMute := widget.NewButton("Testar Mudo", func() { testAction("mute") })
	btnVolUp := widget.NewButton("Testar Volume +", func() { testAction("volume_up") })
	btnVolDown := widget.NewButton("Testar Volume -", func() { testAction("volume_down") })

	quickGrid := container.NewGridWithColumns(3, btnCopy, btnPaste, btnUndo, btnMute, btnVolUp, btnVolDown)

	customEntry := widget.NewEntry()
	customEntry.SetPlaceHolder("ex: ctrl+c, alt+tab, volume_up, f5")

	customBtn := widget.NewButton("Disparar Atalho Personalizado", func() {
		text := customEntry.Text
		if text != "" {
			testAction(text)
		}
	})

	return container.NewVBox(
		title,
		desc,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Atalhos Rápidos Prontos:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		quickGrid,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Disparador Personalizado:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		customEntry,
		customBtn,
		widget.NewSeparator(),
		resultLabel,
	)
}
