package ui

import (
	"fmt"
	"sync"

	"weibo-driver/pkg/config"
	"weibo-driver/pkg/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// DPIView mantém as referências diretas dos widgets para atualização em tempo real
type DPIView struct {
	cfg              *config.Config
	statusLabel      *widget.Label
	slider           *widget.Slider
	sliderValueLabel *widget.Label
	presetButtons    map[int]*widget.Button
	indicatorLabel   *widget.Label
	mu               sync.Mutex
	Container        fyne.CanvasObject
}

// NewDPIView inicializa a aba de DPI com suporte a atualização reativa
func NewDPIView(cfg *config.Config) *DPIView {
	v := &DPIView{
		cfg:           cfg,
		presetButtons: make(map[int]*widget.Button),
	}

	activeProfile := cfg.Profiles[cfg.ActiveProfileID]
	currentDPI := activeProfile.ActiveDPI
	if currentDPI == 0 {
		currentDPI = 1200
	}

	scale := system.DPIToScale(currentDPI)

	title := widget.NewLabelWithStyle("Controle de Sensibilidade e DPI", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	desc := widget.NewLabel("Alterne pelo botão físico do mouse ou pelo aplicativo — ambos sincronizam instantaneamente!")

	v.statusLabel = widget.NewLabel(fmt.Sprintf("DPI Atual: %d (Escala %.2fx)", currentDPI, scale))
	v.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	v.indicatorLabel = widget.NewLabel("Pronto para uso")
	v.indicatorLabel.TextStyle = fyne.TextStyle{Italic: true}

	v.slider = widget.NewSlider(400, 4800)
	v.slider.Step = 100
	v.slider.SetValue(float64(currentDPI))

	v.sliderValueLabel = widget.NewLabel(fmt.Sprintf("%d DPI", currentDPI))

	v.slider.OnChanged = func(val float64) {
		dpi := int(val)
		v.sliderValueLabel.SetText(fmt.Sprintf("%d DPI", dpi))
	}

	applyBtn := widget.NewButton("Aplicar Sensibilidade Manual", func() {
		v.UpdateDPI(int(v.slider.Value))
	})
	applyBtn.Importance = widget.HighImportance

	// Botões de estágios pré-definidos
	presetsTitle := widget.NewLabelWithStyle("Níveis de DPI Rápidos:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	dpiList := []int{800, 1200, 1600, 2400, 3200}
	var btnObjects []fyne.CanvasObject

	for _, d := range dpiList {
		dpiVal := d
		btn := widget.NewButton(fmt.Sprintf("%d DPI", dpiVal), func() {
			v.UpdateDPI(dpiVal)
		})
		if dpiVal == currentDPI {
			btn.Importance = widget.HighImportance
		}
		v.presetButtons[dpiVal] = btn
		btnObjects = append(btnObjects, btn)
	}

	presetGrid := container.NewGridWithColumns(len(dpiList), btnObjects...)
	sliderRow := container.NewBorder(nil, nil, nil, v.sliderValueLabel, v.slider)

	v.Container = container.NewVBox(
		title,
		desc,
		widget.NewSeparator(),
		v.statusLabel,
		v.indicatorLabel,
		widget.NewSeparator(),
		widget.NewLabel("Ajuste Fino Contínuo:"),
		sliderRow,
		applyBtn,
		widget.NewSeparator(),
		presetsTitle,
		presetGrid,
	)

	return v
}

// UpdateDPI atualiza visualmente os widgets e aplica a sensibilidade no sistema
func (v *DPIView) UpdateDPI(dpi int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	scale := system.DPIToScale(dpi)

	// Atualiza os widgets na thread principal do Fyne
	fyne.Do(func() {
		v.slider.SetValue(float64(dpi))
		v.sliderValueLabel.SetText(fmt.Sprintf("%d DPI", dpi))
		v.statusLabel.SetText(fmt.Sprintf("DPI Atual: %d (Escala %.2fx)", dpi, scale))
		v.indicatorLabel.SetText(fmt.Sprintf("⚡ Sensibilidade ajustada para %d DPI!", dpi))

		// Atualiza o destaque dos botões pré-definidos
		for val, btn := range v.presetButtons {
			if val == dpi {
				btn.Importance = widget.HighImportance
			} else {
				btn.Importance = widget.MediumImportance
			}
			btn.Refresh()
		}
	})

	// Aplica no Linux
	_ = system.ApplySensitivity(dpi)

	// Persiste a escolha
	prof := v.cfg.Profiles[v.cfg.ActiveProfileID]
	prof.ActiveDPI = dpi
	v.cfg.Profiles[v.cfg.ActiveProfileID] = prof
	_ = config.Save(v.cfg)
}
