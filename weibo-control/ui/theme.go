package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomTheme implementa o tema moderno escuro para o Weibo Control
type CustomTheme struct{}

var _ fyne.Theme = (*CustomTheme)(nil)

func (m *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 20, G: 24, B: 33, A: 255} // Dark slate
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 30, G: 36, B: 49, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0, G: 168, B: 255, A: 255} // Vibrant cyan/blue
	case theme.ColorNameForeground:
		return color.NRGBA{R: 235, G: 240, B: 250, A: 255}
	case theme.ColorNameButton:
		return color.NRGBA{R: 38, G: 46, B: 62, A: 255}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 70, G: 80, B: 100, A: 255}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 60, G: 72, B: 95, A: 200}
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func (m *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameText:
		return 13
	case theme.SizeNameHeadingText:
		return 18
	default:
		return theme.DefaultTheme().Size(name)
	}
}
